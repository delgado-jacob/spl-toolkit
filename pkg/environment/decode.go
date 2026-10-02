package environment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
)

type locatedInputError struct {
	Path   string
	Offset *int
	Err    error
}

func (e *locatedInputError) Error() string { return e.Err.Error() }
func (e *locatedInputError) Unwrap() error { return e.Err }
func at(path string, err error) error      { return &locatedInputError{Path: path, Err: err} }
func atOffset(path string, offset int64, err error) error {
	n := int(offset)
	return &locatedInputError{Path: path, Offset: &n, Err: err}
}

func pointerPart(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "~", "~0"), "/", "~1")
}

// decodeStrictJSON checks Unicode, duplicate members, and exact tagged names
// before Go's case-insensitive struct decoder can discard conflicting input.
func decodeStrictJSON(raw []byte, target any) error {
	if err := jsoninput.ValidateUnicode(raw); err != nil {
		for i := 0; i < len(raw); {
			_, size := utf8.DecodeRune(raw[i:])
			if size == 1 && raw[i] >= 0x80 {
				return atOffset("", int64(i), err)
			}
			i += size
		}
		if marker := strings.LastIndex(err.Error(), " at byte "); marker >= 0 {
			if n, parseErr := strconv.Atoi(err.Error()[marker+9:]); parseErr == nil {
				return atOffset("", int64(n), err)
			}
		}
		return at("", err)
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var read func(string) (any, error)
	read = func(path string) (any, error) {
		token, err := d.Token()
		if err != nil {
			return nil, atOffset(path, d.InputOffset(), fmt.Errorf("invalid JSON: %w", err))
		}
		switch token {
		case json.Delim('{'):
			object := map[string]any{}
			for d.More() {
				keyToken, err := d.Token()
				if err != nil {
					return nil, atOffset(path, d.InputOffset(), fmt.Errorf("invalid JSON key: %w", err))
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, atOffset(path, d.InputOffset(), fmt.Errorf("expected string object key"))
				}
				memberPath := path + "/" + pointerPart(key)
				if _, found := object[key]; found {
					return nil, atOffset(memberPath, d.InputOffset(), fmt.Errorf("duplicate property %q", key))
				}
				value, err := read(memberPath)
				if err != nil {
					return nil, err
				}
				object[key] = value
			}
			if _, err := d.Token(); err != nil {
				return nil, atOffset(path, d.InputOffset(), err)
			}
			return object, nil
		case json.Delim('['):
			items := []any{}
			for d.More() {
				value, err := read(fmt.Sprintf("%s/%d", path, len(items)))
				if err != nil {
					return nil, err
				}
				items = append(items, value)
			}
			if _, err := d.Token(); err != nil {
				return nil, atOffset(path, d.InputOffset(), err)
			}
			return items, nil
		default:
			return token, nil
		}
	}
	value, err := read("")
	if err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return atOffset("", d.InputOffset(), fmt.Errorf("expected exactly one JSON value"))
	}
	if err := validateExactMembers(value, reflect.TypeOf(target), ""); err != nil {
		return err
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) && typeError.Field != "" {
			path := typeErrorPointer(typeError.Field, reflect.TypeOf(target))
			return atOffset(path, typeError.Offset, fmt.Errorf("invalid artifact: %w", err))
		}
		return atOffset("", d.InputOffset(), fmt.Errorf("invalid artifact: %w", err))
	}
	if _, err := d.Token(); err != io.EOF {
		return atOffset("", d.InputOffset(), fmt.Errorf("expected exactly one JSON value"))
	}
	return nil
}

// UnmarshalTypeError.Field omits array indexes. Return no path if a field
// traverses an array, leaving the decoder's exact byte offset as the location.
func typeErrorPointer(field string, typ reflect.Type) string {
	path := ""
	for _, part := range strings.Split(field, ".") {
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		if typ.Kind() != reflect.Struct {
			return ""
		}
		found := false
		for i := 0; i < typ.NumField(); i++ {
			member := typ.Field(i)
			tag := strings.Split(member.Tag.Get("json"), ",")[0]
			if tag == part {
				typ = member.Type
				found = true
				break
			}
		}
		if !found {
			return ""
		}
		if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
			return ""
		}
		path += "/" + pointerPart(part)
	}
	return path
}

