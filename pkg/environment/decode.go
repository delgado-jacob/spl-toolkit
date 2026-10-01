package environment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
)

// decodeStrictJSON rejects duplicate members at any depth before Go struct decoding.
func decodeStrictJSON(raw []byte, target any) error {
	if err := jsoninput.ValidateUnicode(raw); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var read func() error
	read = func() error {
		token, err := d.Token()
		if err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		switch token {
		case json.Delim('{'):
			seen := map[string]bool{}
			for d.More() {
				keyToken, err := d.Token()
				if err != nil {
					return fmt.Errorf("invalid JSON key: %w", err)
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("expected a string object key")
				}
				if seen[key] {
					return fmt.Errorf("duplicate property %q", key)
				}
				seen[key] = true
				if err := read(); err != nil {
					return err
				}
			}
			if _, err := d.Token(); err != nil {
				return err
			}
		case json.Delim('['):
			for d.More() {
				if err := read(); err != nil {
					return err
				}
			}
			if _, err := d.Token(); err != nil {
				return err
			}
		}
		return nil
	}
	if err := read(); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON value")
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return fmt.Errorf("invalid artifact: %w", err)
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON value")
	}
	return nil
}

// validateSnapshotRawShape preserves required-member and tagged-selector distinctions
// that ordinary struct decoding would lose (notably absent versus null slices).
func validateSnapshotRawShape(raw []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return err
	}
	if root == nil {
		return fmt.Errorf("snapshot must be an object")
	}
	for _, key := range []string{"schema_version", "scope_id", "capture_scope", "origin", "capture", "capabilities", "collections", "objects"} {
		value, found := root[key]
		if !found {
			return fmt.Errorf("missing property %q", key)
		}
		if bytes.Equal(value, []byte("null")) {
			return fmt.Errorf("%s must not be null", key)
		}
	}
	for _, key := range []string{"capabilities", "collections", "objects"} {
		if len(root[key]) == 0 || root[key][0] != '[' {
			return fmt.Errorf("%s must be an array", key)
		}
	}
	var scope map[string]json.RawMessage
	if err := json.Unmarshal(root["capture_scope"], &scope); err != nil || scope == nil {
		return fmt.Errorf("capture_scope must be an object")
	}
	for _, name := range []string{"namespace", "app", "owner"} {
		rawSelector, found := scope[name]
		if !found {
			return fmt.Errorf("missing capture_scope.%s", name)
		}
		var branches map[string]json.RawMessage
		if err := json.Unmarshal(rawSelector, &branches); err != nil || branches == nil {
			return fmt.Errorf("%s selector must be an object", name)
		}
		all, hasAll := branches["all"]
		values, hasValues := branches["values"]
		if hasAll == hasValues {
			return fmt.Errorf("%s selector must use exactly one branch", name)
		}
		if hasAll && bytes.Equal(all, []byte("null")) || hasValues && bytes.Equal(values, []byte("null")) {
			return fmt.Errorf("%s selector must not be null", name)
		}
	}
	return nil
}
