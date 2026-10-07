package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func pointerPart(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}
func artifactPath(path string) bool {
	return path == "/snapshot" || path == "/schema_bundle" || path == "/settings/snapshot" || path == "/settings/schema_bundle"
}
func readWire(raw []byte, typ reflect.Type) (any, map[string]json.RawMessage, map[string]int, error) {
	if err := jsoninput.ValidateUnicode(raw); err != nil {
		return nil, nil, nil, requestErrorOffset("request_invalid", "", err.Error(), unicodeErrorOffset(raw))
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	artifacts := map[string]json.RawMessage{}
	bases := map[string]int{}
	var read func(string) (any, error)
	read = func(path string) (any, error) {
		if artifactPath(path) {
			var value json.RawMessage
			if err := d.Decode(&value); err != nil {
				return nil, requestErrorOffset("request_invalid", path, err.Error(), int(d.InputOffset()))
			}
			artifacts[path] = value
			bases[path] = int(d.InputOffset()) - len(value)
			return value, nil
		}
		token, err := d.Token()
		if err != nil {
			return nil, requestErrorOffset("request_invalid", path, err.Error(), int(d.InputOffset()))
		}
		switch token {
		case json.Delim('{'):
			values := map[string]any{}
			for d.More() {
				keyToken, err := d.Token()
				if err != nil {
					return nil, requestErrorOffset("request_invalid", path, err.Error(), int(d.InputOffset()))
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, requestErrorAt("request_invalid", path, "expected an object key")
				}
				member := path + "/" + pointerPart(key)
				if _, exists := values[key]; exists {
					return nil, requestErrorOffset("request_invalid", member, "duplicate property "+key, int(d.InputOffset()))
				}
				value, err := read(member)
				if err != nil {
					return nil, err
				}
				values[key] = value
			}
			if _, err := d.Token(); err != nil {
				return nil, requestErrorOffset("request_invalid", path, err.Error(), int(d.InputOffset()))
			}
			return values, nil
		case json.Delim('['):
			values := []any{}
			for d.More() {
				value, err := read(fmt.Sprintf("%s/%d", path, len(values)))
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			}
			if _, err := d.Token(); err != nil {
				return nil, requestErrorOffset("request_invalid", path, err.Error(), int(d.InputOffset()))
			}
			return values, nil
		default:
			return token, nil
		}
	}
	value, err := read("")
	if err != nil {
		return nil, nil, nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, nil, nil, requestErrorOffset("request_invalid", "", "expected exactly one JSON value", int(d.InputOffset()))
	}
	if err := validateWireShape(value, typ, ""); err != nil {
		return nil, nil, nil, err
	}
	return value, artifacts, bases, nil
}
func DecodeRequest(raw []byte) (Request, error) {
	_, artifacts, bases, err := readWire(raw, reflect.TypeOf(Request{}))
	if err != nil {
		return Request{}, err
	}
	var admitted Settings
	if err := decodeArtifacts(&admitted, artifacts, bases, "/settings"); err != nil {
		return Request{}, err
	}
	var request Request
	if err := json.Unmarshal(raw, &request); err != nil {
		return Request{}, requestErrorAt("request_invalid", "", err.Error())
	}
	request.Settings.Snapshot, request.Settings.SchemaBundle = admitted.Snapshot, admitted.SchemaBundle
	return normalizeRequest(request)
}
func DecodeSettings(raw []byte) (Settings, error) {
	_, artifacts, bases, err := readWire(raw, reflect.TypeOf(Settings{}))
	if err != nil {
		return Settings{}, err
	}
	var admitted Settings
	if err := decodeArtifacts(&admitted, artifacts, bases, ""); err != nil {
		return Settings{}, err
	}
	var settings Settings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return Settings{}, requestErrorAt("request_invalid", "", err.Error())
	}
	settings.Snapshot, settings.SchemaBundle = admitted.Snapshot, admitted.SchemaBundle
	return normalizeSettings(settings)
}
func decodeArtifacts(settings *Settings, artifacts map[string]json.RawMessage, bases map[string]int, base string) error {
	snapshot, report := environment.DecodeSnapshot(artifacts[base+"/snapshot"])
	if err := artifactError(report, base+"/snapshot", bases[base+"/snapshot"]); err != nil {
		return err
	}
	settings.Snapshot = snapshot
	if raw, ok := artifacts[base+"/schema_bundle"]; ok {
		bundle, report := environment.DecodeSchemaBundle(raw)
		if err := artifactError(report, base+"/schema_bundle", bases[base+"/schema_bundle"]); err != nil {
			return err
		}
		settings.SchemaBundle = &bundle
	}
	return nil
}

