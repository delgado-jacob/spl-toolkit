package closure

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type InputError struct{ Err error }

func (e *InputError) Error() string { return e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }
func IsInputError(err error) bool   { var e *InputError; return errors.As(err, &e) }
func inputError(format string, args ...any) error {
	return &InputError{Err: fmt.Errorf(format, args...)}
}

var objectKinds = map[string]bool{
	"macro": true, "lookup": true, "data_model": true, "dataset": true,
	"saved_search": true, "event_type": true, "tag": true, "calculated_field": true,
	"field_extraction": true, "module": true, "function": true, "external_command": true,
}

// decodeJSON rejects duplicate members at every depth before typed decoding.
func decodeJSON(data []byte) (any, error) {
	if err := jsoninput.ValidateUnicode(data); err != nil {
		return nil, inputError("%v", err)
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
			value := map[string]any{}
			for d.More() {
				keyToken, err := d.Token()
				if err != nil {
					return nil, inputError("invalid JSON key: %v", err)
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, inputError("expected a string object key")
				}
				if _, found := value[key]; found {
					return nil, inputError("duplicate property %q", key)
				}
				value[key], err = read()
				if err != nil {
					return nil, err
				}
			}
			if _, err := d.Token(); err != nil {
				return nil, inputError("invalid object: %v", err)
			}
			return value, nil
		case json.Delim('['):
			value := []any{}
			for d.More() {
				item, err := read()
				if err != nil {
					return nil, err
				}
				value = append(value, item)
			}
			if _, err := d.Token(); err != nil {
				return nil, inputError("invalid array: %v", err)
			}
			return value, nil
		default:
			return token, nil
		}
	}
	value, err := read()
	if err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, inputError("expected exactly one JSON value")
	}
	return value, nil
}

