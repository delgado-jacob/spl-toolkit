package compatibility

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
func shapeCode(path string) string {
	if strings.HasPrefix(path, "/requirements/") {
		return "requirements_refresh_required"
	}
	return "request_invalid"
}

// DecodeRequest admits the exact flat wire shape before Go decoding can discard
// duplicate keys, property case, required arrays, or raw artifact union branches.
func DecodeRequest(raw []byte) (Request, error) {
	if err := jsoninput.ValidateUnicode(raw); err != nil {
		offset := 0
		for offset < len(raw) {
			_, size := utf8.DecodeRune(raw[offset:])
			if size == 1 && raw[offset] >= 0x80 {
				break
			}
			offset += size
		}
		if marker := strings.LastIndex(err.Error(), " at byte "); marker >= 0 {
			if n, e := strconv.Atoi(err.Error()[marker+9:]); e == nil {
				offset = n
			}
		}
		return Request{}, requestErrorOffset("request_invalid", "", err.Error(), offset)
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	artifacts := map[string]json.RawMessage{}
	bases := map[string]int{}
	var read func(string) (any, error)
	read = func(path string) (any, error) {
		if path == "/snapshot" || path == "/schema_bundle" {
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
		return Request{}, err
	}
	if _, err := d.Token(); err != io.EOF {
		return Request{}, requestErrorOffset("request_invalid", "", "expected exactly one JSON value", int(d.InputOffset()))
	}
	if err := validateWireShape(value, reflect.TypeOf(Request{}), ""); err != nil {
		return Request{}, err
	}
	// Artifact owner decoders see the original bytes, including explicit empty
	// selector branches and arbitrary schema keywords, before any remarshal.
	snapshot, report := environment.DecodeSnapshot(artifacts["/snapshot"])
	if err := artifactError(report, "/snapshot", bases["/snapshot"]); err != nil {
		return Request{}, err
	}
	var schemas *environment.SchemaBundle
	if nested, present := artifacts["/schema_bundle"]; present {
		bundle, report := environment.DecodeSchemaBundle(nested)
		if err := artifactError(report, "/schema_bundle", bases["/schema_bundle"]); err != nil {
			return Request{}, err
		}
		schemas = &bundle
	}
	var request Request
	// The exact shape and artifact owners have already admitted every member.
	if err := json.Unmarshal(raw, &request); err != nil {
		return Request{}, requestErrorAt("request_invalid", "", err.Error())
	}
	request.Snapshot = snapshot
	request.SchemaBundle = schemas
	assessment, err := normalizeAssessment(request.assessment())
	if err != nil {
		return Request{}, err
	}
	request.setAssessment(assessment)
	return request, nil
}

func validateWireShape(value any, typ reflect.Type, path string) error {
	if path == "/snapshot" || path == "/schema_bundle" {
		return nil
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	fail := func(message string) error { return requestErrorAt("request_invalid", path, message) }
	if value == nil {
		return requestErrorAt(shapeCode(path), path, "null is not allowed")
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
					return requestErrorAt(shapeCode(path+"/"+key), path+"/"+key, "missing property "+key)
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
func (r Request) assessment() AssessmentRequest {
	return AssessmentRequest{SchemaVersion: r.SchemaVersion, Requirements: r.Requirements, QueryScope: r.QueryScope, InputBindings: r.InputBindings, Document: r.Document, DependencyBindings: r.DependencyBindings}
}
func (r *Request) setAssessment(a AssessmentRequest) {
	r.SchemaVersion = a.SchemaVersion
	r.Requirements = a.Requirements
	r.QueryScope = a.QueryScope
	r.InputBindings = a.InputBindings
	r.Document = a.Document
	r.DependencyBindings = a.DependencyBindings
}

func normalizeAssessment(input AssessmentRequest) (AssessmentRequest, error) {
	if !validUTF8(reflect.ValueOf(input)) {
		return AssessmentRequest{}, requestErrorAt("request_invalid", "", "request contains invalid UTF-8")
	}
	if input.SchemaVersion != 1 {
		return AssessmentRequest{}, requestErrorAt("request_invalid", "/schema_version", "schema_version must be integer 1")
	}
	if err := validateRequirements(input.Requirements); err != nil {
		return AssessmentRequest{}, err
	}
	if input.InputBindings == nil {
		return AssessmentRequest{}, requestErrorAt("request_invalid", "/input_bindings", "input_bindings must be an array")
	}
	out := detach(input)
	for _, sel := range []struct {
		name  string
		value *environment.Selector
	}{{"namespace", &out.QueryScope.Namespace}, {"app", &out.QueryScope.App}, {"owner", &out.QueryScope.Owner}} {
		normalized, err := normalizeSelector(*sel.value, "/query_scope/"+sel.name)
		if err != nil {
			return AssessmentRequest{}, err
		}
		*sel.value = normalized
	}
	if out.Document != nil {
		selection, err := capabilityselector.Normalize(out.Document.Language, out.Document.Profile, out.Document.Version)
		if err != nil {
			return AssessmentRequest{}, requestErrorAt("request_invalid", "/document", err.Error())
		}
		out.Document.Language = selection.Language
		out.Document.Profile = selection.Profile
		out.Document.Version = selection.Version
	}
	for i, b := range out.DependencyBindings {
		path := fmt.Sprintf("/dependency_bindings/%d", i)
		if !validDigest(b.DocumentDigest) || !nonblank(b.Kind) || !nonblank(b.ObjectID) || b.Start < 0 || b.End <= b.Start {
			return AssessmentRequest{}, requestErrorAt("request_invalid", path, "invalid dependency binding")
		}
	}
	return out, nil
}
func normalizeSelector(input environment.Selector, path string) (environment.Selector, error) {
	if input.All != nil {
		if !*input.All || input.Values != nil {
			return environment.Selector{}, requestErrorAt("request_invalid", path, "selector must use exactly one branch")
		}
		yes := true
		return environment.Selector{All: &yes}, nil
	}
	if len(input.Values) == 0 {
		return environment.Selector{}, requestErrorAt("request_invalid", path, "selector requires all or nonempty values")
	}
	out := environment.Selector{Values: append([]string{}, input.Values...)}
	seen := map[string]bool{}
	for _, v := range out.Values {
		if !nonblank(v) || seen[v] {
			return environment.Selector{}, requestErrorAt("request_invalid", path, "selector requires unique nonblank values")
		}
		seen[v] = true
	}
	sort.Strings(out.Values)
	return out, nil
}
func nonblank(s string) bool { return utf8.ValidString(s) && strings.TrimSpace(s) != "" }
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
