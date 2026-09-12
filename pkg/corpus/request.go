package corpus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

type InputError struct{ Err error }

func (e *InputError) Error() string { return e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }
func IsInputError(err error) bool   { var e *InputError; return errors.As(err, &e) }
func inputError(format string, args ...any) error {
	return &InputError{Err: fmt.Errorf(format, args...)}
}

func strictObject(data []byte, required []string, allowed ...string) (map[string]json.RawMessage, error) {
	if err := jsoninput.ValidateUnicode(data); err != nil {
		return nil, inputError("%v", err)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return nil, inputError("expected an object")
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		keyToken, err := d.Token()
		if err != nil {
			return nil, inputError("invalid object key: %v", err)
		}
		key := keyToken.(string)
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
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return nil, inputError("property %q: %v", key, err)
		}
		fields[key] = raw
	}
	if _, err := d.Token(); err != nil {
		return nil, inputError("invalid object: %v", err)
	}
	var trailing any
	if err := d.Decode(&trailing); err != io.EOF {
		return nil, inputError("expected exactly one JSON object")
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return nil, inputError("missing property %q", key)
		}
	}
	return fields, nil
}

func strictString(raw []byte) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '"' {
		return "", inputError("expected a string")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", inputError("invalid string: %v", err)
	}
	return value, nil
}

func DecodeValidationTarget(raw []byte) (*ValidationTarget, error) {
	f, err := strictObject(raw, []string{"kind"}, "kind", "catalog", "identity", "schema", "base_uri", "resources", "selection")
	if err != nil {
		return nil, err
	}
	kind, err := strictString(f["kind"])
	if err != nil {
		return nil, err
	}
	if kind == "field_list" {
		if _, err := strictObject(raw, []string{"kind", "catalog"}, "kind", "catalog"); err != nil {
			return nil, err
		}
		catalog, err := validation.DecodeFieldCatalog(f["catalog"])
		if err != nil {
			return nil, inputError("validation_target: %v", err)
		}
		return &ValidationTarget{Kind: kind, Catalog: &catalog}, nil
	}
	target, err := validation.DecodeSchemaTarget(raw)
	if err != nil {
		return nil, inputError("validation_target: %v", err)
	}
	return &ValidationTarget{Kind: kind, SchemaTarget: &target}, nil
}

// DecodeRequest accepts only inline canonical documents, never local file paths.
func DecodeRequest(data []byte) (Request, error) {
	f, err := strictObject(data, []string{"schema_version", "documents"}, "schema_version", "documents", "validation_target")
	if err != nil {
		return Request{}, err
	}
	if string(bytes.TrimSpace(f["schema_version"])) != "1" {
		return Request{}, inputError("schema_version must be integer 1")
	}
	if len(bytes.TrimSpace(f["documents"])) == 0 || bytes.TrimSpace(f["documents"])[0] != '[' {
		return Request{}, inputError("documents must be an array")
	}
	var items []json.RawMessage
	if err := json.Unmarshal(f["documents"], &items); err != nil || len(items) == 0 {
		return Request{}, inputError("documents must be a nonempty array")
	}
	r := Request{SchemaVersion: 1, Documents: make([]RequestDocument, 0, len(items))}
	seen := map[string]bool{}
	for i, raw := range items {
		entry, err := strictObject(raw, []string{"id", "document"}, "id", "document")
		if err != nil {
			return Request{}, inputError("document %d: %v", i, err)
		}
		id, err := strictString(entry["id"])
		if err != nil || strings.TrimSpace(id) == "" || seen[id] {
			return Request{}, inputError("document %d has an invalid or duplicate id", i)
		}
		seen[id] = true
		documentRaw := append(append([]byte{'['}, entry["document"]...), ']')
		documents, err := validation.DecodeDocuments(documentRaw)
		if err != nil {
			return Request{}, inputError("document %d: %v", i, err)
		}
		r.Documents = append(r.Documents, RequestDocument{ID: id, Document: documents[0]})
	}
	if raw, ok := f["validation_target"]; ok {
		r.ValidationTarget, err = DecodeValidationTarget(raw)
		if err != nil {
			return Request{}, err
		}
	}
	return r, nil
}
