package resolution

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func pointerPart(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}

// DecodeRequest admits the exact wire shape before Go decoding can discard
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
		if path == "/compatibility/snapshot" || path == "/compatibility/schema_bundle" {
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
	snapshot, report := environment.DecodeSnapshot(artifacts["/compatibility/snapshot"])
	if err := artifactError(report, "/compatibility/snapshot", bases["/compatibility/snapshot"]); err != nil {
		return Request{}, err
	}
	var schemas *environment.SchemaBundle
	if nested, present := artifacts["/compatibility/schema_bundle"]; present {
		bundle, report := environment.DecodeSchemaBundle(nested)
		if err := artifactError(report, "/compatibility/schema_bundle", bases["/compatibility/schema_bundle"]); err != nil {
			return Request{}, err
		}
		schemas = &bundle
	}
	var request Request
	// The exact shape and artifact owners have already admitted every member.
	if err := json.Unmarshal(raw, &request); err != nil {
		return Request{}, requestErrorAt("request_invalid", "", err.Error())
	}
	request.Compatibility.Snapshot = snapshot
	request.Compatibility.SchemaBundle = schemas
	return normalizeRequest(request)
}

func validateWireShape(value any, typ reflect.Type, path string) error {
	if path == "/compatibility/snapshot" || path == "/compatibility/schema_bundle" {
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

// normalizeRequest admits typed Go input before any query parsing or enumeration.
func normalizeRequest(input Request) (Request, error) { return normalizeRequestArtifacts(input, true) }

// Shape admission leaves reusable artifact compilation to Prepare at the operation boundary.
func normalizeRequestShape(input Request) (Request, error) {
	return normalizeRequestArtifacts(input, false)
}

func normalizeRequestArtifacts(input Request, validateArtifacts bool) (Request, error) {
	if !validUTF8(reflect.ValueOf(input)) {
		return Request{}, requestErrorAt("request_invalid", "", "request contains invalid UTF-8")
	}
	if input.SchemaVersion != 1 {
		return Request{}, requestErrorAt("request_invalid", "/schema_version", "schema_version must be integer 1")
	}
	if input.Resolutions == nil {
		return Request{}, requestErrorAt("request_invalid", "/resolutions", "resolutions must be an array")
	}
	if input.Compatibility.InputBindings == nil {
		return Request{}, requestErrorAt("request_invalid", "/compatibility/input_bindings", "input_bindings must be an array")
	}
	for _, array := range []struct {
		name    string
		missing bool
	}{{"capabilities", input.Compatibility.Snapshot.Capabilities == nil}, {"collections", input.Compatibility.Snapshot.Collections == nil}, {"objects", input.Compatibility.Snapshot.Objects == nil}} {
		if array.missing {
			return Request{}, requestErrorAt("request_invalid", "/compatibility/snapshot/"+array.name, array.name+" must be an array")
		}
	}
	if input.Compatibility.SchemaBundle != nil && (input.Compatibility.SchemaBundle.Schemas == nil || input.Compatibility.SchemaBundle.Bindings == nil) {
		return Request{}, requestErrorAt("request_invalid", "/compatibility/schema_bundle", "schemas and bindings must be arrays")
	}
	if validateArtifacts {
		_, report, err := environment.PrepareSnapshot(input.Compatibility.Snapshot)
		if err != nil {
			return Request{}, err
		}
		if err := artifactError(report, "/compatibility/snapshot", 0); err != nil {
			return Request{}, err
		}
		if input.Compatibility.SchemaBundle != nil {
			_, report, err := environment.PrepareSchemaBundle(*input.Compatibility.SchemaBundle)
			if err != nil {
				return Request{}, err
			}
			if err := artifactError(report, "/compatibility/schema_bundle", 0); err != nil {
				return Request{}, err
			}
		}
	}
	normalized, err := normalizePreparedRequest(PreparedRequest{SchemaVersion: input.SchemaVersion, Document: input.Document, Resolutions: input.Resolutions, MaxVariants: input.MaxVariants, Compatibility: compatibility.ResolutionAssessment{QueryScope: input.Compatibility.QueryScope, InputBindings: input.Compatibility.InputBindings, DependencyBindings: input.Compatibility.DependencyBindings}})
	if err != nil {
		return Request{}, err
	}
	out := detach(input)
	out.Document, out.Resolutions, out.MaxVariants = normalized.Document, normalized.Resolutions, normalized.MaxVariants
	out.Compatibility.QueryScope, out.Compatibility.InputBindings, out.Compatibility.DependencyBindings = normalized.Compatibility.QueryScope, normalized.Compatibility.InputBindings, normalized.Compatibility.DependencyBindings
	return out, nil
}

func normalizePreparedRequest(input PreparedRequest) (PreparedRequest, error) {
	if !validUTF8(reflect.ValueOf(input)) {
		return PreparedRequest{}, requestErrorAt("request_invalid", "", "request contains invalid UTF-8")
	}
	if input.SchemaVersion != 1 {
		return PreparedRequest{}, requestErrorAt("request_invalid", "/schema_version", "schema_version must be integer 1")
	}
	if input.Resolutions == nil {
		return PreparedRequest{}, requestErrorAt("request_invalid", "/resolutions", "resolutions must be an array")
	}
	if input.Compatibility.InputBindings == nil {
		return PreparedRequest{}, requestErrorAt("request_invalid", "/compatibility/input_bindings", "input_bindings must be an array")
	}
	out := detach(input)
	selection, err := capabilityselector.Normalize(out.Document.Language, out.Document.Profile, out.Document.Version)
	if err != nil {
		return PreparedRequest{}, requestErrorAt("request_invalid", "/document", err.Error())
	}
	out.Document.Language, out.Document.Profile, out.Document.Version = selection.Language, selection.Profile, selection.Version
	seen := map[string]bool{}
	for i, r := range out.Resolutions {
		path := fmt.Sprintf("/resolutions/%d", i)
		if !markerPattern.MatchString(r.Placeholder) || seen[r.Placeholder] {
			return PreparedRequest{}, requestErrorAt("request_invalid", path+"/placeholder", "placeholder must be a unique named marker")
		}
		seen[r.Placeholder] = true
		if !supportedChoiceKind(r.Kind) {
			return PreparedRequest{}, requestErrorAt("request_invalid", path+"/kind", "unsupported resolution kind")
		}
		if len(r.Values) == 0 {
			return PreparedRequest{}, requestErrorAt("request_invalid", path+"/values", "values must be nonempty")
		}
		values := map[string]bool{}
		for j, v := range r.Values {
			if !nonblank(v) || values[v] {
				return PreparedRequest{}, requestErrorAt("request_invalid", fmt.Sprintf("%s/values/%d", path, j), "values must be unique and nonblank")
			}
			values[v] = true
		}
	}
	for _, sel := range []struct {
		name  string
		value *environment.Selector
	}{{"namespace", &out.Compatibility.QueryScope.Namespace}, {"app", &out.Compatibility.QueryScope.App}, {"owner", &out.Compatibility.QueryScope.Owner}} {
		normalized, err := normalizeSelector(*sel.value, "/compatibility/query_scope/"+sel.name)
		if err != nil {
			return PreparedRequest{}, err
		}
		*sel.value = normalized
	}
	bindings := map[string]map[string]bool{}
	objects := map[string]environment.ObjectIdentity{}
	for i, b := range out.Compatibility.InputBindings {
		path := fmt.Sprintf("/compatibility/input_bindings/%d", i)
		if !nonblank(b.OriginalInputID) || !nonblank(b.ObjectID) || !supportedBindingKind(b.Expected.Kind) || !nonblank(b.Expected.Name) || (b.Expected.Kind != "dataset" && (b.Expected.Namespace != "" || b.Expected.App != "" || b.Expected.Owner != "")) || (b.ResolvedValue != nil && !nonblank(*b.ResolvedValue)) {
			return PreparedRequest{}, requestErrorAt("binding_invalid", path, "invalid resolution binding")
		}
		if bindings[b.OriginalInputID] == nil {
			bindings[b.OriginalInputID] = map[string]bool{}
		}
		key := "plain"
		if b.ResolvedValue != nil {
			key = "value:" + *b.ResolvedValue
		}
		entries := bindings[b.OriginalInputID]
		if entries[key] || (key == "plain" && len(entries) > 0) || entries["plain"] {
			return PreparedRequest{}, requestErrorAt("binding_invalid", path, "duplicate or contradictory binding key")
		}
		entries[key] = true
		if previous, ok := objects[b.ObjectID]; ok && previous != b.Expected {
			return PreparedRequest{}, requestErrorAt("binding_invalid", path+"/object_id", "contradictory expected identities")
		}
		objects[b.ObjectID] = b.Expected
	}
	dependencies := map[string]bool{}
	for i, b := range out.Compatibility.DependencyBindings {
		path := fmt.Sprintf("/compatibility/dependency_bindings/%d", i)
		digest, e := hex.DecodeString(strings.TrimPrefix(b.DocumentDigest, "sha256:"))
		key := fmt.Sprintf("%s/%s/%d/%d", b.DocumentDigest, b.Kind, b.Start, b.End)
		if !strings.HasPrefix(b.DocumentDigest, "sha256:") || b.DocumentDigest != strings.ToLower(b.DocumentDigest) || e != nil || len(digest) != sha256.Size || !nonblank(b.Kind) || !nonblank(b.ObjectID) || b.Start < 0 || b.End <= b.Start || dependencies[key] {
			return PreparedRequest{}, requestErrorAt("binding_invalid", path, "invalid dependency binding")
		}
		dependencies[key] = true
	}
	if _, _, err := admitFanout(out.Resolutions, out.MaxVariants); err != nil {
		return PreparedRequest{}, err
	}
	return out, nil
}

var markerPattern = regexp.MustCompile(`^\$[A-Za-z_][A-Za-z0-9_]*$`)

func supportedChoiceKind(kind string) bool {
	return supportedBindingKind(kind) || kind == "lookup" || kind == "data_model"
}
func supportedBindingKind(kind string) bool {
	return kind == "dataset" || kind == "index" || kind == "source" || kind == "sourcetype"
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
