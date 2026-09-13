package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestFraming(t *testing.T) {
	for _, body := range []string{`{"id":9007199254740993,"jsonrpc":"2.0","method":"x"}`, `{"x":"😀"}`} {
		wire := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
		got, err := readFrame(bufio.NewReader(strings.NewReader(wire)))
		if err != nil || string(got) != body {
			t.Fatalf("%s: %s %v", wire, got, err)
		}
	}
	for _, wire := range []string{"Content-Length: -1\r\n\r\n", "Content-Length: +1\r\n\r\nx", "Content-Length: 1\r\nContent-Length: 1\r\n\r\nx", "Content-Length: 67108865\r\n\r\n", "Content-Length: 999999999999999999999\r\n\r\n", "Content-Length: 2\r\n\r\nx", "Content-Length: 1\r\nContent-Type: application/vscode-jsonrpc; charset=latin1\r\n\r\nx", strings.Repeat("x", 8193), "Content-Length: 1\n\nx"} {
		if _, err := readFrame(bufio.NewReader(strings.NewReader(wire))); err == nil {
			t.Fatalf("accepted malformed frame %.100s", wire)
		}
	}
	prefix := "Content-Length: 0\r\nX: "
	exact := prefix + strings.Repeat("a", 8192-len(prefix)-4) + "\r\n\r\n"
	if _, err := readFrame(bufio.NewReader(strings.NewReader(exact))); err != nil {
		t.Fatal(err)
	}
	if _, err := readFrame(bufio.NewReader(strings.NewReader(strings.Replace(exact, "X: ", "X: a", 1)))); err == nil {
		t.Fatal("header overflow accepted")
	}
	body := bytes.Repeat([]byte(" "), 67108864)
	if got, err := readFrame(bufio.NewReader(io.MultiReader(strings.NewReader("Content-Length: 67108864\r\n\r\n"), bytes.NewReader(body)))); err != nil || len(got) != len(body) {
		t.Fatalf("exact body limit: %d %v", len(got), err)
	}
}

func TestSerializedWrites(t *testing.T) {
	var out bytes.Buffer
	w := frameWriter{out: &out}
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := w.write(map[string]any{"id": i, "text": "😀"}); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	r := bufio.NewReader(&out)
	for i := 0; i < 30; i++ {
		b, err := readFrame(r)
		if err != nil || !json.Valid(b) {
			t.Fatalf("invalid frame: %s %v", b, err)
		}
	}
}

func TestRequestIDsAndInvalidMessages(t *testing.T) {
	for _, id := range []string{`9007199254740993`, `-9007199254740993`, `"abc"`, `0`} {
		m, err := decodeMessage([]byte(`{"jsonrpc":"2.0","id":` + id + `,"method":"x"}`))
		if err != nil || string(m.ID) != id {
			t.Fatalf("id changed: %s %v", m.ID, err)
		}
	}
	for _, raw := range []string{`[]`, `null`, `{"jsonrpc":"2.0","id":1.5,"method":"x"}`, `{"jsonrpc":"2.0","id":true,"method":"x"}`, `{"jsonrpc":"2.0","id":1,"id":2,"method":"x"}`} {
		if _, err := decodeMessage([]byte(raw)); err == nil || err.Code != -32600 {
			t.Fatalf("accepted %s: %v", raw, err)
		}
	}
	if _, err := decodeMessage([]byte(`{`)); err == nil || err.Code != -32700 {
		t.Fatalf("parse error: %v", err)
	}
}

func TestHeaderRejectsNonASCII(t *testing.T) {
	if _, err := readFrame(bufio.NewReader(strings.NewReader("Content-Length: 0\r\nX-Name: 😀\r\n\r\n"))); err == nil {
		t.Fatal("non-ASCII header accepted")
	}
}

type failingOnceWriter struct{ calls int }

func (w *failingOnceWriter) Write(b []byte) (int, error) {
	w.calls++
	if w.calls == 1 {
		return 0, io.ErrClosedPipe
	}
	return len(b), nil
}
func TestWriterFailureCannotResumeCorruptStream(t *testing.T) {
	out := &failingOnceWriter{}
	writer := frameWriter{out: out}
	if writer.write(map[string]any{"id": 1}) == nil {
		t.Fatal("write failure hidden")
	}
	if writer.write(map[string]any{"id": 2}) == nil {
		t.Fatal("writer resumed after framing failure")
	}
	if out.calls != 1 {
		t.Fatalf("wrote after broken frame: %d", out.calls)
	}
}