func fields(v any, required []string, allowed ...string) (map[string]any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, inputError("expected an object")
	}
	known := map[string]bool{}
	for _, key := range allowed {
		known[key] = true
	}
	for key := range m {
		if !known[key] {
			return nil, inputError("unknown property %q", key)
		}
	}
	for _, key := range required {
		if _, ok := m[key]; !ok {
			return nil, inputError("missing property %q", key)
		}
	}
	return m, nil
}
func stringField(m map[string]any, key string, required bool) (string, error) {
	v, found := m[key]
	if !found && !required {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", inputError("%s must be a string", key)
	}
	return s, nil
}
func integerField(m map[string]any, key string, required bool) (int, bool, error) {
	v, found := m[key]
	if !found {
		if required {
			return 0, false, inputError("missing property %q", key)
		}
		return 0, false, nil
	}
	n, ok := v.(json.Number)
	if !ok {
		return 0, true, inputError("%s must be an integer", key)
	}
	var out int
	if err := json.Unmarshal([]byte(n), &out); err != nil {
		return 0, true, inputError("%s must be an integer", key)
	}
	return out, true, nil
}
func arrayField(m map[string]any, key string, required bool) ([]any, error) {
	v, found := m[key]
	if !found && !required {
		return []any{}, nil
	}
	a, ok := v.([]any)
	if !ok {
		return nil, inputError("%s must be an array", key)
	}
	return a, nil
}
func nonblank(s, key string) error {
	if strings.TrimSpace(s) == "" {
		return inputError("%s must be nonempty", key)
	}
	if !utf8.ValidString(s) {
		return inputError("%s must be valid UTF-8", key)
	}
	return nil
}
func validKind(kind string) error {
	if !objectKinds[kind] {
		return inputError("unsupported object kind %q", kind)
	}
	return nil
}
func decodeDocument(v any) (analysis.QueryDocument, error) {
	m, err := fields(v, []string{"text"}, "text", "language", "profile", "version", "source_id")
	if err != nil {
		return analysis.QueryDocument{}, err
	}
	d := analysis.QueryDocument{}
	for _, item := range []struct {
		key  string
		dest *string
	}{
		{"text", &d.Text}, {"language", &d.Language}, {"profile", &d.Profile}, {"version", &d.Version}, {"source_id", &d.SourceID},
	} {
		*item.dest, err = stringField(m, item.key, item.key == "text")
		if err != nil {
			return analysis.QueryDocument{}, err
		}
	}
	return normalizeDocument(d)
}
func normalizeDocument(d analysis.QueryDocument) (analysis.QueryDocument, error) {
	selection, err := capabilityselector.Normalize(d.Language, d.Profile, d.Version)
	if err != nil {
		return analysis.QueryDocument{}, inputError("document selectors: %v", err)
	}
	if !utf8.ValidString(d.Text) || !utf8.ValidString(d.SourceID) {
		return analysis.QueryDocument{}, inputError("document text and source_id must be valid UTF-8")
	}
	d.Language, d.Profile, d.Version = selection.Language, selection.Profile, selection.Version
	return d, nil
}
func decodeCollection(v any) (Collection, error) {
	m, err := fields(v, []string{"kind", "coverage"}, "kind", "coverage")
	if err != nil {
		return Collection{}, err
	}
	kind, err := stringField(m, "kind", true)
	if err != nil {
		return Collection{}, err
	}
	coverage, err := stringField(m, "coverage", true)
	if err != nil {
		return Collection{}, err
	}
	return Collection{Kind: kind, Coverage: coverage}, nil
}
func decodeRelation(v any) (Relation, error) {
	m, err := fields(v, []string{"kind", "name"}, "kind", "name", "start", "end", "property")
	if err != nil {
		return Relation{}, err
	}
	r := Relation{}
	r.Kind, err = stringField(m, "kind", true)
	if err != nil {
		return r, err
	}
	r.Name, err = stringField(m, "name", true)
	if err != nil {
		return r, err
	}
	if _, found := m["start"]; found {
		n, _, err := integerField(m, "start", true)
		if err != nil {
			return r, err
		}
		r.Start = &n
	}
	if _, found := m["end"]; found {
		n, _, err := integerField(m, "end", true)
		if err != nil {
			return r, err
		}
		r.End = &n
	}
	if _, found := m["property"]; found {
		s, err := stringField(m, "property", true)
		if err != nil {
			return r, err
		}
		r.Property = &s
	}
	return r, nil
}
func decodeDefinition(v any) (Definition, error) {
	m, err := fields(v, []string{"id", "kind", "name", "source_id"}, "id", "kind", "name", "app", "owner", "sharing", "source_id", "document", "arity", "arguments", "eval_based", "validation", "relations")
	if err != nil {
		return Definition{}, err
	}
	o := Definition{Relations: []Relation{}}
	for _, item := range []struct {
		key      string
		dest     *string
		required bool
	}{
		{"id", &o.ID, true}, {"kind", &o.Kind, true}, {"name", &o.Name, true}, {"app", &o.App, false}, {"owner", &o.Owner, false}, {"sharing", &o.Sharing, false}, {"source_id", &o.SourceID, true},
	} {
		*item.dest, err = stringField(m, item.key, item.required)
		if err != nil {
			return Definition{}, err
		}
	}
	if o.Kind != "macro" {
		for _, key := range []string{"arity", "arguments", "eval_based", "validation"} {
			if _, found := m[key]; found {
				return Definition{}, inputError("%s is macro-only metadata", key)
			}
		}
	}
	if v, found := m["document"]; found {
		d, err := decodeDocument(v)
		if err != nil {
			return Definition{}, err
		}
		o.Document = &d
	}
	if _, found := m["arity"]; found {
		n, _, err := integerField(m, "arity", true)
		if err != nil {
			return Definition{}, err
		}
		o.Arity = &n
	}
	if _, found := m["arguments"]; found {
		values, err := arrayField(m, "arguments", true)
		if err != nil {
			return Definition{}, err
		}
		o.Arguments = make([]string, 0, len(values))
		for _, v := range values {
			s, ok := v.(string)
			if !ok {
				return Definition{}, inputError("argument names must be strings")
			}
			o.Arguments = append(o.Arguments, s)
		}
	}
	if v, found := m["eval_based"]; found {
		b, ok := v.(bool)
		if !ok {
			return Definition{}, inputError("eval_based must be a boolean")
		}
		o.EvalBased = &b
	}
	if _, found := m["validation"]; found {
		s, err := stringField(m, "validation", true)
		if err != nil {
			return Definition{}, err
		}
		o.Validation = &s
	}
	values, err := arrayField(m, "relations", false)
	if err != nil {
		return Definition{}, err
	}
	for _, v := range values {
		r, err := decodeRelation(v)
		if err != nil {
			return Definition{}, err
		}
		o.Relations = append(o.Relations, r)
	}
	return o, nil
}
func decodeBundle(v any) (DefinitionBundle, error) {
	m, err := fields(v, []string{"schema_version", "scope_id"}, "schema_version", "scope_id", "collections", "objects")
	if err != nil {
		return DefinitionBundle{}, err
	}
	version, _, err := integerField(m, "schema_version", true)
	if err != nil {
		return DefinitionBundle{}, err
	}
	scope, err := stringField(m, "scope_id", true)
	if err != nil {
		return DefinitionBundle{}, err
	}
	b := DefinitionBundle{SchemaVersion: version, ScopeID: scope, Collections: []Collection{}, Objects: []Definition{}}
	values, err := arrayField(m, "collections", false)
	if err != nil {
		return DefinitionBundle{}, err
	}
	for _, v := range values {
		c, err := decodeCollection(v)
		if err != nil {
			return DefinitionBundle{}, err
		}
		b.Collections = append(b.Collections, c)
	}
	values, err = arrayField(m, "objects", false)
	if err != nil {
		return DefinitionBundle{}, err
	}
	for _, v := range values {
		o, err := decodeDefinition(v)
		if err != nil {
			return DefinitionBundle{}, err
		}
		b.Objects = append(b.Objects, o)
	}
	return normalizeBundle(b)
}
func decodeBinding(v any) (Binding, error) {
	m, err := fields(v, []string{"document_digest", "kind", "start", "end", "object_id"}, "document_digest", "kind", "start", "end", "object_id")
	if err != nil {
		return Binding{}, err
	}
	b := Binding{}
	b.DocumentDigest, err = stringField(m, "document_digest", true)
	if err != nil {
		return b, err
	}
	b.Kind, err = stringField(m, "kind", true)
	if err != nil {
		return b, err
	}
	b.ObjectID, err = stringField(m, "object_id", true)
	if err != nil {
		return b, err
	}
	b.Start, _, err = integerField(m, "start", true)
	if err != nil {
		return b, err
	}
	b.End, _, err = integerField(m, "end", true)
	if err != nil {
		return b, err
	}
	return b, nil
}