// normalizeRequest checks typed input before detachment can replace invalid text.
func normalizeRequest(input Request) (Request, error) {
	if !validUTF8(reflect.ValueOf(input)) {
		return Request{}, requestErrorAt("request_invalid", "", "request contains invalid UTF-8")
	}
	if input.SchemaVersion != 1 {
		return Request{}, requestErrorAt("request_invalid", "/schema_version", "schema_version must be integer 1")
	}
	if len(input.Documents) == 0 {
		return Request{}, requestErrorAt("request_invalid", "/documents", "documents must be a nonempty array")
	}
	switch input.Format {
	case "", "text", "json", "sarif", "graph", "bom":
	default:
		return Request{}, requestErrorAt("request_invalid", "/format", "unsupported format")
	}
	settings, err := normalizeSettingsAt(input.Settings, "/settings")
	if err != nil {
		return Request{}, err
	}
	out := detach(input)
	out.Settings = settings
	seen := map[string]bool{}
	for i := range out.Documents {
		entry := &out.Documents[i]
		path := fmt.Sprintf("/documents/%d", i)
		if strings.TrimSpace(entry.ID) == "" || seen[entry.ID] {
			return Request{}, requestErrorAt("request_invalid", path+"/id", "id must be unique and nonblank")
		}
		seen[entry.ID] = true
		selected, err := capabilityselector.Normalize(entry.Document.Language, entry.Document.Profile, entry.Document.Version)
		if err != nil {
			return Request{}, requestErrorAt("request_invalid", path+"/document", err.Error())
		}
		entry.Document.Language, entry.Document.Profile, entry.Document.Version = selected.Language, selected.Profile, selected.Version
	}
	configured := map[string]bool{}
	for i, e := range settings.Entries {
		if !seen[e.ID] {
			return Request{}, requestErrorAt("request_invalid", fmt.Sprintf("/settings/entries/%d/id", i), "settings id is not selected")
		}
		configured[e.ID] = true
	}
	for i, e := range out.Documents {
		if !configured[e.ID] {
			return Request{}, requestErrorAt("request_invalid", fmt.Sprintf("/documents/%d/id", i), "missing entry settings")
		}
	}
	return out, nil
}
func normalizeSettings(input Settings) (Settings, error) { return normalizeSettingsAt(input, "") }
func normalizeSettingsAt(input Settings, base string) (Settings, error) {
	if !validUTF8(reflect.ValueOf(input)) {
		return Settings{}, requestErrorAt("request_invalid", base, "settings contain invalid UTF-8")
	}
	if input.SchemaVersion != 1 {
		return Settings{}, requestErrorAt("request_invalid", base+"/schema_version", "schema_version must be integer 1")
	}
	if input.Entries == nil {
		return Settings{}, requestErrorAt("request_invalid", base+"/entries", "entries must be an array")
	}
	// Admit typed artifact shapes using their canonical owners, before copying.
	if input.Snapshot.Capabilities == nil || input.Snapshot.Collections == nil || input.Snapshot.Objects == nil {
		return Settings{}, requestErrorAt("request_invalid", base+"/snapshot", "capabilities, collections and objects must be arrays")
	}
	if input.SchemaBundle != nil && (input.SchemaBundle.Schemas == nil || input.SchemaBundle.Bindings == nil) {
		return Settings{}, requestErrorAt("request_invalid", base+"/schema_bundle", "schemas and bindings must be arrays")
	}
	out := detach(input)
	seen := map[string]bool{}
	for i, e := range out.Entries {
		path := fmt.Sprintf("%s/entries/%d", base, i)
		if strings.TrimSpace(e.ID) == "" || seen[e.ID] {
			return Settings{}, requestErrorAt("request_invalid", path+"/id", "id must be unique and nonblank")
		}
		seen[e.ID] = true
		if (e.Compatibility == nil) == (e.Resolution == nil) {
			return Settings{}, requestErrorAt("request_invalid", path, "exactly one mode payload is required")
		}
		var scope environment.CaptureScope
		if e.Compatibility != nil {
			if e.Compatibility.InputBindings == nil {
				return Settings{}, requestErrorAt("request_invalid", path+"/compatibility/input_bindings", "input_bindings must be an array")
			}
			scope = e.Compatibility.QueryScope
			path += "/compatibility/query_scope"
		} else {
			if e.Resolution.Resolutions == nil || e.Resolution.Compatibility.InputBindings == nil {
				return Settings{}, requestErrorAt("request_invalid", path+"/resolution", "resolutions and input_bindings must be arrays")
			}
			for j, choice := range e.Resolution.Resolutions {
				if choice.Values == nil {
					return Settings{}, requestErrorAt("request_invalid", fmt.Sprintf("%s/resolution/resolutions/%d/values", path, j), "values must be an array")
				}
			}
			scope = e.Resolution.Compatibility.QueryScope
			path += "/resolution/compatibility/query_scope"
		}
		for _, sel := range []struct {
			name  string
			value environment.Selector
		}{{"namespace", scope.Namespace}, {"app", scope.App}, {"owner", scope.Owner}} {
			v := sel.value
			if v.All != nil {
				if !*v.All || v.Values != nil {
					return Settings{}, requestErrorAt("request_invalid", path+"/"+sel.name, "selector must use exactly one branch")
				}
			} else {
				if len(v.Values) == 0 {
					return Settings{}, requestErrorAt("request_invalid", path+"/"+sel.name, "selector requires nonempty values")
				}

			}
		}
	}
	return out, nil
}
func validateWireShape(value any, typ reflect.Type, path string) error {
	if artifactPath(path) {
		return nil
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	fail := func(message string) error { return requestErrorAt("request_invalid", path, message) }
	if value == nil {
		return requestErrorAt("request_invalid", path, "null is not allowed")
	}
	if typ == reflect.TypeOf(json.RawMessage{}) {
		return nil
	}
	switch typ.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return fail("expected an object")
		}
		fields := map[string]reflect.StructField{}
		ordered := []string{}
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if !f.IsExported() {
				continue
			}
			tag := strings.Split(f.Tag.Get("json"), ",")[0]
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = f.Name
			}
			fields[tag] = f
			ordered = append(ordered, tag)
		}
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if _, known := fields[key]; !known {
				return requestErrorAt("request_invalid", path+"/"+pointerPart(key), "unknown property "+key)
			}
		}
		for _, key := range ordered {
			f := fields[key]
			nested, present := object[key]
			required := !strings.Contains(f.Tag.Get("json"), ",omitempty")
			if typ == reflect.TypeOf(analysis.QueryDocument{}) {
				required = key == "text"
			}
			if !present {
				if required {
					return requestErrorAt("request_invalid", path+"/"+key, "missing property "+key)
				}
				continue
			}
			if err := validateWireShape(nested, f.Type, path+"/"+pointerPart(key)); err != nil {
				return err
			}
		}

	case reflect.Slice, reflect.Array:
		items, ok := value.([]any)
		if !ok {
			return fail("expected an array")
		}
		for i, item := range items {
			if err := validateWireShape(item, typ.Elem(), fmt.Sprintf("%s/%d", path, i)); err != nil {
				return err
			}
		}
	case reflect.String:
		if _, ok := value.(string); !ok {
			return fail("expected a string")
		}
	case reflect.Bool:
		if _, ok := value.(bool); !ok {
			return fail("expected a boolean")
		}
	case reflect.Uint64:
		n, ok := value.(json.Number)
		if !ok {
			return fail("expected an unsigned integer")
		}
		if _, err := strconv.ParseUint(string(n), 10, 64); err != nil {
			return fail("expected an unsigned integer")
		}
	case reflect.Int:
		n, ok := value.(json.Number)
		if !ok {
			return fail("expected an integer")
		}
		if _, err := strconv.Atoi(string(n)); err != nil {
			return fail("expected an integer")
		}
	}
	return nil
}

