package validation

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"unicode/utf8"
)

func DecodeSchemaTarget(data []byte) (SchemaTarget, error) {
	var t SchemaTarget
	f, err := object(data, []string{"kind"}, "kind", "identity", "schema", "base_uri", "resources", "catalog", "selection")
	if err != nil {
		return t, err
	}
	for _, p := range []struct {
		k string
		v *string
	}{{"kind", &t.Kind}, {"identity", &t.Identity}, {"base_uri", &t.BaseURI}} {
		if raw, ok := f[p.k]; ok {
			*p.v, err = stringValue(raw)
			if err != nil {
				return t, err
			}
			if !validSchemaName(*p.v) {
				return t, inputError("%s must be nonblank UTF-8", p.k)
			}
		}
	}
	switch t.Kind {
	case "json_schema":
		if _, ok := f["schema"]; !ok {
			return t, inputError("schema is required")
		}
		if _, ok := f["catalog"]; ok {
			return t, inputError("catalog is not allowed for json_schema")
		}
		if _, ok := f["selection"]; ok {
			return t, inputError("selection is not allowed for json_schema")
		}
		t.Schema = append(json.RawMessage(nil), f["schema"]...)
		if _, err = decodeUniqueJSON(t.Schema); err != nil {
			return t, err
		}
		if raw, ok := f["resources"]; ok {
			v, e := decodeUniqueJSON(raw)
			if e != nil {
				return t, e
			}
			m, ok := v.(map[string]any)
			if !ok {
				return t, inputError("resources must be an object")
			}
			t.Resources = make(map[string]json.RawMessage, len(m))
			for key, val := range m {
				if !validSchemaName(key) {
					return t, inputError("invalid resource identity")
				}
				t.Resources[key], _ = json.Marshal(val)
			}
		}
	case "ocsf":
		for _, key := range []string{"schema", "base_uri", "resources"} {
			if _, ok := f[key]; ok {
				return t, inputError("%s is not allowed for ocsf", key)
			}
		}
		if _, ok := f["catalog"]; !ok {
			return t, inputError("catalog is required")
		}
		t.Catalog = append(json.RawMessage(nil), f["catalog"]...)
		if _, err = decodeUniqueJSON(t.Catalog); err != nil {
			return t, err
		}
		t.Selection, err = decodeOCSFSelection(f["selection"])
		if err != nil {
			return t, err
		}
	default:
		return t, inputError("unsupported target kind %q", t.Kind)
	}
	return t, nil
}
func validSchemaName(s string) bool { return utf8.ValidString(s) && strings.TrimSpace(s) != "" }
func decodeOCSFSelection(raw []byte) (*OCSFSelection, error) {
	f, err := object(raw, []string{"version"}, "version", "class", "class_uid", "category", "category_uid", "profiles", "extensions")
	if err != nil {
		return nil, err
	}
	s := &OCSFSelection{Profiles: []string{}, Extensions: []string{}}
	count := 0
	for _, p := range []struct {
		k string
		v *string
	}{{"version", &s.Version}, {"class", &s.Class}, {"category", &s.Category}} {
		if v, ok := f[p.k]; ok {
			*p.v, err = stringValue(v)
			if err != nil || !validSchemaName(*p.v) {
				return nil, inputError("invalid selection %s", p.k)
			}
			if p.k != "version" {
				count++
			}
		}
	}
	for _, p := range []struct {
		k string
		v **int64
	}{{"class_uid", &s.ClassUID}, {"category_uid", &s.CategoryUID}} {
		if v, ok := f[p.k]; ok {
			var n int64
			if bytes.Equal(bytes.TrimSpace(v), []byte("null")) || json.Unmarshal(v, &n) != nil {
				return nil, inputError("%s must be int64", p.k)
			}
			*p.v = &n
			count++
		}
	}
	if count != 1 {
		return nil, inputError("selection requires exactly one class or category")
	}
	for _, p := range []struct {
		k string
		v *[]string
	}{{"profiles", &s.Profiles}, {"extensions", &s.Extensions}} {
		if v, ok := f[p.k]; ok {
			items, e := array(v)
			if e != nil {
				return nil, e
			}
			seen := map[string]bool{}
			for _, item := range items {
				name, e := stringValue(item)
				if e != nil || !validSchemaName(name) || seen[name] {
					return nil, inputError("invalid or duplicate %s name", p.k)
				}
				seen[name] = true
				*p.v = append(*p.v, name)
			}
			sort.Strings(*p.v)
		}
	}
	return s, nil
}

// Typed targets pass through the same presence and Unicode gates. Non-nil empty
// raw members and maps remain present so mixed unions cannot disappear via omitempty.
func normalizeSchemaTarget(t SchemaTarget) (SchemaTarget, error) {
	m := map[string]any{"kind": t.Kind}
	if t.Identity != "" {
		m["identity"] = t.Identity
	}
	if t.BaseURI != "" {
		m["base_uri"] = t.BaseURI
	}
	if t.Schema != nil {
		m["schema"] = t.Schema
	}
	if t.Catalog != nil {
		m["catalog"] = t.Catalog
	}
	if t.Resources != nil {
		m["resources"] = t.Resources
	}
	if t.Selection != nil {
		s := *t.Selection
		s.Profiles = append([]string{}, s.Profiles...)
		s.Extensions = append([]string{}, s.Extensions...)
		m["selection"] = &s
	}
	for _, s := range []string{t.Kind, t.Identity, t.BaseURI} {
		if !utf8.ValidString(s) {
			return t, inputError("invalid UTF-8 target string")
		}
	}
	if t.Selection != nil {
		s := t.Selection
		for _, v := range append(append([]string{s.Version, s.Class, s.Category}, s.Profiles...), s.Extensions...) {
			if !utf8.ValidString(v) {
				return t, inputError("invalid UTF-8 selection")
			}
		}
	}
	for key, raw := range t.Resources {
		if !utf8.ValidString(key) {
			return t, inputError("invalid UTF-8 resource identity")
		}
		if _, e := decodeUniqueJSON(raw); e != nil {
			return t, e
		}
	}
	for _, raw := range []json.RawMessage{t.Schema, t.Catalog} {
		if raw != nil {
			if _, e := decodeUniqueJSON(raw); e != nil {
				return t, e
			}
		}
	}
	raw, e := json.Marshal(m)
	if e != nil {
		return t, inputError("invalid target: %v", e)
	}
	return DecodeSchemaTarget(raw)
}