// DecodeRequest validates the supplied bundle and structurally checks bindings.
// Matching bindings to exact analysis occurrences is performed by evaluation.
func DecodeRequest(data []byte) (Request, error) {
	value, err := decodeJSON(data)
	if err != nil {
		return Request{}, err
	}
	m, err := fields(value, []string{"schema_version", "document", "bundle", "bindings"}, "schema_version", "document", "bundle", "bindings")
	if err != nil {
		return Request{}, err
	}
	version, _, err := integerField(m, "schema_version", true)
	if err != nil {
		return Request{}, err
	}
	if version != 1 {
		return Request{}, inputError("schema_version must be integer 1")
	}
	document, err := decodeDocument(m["document"])
	if err != nil {
		return Request{}, err
	}
	bundle, err := decodeBundle(m["bundle"])
	if err != nil {
		return Request{}, err
	}
	values, err := arrayField(m, "bindings", true)
	if err != nil {
		return Request{}, err
	}
	bindings := make([]Binding, 0, len(values))
	objects := map[string]Definition{}
	for _, object := range bundle.Objects {
		objects[object.ID] = object
	}
	sum := sha256.Sum256([]byte(document.Text))
	digest := "sha256:" + hex.EncodeToString(sum[:])
	for _, value := range values {
		binding, err := decodeBinding(value)
		if err != nil {
			return Request{}, err
		}
		if err := validKind(binding.Kind); err != nil {
			return Request{}, err
		}
		object, found := objects[binding.ObjectID]
		if !found || object.Kind != binding.Kind || binding.DocumentDigest != digest || !validRange(document.Text, binding.Start, binding.End) {
			return Request{}, inputError("binding has invalid object, digest, or source range")
		}
		bindings = append(bindings, binding)
	}
	return Request{SchemaVersion: 1, Document: document, Bundle: bundle, Bindings: bindings}, nil
}

func validRange(text string, start, end int) bool {
	return start >= 0 && end > start && end <= len(text) && utf8.ValidString(text[:start]) && utf8.ValidString(text[:end])
}
func validPointer(pointer string) bool {
	if !strings.HasPrefix(pointer, "/") {
		return false
	}
	for i := 0; i < len(pointer); i++ {
		if pointer[i] == '~' {
			if i+1 >= len(pointer) || (pointer[i+1] != '0' && pointer[i+1] != '1') {
				return false
			}
			i++
		}
	}
	return true
}
func relationLess(a, b Relation) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	if a.Name != b.Name {
		return a.Name < b.Name
	}
	if a.Property != nil || b.Property != nil {
		if a.Property == nil {
			return true
		}
		if b.Property == nil {
			return false
		}
		return *a.Property < *b.Property
	}
	if *a.Start != *b.Start {
		return *a.Start < *b.Start
	}
	return *a.End < *b.End
}

