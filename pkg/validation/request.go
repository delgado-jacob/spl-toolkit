package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// InputError identifies invalid catalog/request data and unsupported document options.
type InputError struct{ Err error }

func (e *InputError) Error() string { return e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }
func IsInputError(err error) bool   { var input *InputError; return errors.As(err, &input) }
func inputError(format string, args ...any) error {
	return &InputError{Err: fmt.Errorf(format, args...)}
}

// checkJSON preserves source Unicode before encoding/json can replace it, and
// enforces one complete value. Object policy is handled by object below.
func checkJSON(data []byte) error {
	if err := jsoninput.ValidateUnicode(data); err != nil {
		return inputError("%v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return inputError("invalid JSON: %v", err)
	}
	if err := decoder.Decode(&raw); err != io.EOF {
		return inputError("expected exactly one JSON value")
	}
	return nil
}

// object is the single strict object-key/presence gate for validation inputs.
func object(data []byte, required []string, allowed ...string) (map[string]json.RawMessage, error) {
	if err := checkJSON(data); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, inputError("expected an object")
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, inputError("invalid object key: %v", err)
		}
		key, ok := token.(string)
		if !ok {
			return nil, inputError("expected string object key")
		}
		if _, exists := fields[key]; exists {
			return nil, inputError("duplicate property %q", key)
		}
		known := false
		for _, name := range allowed {
			if name == key {
				known = true
				break
			}
		}
		if !known {
			return nil, inputError("unknown property %q", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, inputError("invalid property %q: %v", key, err)
		}
		fields[key] = value
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return nil, inputError("missing property %q", key)
		}
	}
	return fields, nil
}

func stringValue(data []byte) (string, error) {
	if len(data) == 0 || data[0] != '"' {
		return "", inputError("expected a string")
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", inputError("invalid string: %v", err)
	}
	return value, nil
}
func array(data []byte) ([]json.RawMessage, error) {
	if err := checkJSON(data); err != nil {
		return nil, err
	}
	data = bytes.TrimSpace(data)
	if data[0] != '[' {
		return nil, inputError("expected an array")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, inputError("invalid array: %v", err)
	}
	return values, nil
}

func normalizeDocument(doc analysis.QueryDocument) (analysis.QueryDocument, error) {
	manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: doc.Language, Profile: doc.Profile, Version: doc.Version})
	if err != nil {
		return doc, inputError("%v", err)
	}
	doc.Language, doc.Profile, doc.Version = manifest.Language, manifest.Profile, manifest.Version
	if !utf8.ValidString(doc.Text) || !utf8.ValidString(doc.SourceID) {
		return doc, inputError("document text and source_id must be valid UTF-8")
	}
	return doc, nil
}
func decodeDocument(data []byte) (analysis.QueryDocument, error) {
	var doc analysis.QueryDocument
	fields, err := object(data, []string{"text"}, "text", "language", "profile", "version", "source_id")
	if err != nil {
		return doc, err
	}
	for _, entry := range []struct {
		key  string
		dest *string
	}{{"text", &doc.Text}, {"language", &doc.Language}, {"profile", &doc.Profile}, {"version", &doc.Version}, {"source_id", &doc.SourceID}} {
		if raw, ok := fields[entry.key]; ok {
			value, err := stringValue(raw)
			if err != nil {
				return doc, inputError("document %s: %v", entry.key, err)
			}
			*entry.dest = value
		}
	}
	return normalizeDocument(doc)
}

func DecodeDocuments(data []byte) ([]analysis.QueryDocument, error) {
	values, err := array(data)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, inputError("documents must be nonempty")
	}
	documents := make([]analysis.QueryDocument, 0, len(values))
	for i, value := range values {
		doc, err := decodeDocument(value)
		if err != nil {
			return nil, inputError("document %d: %v", i, err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}
func DecodeRequest(data []byte) (Request, error) {
	var request Request
	fields, err := object(data, []string{"document", "catalog"}, "document", "catalog")
	if err != nil {
		return request, err
	}
	doc, err := decodeDocument(fields["document"])
	if err != nil {
		return request, err
	}
	catalog, err := DecodeFieldCatalog(fields["catalog"])
	if err != nil {
		return request, err
	}
	return Request{Document: doc, Catalog: catalog}, nil
}
func DecodeBatchRequest(data []byte) (BatchRequest, error) {
	var request BatchRequest
	fields, err := object(data, []string{"documents", "catalog"}, "documents", "catalog")
	if err != nil {
		return request, err
	}
	docs, err := DecodeDocuments(fields["documents"])
	if err != nil {
		return request, err
	}
	catalog, err := DecodeFieldCatalog(fields["catalog"])
	if err != nil {
		return request, err
	}
	return BatchRequest{Documents: docs, Catalog: catalog}, nil
}

// decodeUniqueJSON keeps numbers exact and rejects duplicate keys throughout,
// including annotation data. Schema-bearing positions are identified separately.
func decodeUniqueJSON(data []byte) (any, error) {
	if err := checkJSON(data); err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var read func() (any, error)
	read = func() (any, error) {
		token, err := d.Token()
		if err != nil {
			return nil, inputError("invalid JSON: %v", err)
		}
		switch token {
		case json.Delim('{'):
			m := map[string]any{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return nil, inputError("invalid JSON key")
				}
				k := key.(string)
				if _, ok := m[k]; ok {
					return nil, inputError("duplicate property %q", k)
				}
				v, err := read()
				if err != nil {
					return nil, err
				}
				m[k] = v
			}
			_, err = d.Token()
			return m, err
		case json.Delim('['):
			a := []any{}
			for d.More() {
				v, err := read()
				if err != nil {
					return nil, err
				}
				a = append(a, v)
			}
			_, err = d.Token()
			return a, err
		default:
			return token, nil
		}
	}
	return read()
}
