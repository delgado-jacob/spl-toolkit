package validation

import (
	"encoding/json"
	"math/big"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type schemaDocument struct {
	raw   any
	nodes map[string]*schemaNode
}
type schemaNode struct {
	value                               any
	object                              map[string]any
	doc                                 *schemaDocument
	pointer, resourcePointer, base, uri string
	children                            map[string]*schemaNode
	patterns                            map[string]schemaPattern
	refs                                map[string]*schemaNode
}
type schemaIndex struct {
	resources   map[string]*schemaNode
	anchors     map[string]*schemaNode
	nodes       []*schemaNode
	limitations map[string]bool
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func pointerEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}
func parseSchemaURI(s string) (*url.URL, error) {
	u, e := url.Parse(s)
	if e != nil {
		return nil, inputError("invalid schema URI %q", s)
	}
	if _, e = pointerFragment(u.Fragment); e != nil {
		return nil, e
	}
	return u, nil
}
func pointerFragment(s string) ([]string, error) {
	if s == "" || !strings.HasPrefix(s, "/") {
		return nil, nil
	}
	parts := strings.Split(s[1:], "/")
	for i, p := range parts {
		var b strings.Builder
		for j := 0; j < len(p); j++ {
			if p[j] == '~' {
				j++
				if j == len(p) || (p[j] != '0' && p[j] != '1') {
					return nil, inputError("invalid JSON Pointer escape")
				}
				if p[j] == '0' {
					b.WriteByte('~')
				} else {
					b.WriteByte('/')
				}
			} else {
				b.WriteByte(p[j])
			}
		}
		parts[i] = b.String()
	}
	return parts, nil
}
func resolveURI(base, ref string) (string, error) {
	r, e := parseSchemaURI(ref)
	if e != nil {
		return "", e
	}
	b, e := url.Parse(base)
	if e != nil {
		return "", inputError("invalid base URI")
	}
	if r.IsAbs() {
		return r.String(), nil
	}
	if ref == "" || strings.HasPrefix(ref, "#") {
		b.Fragment = r.Fragment
		b.RawFragment = r.RawFragment
		return b.String(), nil
	}
	if !b.IsAbs() || b.Opaque != "" {
		return "", nil
	}
	return b.ResolveReference(r).String(), nil
}
func (idx *schemaIndex) register(uri string, n *schemaNode) error {
	if old := idx.resources[uri]; old != nil && old != n {
		return inputError("conflicting schema resource %q", uri)
	}
	idx.resources[uri] = n
	return nil
}
func newSchemaIndex(t SchemaTarget) (*schemaIndex, *schemaNode, error) {
	idx := &schemaIndex{resources: map[string]*schemaNode{}, anchors: map[string]*schemaNode{}, limitations: map[string]bool{}}
	base := t.BaseURI
	if base == "" {
		base = jsonSchemaRoot
	} else {
		u, e := parseSchemaURI(base)
		if e != nil {
			return nil, nil, e
		}
		if !u.IsAbs() || u.Fragment != "" {
			return nil, nil, inputError("base_uri must be absolute without fragment")
		}
	}
	build := func(raw json.RawMessage, base string) (*schemaNode, error) {
		v, e := decodeUniqueJSON(raw)
		if e != nil {
			return nil, e
		}
		d := &schemaDocument{raw: v, nodes: map[string]*schemaNode{}}
		return idx.build(v, d, "", base, base, "")
	}
	root, e := build(t.Schema, base)
	if e != nil {
		return nil, nil, e
	}
	if t.BaseURI != "" || root.base == jsonSchemaRoot {
		if e = idx.register(base, root); e != nil {
			return nil, nil, e
		}
	}
	for _, uri := range sortedKeys(t.Resources) {
		u, e := parseSchemaURI(uri)
		if e != nil {
			return nil, nil, e
		}
		if !u.IsAbs() || u.Fragment != "" {
			return nil, nil, inputError("resource key must be absolute without fragment")
		}
		node, e := build(t.Resources[uri], uri)
		if e != nil {
			return nil, nil, e
		}
		if e = idx.register(uri, node); e != nil {
			return nil, nil, e
		}
	}
	for _, n := range idx.nodes {
		for _, k := range []string{"$ref", "$dynamicRef"} {
			if ref, ok := n.object[k].(string); ok {
				target, e := idx.resolve(n, ref)
				if e != nil {
					return nil, nil, e
				}
				n.refs[k] = target
				if target == nil {
					idx.limitations["unresolved_ref"] = true
				}
			}
		}
	}
	return idx, root, nil
}

var schemaMapKeywords = []string{"$defs", "definitions", "properties", "patternProperties", "dependentSchemas"}
var schemaSingleKeywords = []string{"additionalProperties", "unevaluatedProperties", "propertyNames", "items", "contains", "additionalItems", "unevaluatedItems", "not", "if", "then", "else", "contentSchema"}
var schemaArrayKeywords = []string{"allOf", "anyOf", "oneOf", "prefixItems"}

func (idx *schemaIndex) build(v any, d *schemaDocument, ptr, base, uri, rptr string) (*schemaNode, error) {
	n := &schemaNode{value: v, doc: d, pointer: ptr, base: base, uri: uri, resourcePointer: rptr, children: map[string]*schemaNode{}, patterns: map[string]schemaPattern{}, refs: map[string]*schemaNode{}}
	if _, ok := v.(bool); !ok {
		var yes bool
		n.object, yes = v.(map[string]any)
		if !yes {
			return nil, inputError("schema at %s must be object or boolean", ptr)
		}
	}
	d.nodes[ptr] = n
	idx.nodes = append(idx.nodes, n)
	if dialect, ok := n.object["$schema"]; ok {
		if dialect != jsonSchemaDialect && dialect != jsonSchemaDialect+"#" {
			return nil, inputError("unsupported schema dialect at %s", ptr)
		}
	}
	if id, ok := n.object["$id"]; ok {
		s, ok := id.(string)
		if !ok {
			return nil, inputError("$id must be string")
		}
		u, e := parseSchemaURI(s)
		if e != nil {
			return nil, e
		}
		if u.Fragment != "" {
			return nil, inputError("$id cannot contain nonempty fragment")
		}
		resolved, e := resolveURI(base, s)
		if e != nil {
			return nil, e
		}
		if resolved == "" {
			n.base = ""
		} else {
			n.base = resolved
			n.uri = resolved
			n.resourcePointer = ""
			if e = idx.register(resolved, n); e != nil {
				return nil, e
			}
		}
	}
	for _, key := range []string{"$anchor", "$dynamicAnchor"} {
		if value, ok := n.object[key]; ok {
			s, ok := value.(string)
			if !ok || !validAnchor(s) {
				return nil, inputError("invalid %s", key)
			}
			identity := n.uri + "#" + s
			if idx.anchors[identity] != nil {
				return nil, inputError("duplicate anchor %q", s)
			}
			idx.anchors[identity] = n
			if key == "$dynamicAnchor" {
				idx.limitations["dynamic_ref"] = true
			}
		}
	}
	for _, key := range []string{"$ref", "$dynamicRef"} {
		if v, ok := n.object[key]; ok {
			s, ok := v.(string)
			if !ok {
				return nil, inputError("%s must be string", key)
			}
			if _, e := parseSchemaURI(s); e != nil {
				return nil, e
			}
			if key == "$dynamicRef" {
				idx.limitations["dynamic_ref"] = true
			}
		}
	}
	if e := validateSchemaShapes(n.object); e != nil {
		return nil, inputError("schema %s: %v", ptr, e)
	}
	for _, k := range []string{"unevaluatedProperties", "not", "dependentSchemas", "dependentRequired", "propertyNames", "minProperties", "maxProperties", "const", "enum"} {
		if _, ok := n.object[k]; ok {
			idx.limitations[capabilityForKeyword(k)] = true
		}
	}
	if _, ok := n.object["if"]; ok {
		idx.limitations["conditional_schema"] = true
	}
	if vocab, ok := n.object["$vocabulary"].(map[string]any); ok {
		for uri, v := range vocab {
			if v == true && !knownSchemaVocabulary(uri) {
				idx.limitations["required_vocabulary"] = true
			}
		}
	}
	child := func(key string, v any) error {
		c, e := idx.build(v, d, ptr+"/"+key, n.base, n.uri, n.resourcePointer+"/"+key)
		if e == nil {
			n.children[key] = c
		}
		return e
	}
	for _, k := range schemaMapKeywords {
		if raw, ok := n.object[k]; ok {
			m, ok := raw.(map[string]any)
			if !ok {
				return nil, inputError("%s must be object", k)
			}
			for _, name := range sortedKeys(m) {
				if k == "patternProperties" {
					p, e := compileSchemaPattern(name)
					if e != nil {
						return nil, e
					}
					n.patterns[name] = p
					idx.limitations["pattern_properties"] = true
					if p.re == nil {
						idx.limitations["unsupported_pattern"] = true
					}
				}
				if e := child(k+"/"+pointerEscape(name), m[name]); e != nil {
					return nil, e
				}
			}
		}
	}
	for _, k := range schemaSingleKeywords {
		if v, ok := n.object[k]; ok {
			if e := child(k, v); e != nil {
				return nil, e
			}
		}
	}
	for _, k := range schemaArrayKeywords {
		if v, ok := n.object[k]; ok {
			a, ok := v.([]any)
			if !ok || ((k == "allOf" || k == "anyOf" || k == "oneOf") && len(a) == 0) {
				return nil, inputError("%s must be nonempty schema array", k)
			}
			for i, v := range a {
				if e := child(k+"/"+strconv.Itoa(i), v); e != nil {
					return nil, e
				}
			}
		}
	}
	return n, nil
}
func validAnchor(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			continue
		}
		if i > 0 && ((c >= '0' && c <= '9') || c == '-' || c == '.') {
			continue
		}
		return false
	}
	return true
}
func validateSchemaShapes(m map[string]any) error {
	if v, ok := m["type"]; ok {
		a, ok := v.([]any)
		if !ok {
			a = []any{v}
		}
		if len(a) == 0 {
			return inputError("empty type")
		}
		seen := map[string]bool{}
		for _, v := range a {
			s, ok := v.(string)
			if !ok || seen[s] || !strings.Contains("|null|boolean|object|array|number|string|integer|", "|"+s+"|") {
				return inputError("invalid type")
			}
			seen[s] = true
		}
	}
	checkNames := func(v any) error {
		a, ok := v.([]any)
		if !ok {
			return inputError("expected string array")
		}
		seen := map[string]bool{}
		for _, v := range a {
			s, ok := v.(string)
			if !ok || seen[s] {
				return inputError("invalid or duplicate string")
			}
			seen[s] = true
		}
		return nil
	}
	if v, ok := m["required"]; ok {
		if e := checkNames(v); e != nil {
			return e
		}
	}
	if v, ok := m["dependentRequired"]; ok {
		d, ok := v.(map[string]any)
		if !ok {
			return inputError("dependentRequired must be object")
		}
		for _, v := range d {
			if e := checkNames(v); e != nil {
				return e
			}
		}
	}
	for _, k := range []string{"minProperties", "maxProperties", "minItems", "maxItems", "minLength", "maxLength", "minContains", "maxContains"} {
		if v, ok := m[k]; ok {
			num, ok := v.(json.Number)
			if !ok {
				return inputError("%s must be nonnegative integer", k)
			}
			f, ok := new(big.Rat).SetString(string(num))
			if !ok || f.Sign() < 0 || !f.IsInt() {
				return inputError("%s must be nonnegative integer", k)
			}
		}
	}
	for _, k := range []string{"uniqueItems", "readOnly", "writeOnly", "deprecated"} {
		if v, ok := m[k]; ok {
			if _, ok := v.(bool); !ok {
				return inputError("%s must be boolean", k)
			}
		}
	}
	for _, k := range []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf"} {
		if v, ok := m[k]; ok {
			num, ok := v.(json.Number)
			if !ok {
				return inputError("%s must be number", k)
			}
			if k == "multipleOf" {
				r, ok := new(big.Rat).SetString(string(num))
				if !ok || r.Sign() <= 0 {
					return inputError("multipleOf must be positive")
				}
			}
		}
	}
	if v, ok := m["enum"]; ok {
		a, ok := v.([]any)
		if !ok || len(a) == 0 {
			return inputError("enum must be nonempty array")
		}
	}
	if v, ok := m["$vocabulary"]; ok {
		d, ok := v.(map[string]any)
		if !ok {
			return inputError("$vocabulary must be object")
		}
		for k, v := range d {
			u, e := parseSchemaURI(k)
			if e != nil || !u.IsAbs() {
				return inputError("invalid vocabulary URI")
			}
			if _, ok := v.(bool); !ok {
				return inputError("vocabulary requirement must be boolean")
			}
		}
	}
	for _, k := range []string{"pattern", "format", "title", "description", "$comment"} {
		if v, ok := m[k]; ok {
			if _, ok := v.(string); !ok {
				return inputError("%s must be string", k)
			}
		}
	}
	return nil
}
func (idx *schemaIndex) resolve(n *schemaNode, ref string) (*schemaNode, error) {
	resolved, e := resolveURI(n.base, ref)
	if e != nil || resolved == "" {
		return nil, e
	}
	u, e := parseSchemaURI(resolved)
	if e != nil {
		return nil, e
	}
	fragment := u.Fragment
	u.Fragment = ""
	u.RawFragment = ""
	uri := u.String()
	root := idx.resources[uri]
	if root == nil {
		return nil, nil
	}
	if fragment == "" {
		return root, nil
	}
	if !strings.HasPrefix(fragment, "/") {
		return idx.anchors[uri+"#"+fragment], nil
	}
	parts, e := pointerFragment(fragment)
	if e != nil {
		return nil, e
	}
	v := root.value
	ptr := root.pointer
	for _, part := range parts {
		ptr += "/" + pointerEscape(part)
		switch m := v.(type) {
		case map[string]any:
			var ok bool
			v, ok = m[part]
			if !ok {
				return nil, nil
			}
		case []any:
			i, e := strconv.Atoi(part)
			if e != nil || i < 0 || i >= len(m) || strconv.Itoa(i) != part {
				return nil, nil
			}
			v = m[i]
		default:
			return nil, nil
		}
	}
	target := root.doc.nodes[ptr]
	if target == nil {
		return nil, inputError("reference points to non-schema value at %s", ptr)
	}
	return target, nil
}

func knownSchemaVocabulary(uri string) bool {
	switch uri {
	case "https://json-schema.org/draft/2020-12/vocab/core", "https://json-schema.org/draft/2020-12/vocab/applicator", "https://json-schema.org/draft/2020-12/vocab/unevaluated", "https://json-schema.org/draft/2020-12/vocab/validation", "https://json-schema.org/draft/2020-12/vocab/meta-data", "https://json-schema.org/draft/2020-12/vocab/format-annotation", "https://json-schema.org/draft/2020-12/vocab/format-assertion", "https://json-schema.org/draft/2020-12/vocab/content":
		return true
	}
	return false
}