// normalizeBundle is the shared canonical form for all caller surfaces and digests.
func normalizeBundle(input DefinitionBundle) (DefinitionBundle, error) {
	if input.SchemaVersion != 1 {
		return DefinitionBundle{}, inputError("bundle schema_version must be integer 1")
	}
	if err := nonblank(input.ScopeID, "scope_id"); err != nil {
		return DefinitionBundle{}, err
	}
	out := DefinitionBundle{SchemaVersion: 1, ScopeID: input.ScopeID, Collections: make([]Collection, 0, len(input.Collections)), Objects: make([]Definition, 0, len(input.Objects))}
	coverage := map[string]string{}
	for _, c := range input.Collections {
		if err := validKind(c.Kind); err != nil {
			return DefinitionBundle{}, err
		}
		if c.Coverage != "complete" && c.Coverage != "partial" && c.Coverage != "unavailable" {
			return DefinitionBundle{}, inputError("invalid collection coverage %q", c.Coverage)
		}
		if _, found := coverage[c.Kind]; found {
			return DefinitionBundle{}, inputError("duplicate or contradictory collection %q", c.Kind)
		}
		coverage[c.Kind] = c.Coverage
		out.Collections = append(out.Collections, c)
	}
	sort.Slice(out.Collections, func(i, j int) bool { return out.Collections[i].Kind < out.Collections[j].Kind })
	seen := map[string]bool{}
	for _, o := range input.Objects {
		if err := nonblank(o.ID, "object id"); err != nil {
			return DefinitionBundle{}, err
		}
		if seen[o.ID] {
			return DefinitionBundle{}, inputError("duplicate object id %q", o.ID)
		}
		seen[o.ID] = true
		if err := validKind(o.Kind); err != nil {
			return DefinitionBundle{}, err
		}
		for _, field := range []struct{ name, value string }{{"name", o.Name}, {"source_id", o.SourceID}} {
			if err := nonblank(field.value, field.name); err != nil {
				return DefinitionBundle{}, err
			}
		}
		for _, s := range []string{o.App, o.Owner, o.Sharing} {
			if !utf8.ValidString(s) {
				return DefinitionBundle{}, inputError("object metadata must be valid UTF-8")
			}
		}
		if o.Document != nil {
			d, err := normalizeDocument(*o.Document)
			if err != nil {
				return DefinitionBundle{}, err
			}
			o.Document = &d
		}
		if o.Kind == "macro" {
			if o.Arity == nil || *o.Arity < 0 || len(o.Arguments) != *o.Arity {
				return DefinitionBundle{}, inputError("macro arity and arguments are inconsistent")
			}
			args := map[string]bool{}
			for _, arg := range o.Arguments {
				if err := nonblank(arg, "argument name"); err != nil {
					return DefinitionBundle{}, err
				}
				if args[arg] {
					return DefinitionBundle{}, inputError("repeated macro argument name %q", arg)
				}
				args[arg] = true
			}
			o.Arguments = append([]string{}, o.Arguments...)
		} else if o.Arity != nil || len(o.Arguments) > 0 || o.EvalBased != nil || o.Validation != nil {
			return DefinitionBundle{}, inputError("macro metadata on non-macro object")
		}
		if o.Validation != nil && !utf8.ValidString(*o.Validation) {
			return DefinitionBundle{}, inputError("macro validation must be valid UTF-8")
		}
		relations := make([]Relation, 0, len(o.Relations))
		for _, r := range o.Relations {
			if err := validKind(r.Kind); err != nil {
				return DefinitionBundle{}, err
			}
			if err := nonblank(r.Name, "relation name"); err != nil {
				return DefinitionBundle{}, err
			}
			if r.Property != nil {
				if r.Start != nil || r.End != nil || !validPointer(*r.Property) || !utf8.ValidString(*r.Property) {
					return DefinitionBundle{}, inputError("invalid relation property evidence")
				}
			} else if r.Start == nil || r.End == nil || o.Document == nil || !validRange(o.Document.Text, *r.Start, *r.End) {
				return DefinitionBundle{}, inputError("invalid relation source evidence")
			}
			relations = append(relations, r)
		}
		sort.Slice(relations, func(i, j int) bool { return relationLess(relations[i], relations[j]) })
		o.Relations = relations
		out.Objects = append(out.Objects, o)
	}
	sort.Slice(out.Objects, func(i, j int) bool { return out.Objects[i].ID < out.Objects[j].ID })
	return out, nil
}
