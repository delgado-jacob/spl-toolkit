// Package lsp implements the local LSP 3.17 full-sync adapter.
package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"strconv"
	"strings"
	"sync"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
)

const maxHeaderBytes = 8192
const maxBodyBytes = 67108864

// readFrame never scans or allocates beyond the declared transport bounds.
func readFrame(r *bufio.Reader) ([]byte, error) {
	header := make([]byte, 0, 256)
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && len(header) > 0 {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
		if b > 127 || (b < 32 && b != '\r' && b != '\n' && b != '\t') {
			return nil, fmt.Errorf("LSP headers require ASCII text")
		}
		header = append(header, b)
		if bytes.HasSuffix(header, []byte("\r\n\r\n")) {
			break
		}
		if len(header) >= maxHeaderBytes {
			return nil, fmt.Errorf("LSP header exceeds %d bytes", maxHeaderBytes)
		}
	}
	length := -1
	contentTypeSeen := false
	for _, line := range strings.Split(string(header[:len(header)-4]), "\r\n") {
		name, value, ok := strings.Cut(line, ":")
		if !ok || strings.ContainsAny(line, "\r\n") {
			return nil, fmt.Errorf("malformed LSP header")
		}
		value = strings.TrimSpace(value)
		switch strings.ToLower(name) {
		case "content-length":
			if length >= 0 || value == "" {
				return nil, fmt.Errorf("missing or duplicate Content-Length")
			}
			for _, c := range value {
				if c < '0' || c > '9' {
					return nil, fmt.Errorf("Content-Length must be decimal")
				}
			}
			n, err := strconv.ParseUint(value, 10, 64)
			if err != nil || n > maxBodyBytes {
				return nil, fmt.Errorf("Content-Length exceeds %d bytes", maxBodyBytes)
			}
			length = int(n)
		case "content-type":
			if contentTypeSeen {
				return nil, fmt.Errorf("duplicate Content-Type")
			}
			contentTypeSeen = true
			_, params, err := mime.ParseMediaType(value)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Type")
			}
			charset := strings.ToLower(params["charset"])
			if charset != "" && charset != "utf-8" && charset != "utf8" {
				return nil, fmt.Errorf("unsupported LSP charset %q", charset)
			}
		}
	}
	if length < 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}
	body := make([]byte, length)
	_, err := io.ReadFull(r, body)
	return body, err
}

type frameWriter struct {
	mu  sync.Mutex
	out io.Writer
	err error
}

func (w *frameWriter) write(value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return w.err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if err = writeAll(w.out, []byte(header)); err != nil {
		w.err = err
		return err
	}
	w.err = writeAll(w.out, body)
	return w.err
}
func writeAll(w io.Writer, b []byte) error {
	for len(b) > 0 {
		n, err := w.Write(b)
		if err != nil {
			return err
		}
		if n <= 0 || n > len(b) {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}

type message struct {
	ID     json.RawMessage
	Method string
	Params json.RawMessage
}
type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *responseError) Error() string { return e.Message }

// object retains raw values and rejects duplicate keys instead of silently
// accepting the last security-sensitive configuration or protocol property.
func object(raw []byte) (map[string]json.RawMessage, error) {
	if err := jsoninput.ValidateUnicode(raw); err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	tok, err := d.Token()
	if err != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("expected object")
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		tok, err = d.Token()
		if err != nil {
			return nil, err
		}
		key := tok.(string)
		if _, ok := fields[key]; ok {
			return nil, fmt.Errorf("duplicate property %q", key)
		}
		var v json.RawMessage
		if err = d.Decode(&v); err != nil {
			return nil, err
		}
		fields[key] = v
	}
	if _, err = d.Token(); err != nil {
		return nil, err
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("expected one object")
	}
	return fields, nil
}
func validID(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	if raw[0] == '"' {
		var s string
		return json.Unmarshal(raw, &s) == nil
	}
	s := string(raw)
	if strings.HasPrefix(s, "-") {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Request keys compare JSON values while responses retain the original token.
func idKey(raw json.RawMessage) string {
	if len(raw) > 0 && raw[0] == '"' {
		var value string
		_ = json.Unmarshal(raw, &value)
		return "s:" + value
	}
	if string(raw) == "-0" {
		return "n:0"
	}
	return "n:" + string(raw)
}
func decodeMessage(raw []byte) (message, *responseError) {
	m := message{}
	if !json.Valid(raw) {
		return m, &responseError{-32700, "Parse error"}
	}
	f, err := object(raw)
	if err != nil {
		return m, &responseError{-32600, "Invalid Request: " + err.Error()}
	}
	var version string
	if json.Unmarshal(f["jsonrpc"], &version) != nil || version != "2.0" || json.Unmarshal(f["method"], &m.Method) != nil || m.Method == "" {
		return message{}, &responseError{-32600, "Invalid Request"}
	}
	if id, ok := f["id"]; ok {
		if !validID(id) {
			return message{}, &responseError{-32600, "Invalid request ID"}
		}
		m.ID = id
	}
	m.Params = f["params"]
	return m, nil
}

func (w *frameWriter) reply(id json.RawMessage, result any, rpcErr *responseError) error {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	if rpcErr != nil {
		return w.write(struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Error   *responseError  `json:"error"`
		}{"2.0", id, rpcErr})
	}
	return w.write(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  any             `json:"result"`
	}{"2.0", id, result})
}
func (w *frameWriter) notify(method string, params any) error {
	return w.write(struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}{"2.0", method, params})
}