func validUTF8(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return utf8.ValidString(value.String())
	case reflect.Pointer, reflect.Interface:
		return value.IsNil() || validUTF8(value.Elem())
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Type().Field(i).IsExported() && !validUTF8(value.Field(i)) {
				return false
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if !validUTF8(value.Index(i)) {
				return false
			}
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			if !validUTF8(key) || !validUTF8(value.MapIndex(key)) {
				return false
			}
		}
	}
	return true
}

// Reflection copying preserves nil versus empty unions and detaches all nested
// slices, maps, and pointers without lossy JSON omitempty round trips.
func detach[T any](value T) T { return copyValue(reflect.ValueOf(value)).Interface().(T) }
func copyValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		out := reflect.New(value.Type().Elem())
		out.Elem().Set(copyValue(value.Elem()))
		return out
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		out := reflect.New(value.Type()).Elem()
		out.Set(copyValue(value.Elem()))
		return out
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			out.Index(i).Set(copyValue(value.Index(i)))
		}
		return out
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		out := reflect.MakeMapWithSize(value.Type(), value.Len())
		for _, key := range value.MapKeys() {
			out.SetMapIndex(copyValue(key), copyValue(value.MapIndex(key)))
		}
		return out
	case reflect.Struct:
		out := reflect.New(value.Type()).Elem()
		for i := 0; i < value.NumField(); i++ {
			out.Field(i).Set(copyValue(value.Field(i)))
		}
		return out
	default:
		return value
	}
}

// unicodeErrorOffset locates the encoding fault already identified by ValidateUnicode.
// It reads source bytes rather than depending on the owner's human-readable error.
func unicodeErrorOffset(raw []byte) int {
	if !utf8.Valid(raw) {
		for i := 0; i < len(raw); {
			_, size := utf8.DecodeRune(raw[i:])
			if size == 1 && raw[i] >= 0x80 {
				return i
			}
			i += size
		}
	}
	escape := func(i int) (uint64, bool) {
		if i+6 > len(raw) || raw[i] != '\\' || raw[i+1] != 'u' {
			return 0, false
		}
		n, err := strconv.ParseUint(string(raw[i+2:i+6]), 16, 16)
		return n, err == nil
	}
	inString := false
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || i+1 >= len(raw) {
				continue
			}
			if raw[i+1] != 'u' {
				i++
				continue
			}
			n, ok := escape(i)
			if !ok {
				continue
			}
			if n >= 0xd800 && n <= 0xdbff {
				low, paired := escape(i + 6)
				if !paired || low < 0xdc00 || low > 0xdfff {
					return i
				}
				i += 11
			} else if n >= 0xdc00 && n <= 0xdfff {
				return i
			} else {
				i += 5
			}
		}
	}
	return 0
}