func validateExactMembers(value any, typ reflect.Type, path string) error {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflect.TypeOf(json.RawMessage{}) {
		return nil
	}
	if value == nil {
		return at(path, fmt.Errorf("null is not allowed"))
	}
	switch typ.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		} // typed decoder reports the type error
		fields := map[string]reflect.Type{}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}
			tag := strings.Split(field.Tag.Get("json"), ",")[0]
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = field.Name
			}
			fields[tag] = field.Type
		}
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			nested := object[key]
			fieldType, known := fields[key]
			memberPath := path + "/" + pointerPart(key)
			if !known {
				return at(memberPath, fmt.Errorf("unknown property %q", key))
			}
			if err := validateExactMembers(nested, fieldType, memberPath); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		items, ok := value.([]any)
		if !ok {
			return nil
		}
		for i, item := range items {
			if err := validateExactMembers(item, typ.Elem(), fmt.Sprintf("%s/%d", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func inputLocation(err error) (string, *int) {
	var located *locatedInputError
	if errors.As(err, &located) {
		return located.Path, located.Offset
	}
	return "", nil
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
			return at("/"+pointerPart(key), fmt.Errorf("missing property %q", key))
		}
		if bytes.Equal(value, []byte("null")) {
			return at("/"+pointerPart(key), fmt.Errorf("%s must not be null", key))
		}
	}
	for _, key := range []string{"capabilities", "collections", "objects"} {
		if len(root[key]) == 0 || root[key][0] != '[' {
			return at("/"+pointerPart(key), fmt.Errorf("%s must be an array", key))
		}
	}
	var version int
	_ = json.Unmarshal(root["schema_version"], &version)
	if observation, found := root["observation"]; found {
		if version == 1 {
			return at("/observation", fmt.Errorf("Snapshot v1 forbids observation"))
		}
		if err := validateObservationRawShape(observation); err != nil {
			return err
		}
	}
	var scope map[string]json.RawMessage
	if err := json.Unmarshal(root["capture_scope"], &scope); err != nil || scope == nil {
		return fmt.Errorf("capture_scope must be an object")
	}
	for _, name := range []string{"namespace", "app", "owner"} {
		rawSelector, found := scope[name]
		if !found {
			return at("/capture_scope/"+name, fmt.Errorf("missing capture_scope.%s", name))
		}
		var branches map[string]json.RawMessage
		if err := json.Unmarshal(rawSelector, &branches); err != nil || branches == nil {
			return at("/capture_scope/"+name, fmt.Errorf("%s selector must be an object", name))
		}
		all, hasAll := branches["all"]
		values, hasValues := branches["values"]
		if hasAll == hasValues {
			return at("/capture_scope/"+name, fmt.Errorf("%s selector must use exactly one branch", name))
		}
		if hasAll && bytes.Equal(all, []byte("null")) || hasValues && bytes.Equal(values, []byte("null")) {
			return at("/capture_scope/"+name, fmt.Errorf("%s selector must not be null", name))
		}
	}
	var objects []map[string]json.RawMessage
	if err := json.Unmarshal(root["objects"], &objects); err == nil {
		for i, object := range objects {
			if rawDocument, found := object["document"]; found {
				var document map[string]json.RawMessage
				if err := json.Unmarshal(rawDocument, &document); err != nil || document == nil {
					return at(fmt.Sprintf("/objects/%d/document", i), fmt.Errorf("document must be an object"))
				}
				if _, found := document["text"]; !found {
					return at(fmt.Sprintf("/objects/%d/document/text", i), fmt.Errorf("missing document text"))
				}
			}
		}
	}
	return nil
}
