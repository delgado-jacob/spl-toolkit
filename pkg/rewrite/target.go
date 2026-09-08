package rewrite

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// ValidationTarget selects one existing destination validation contract. Schema
// targets retain the existing flat SchemaTarget wire shape, without an envelope.
type ValidationTarget struct {
	Kind         string
	Catalog      *validation.FieldCatalog
	SchemaTarget *validation.SchemaTarget
}

// CandidateValidation retains one complete original validation report.
type CandidateValidation struct {
	Kind      string                   `json:"kind"`
	FieldList *validation.Report       `json:"field_list,omitempty"`
	Schema    *validation.SchemaReport `json:"schema,omitempty"`
}

func (c CandidateValidation) MarshalJSON() ([]byte, error) {
	switch c.Kind {
	case "field_list":
		if c.FieldList == nil || c.Schema != nil {
			return nil, inputError("field_list validation requires exactly one field-list report")
		}
	case "json_schema", "ocsf":
		if c.Schema == nil || c.FieldList != nil {
			return nil, inputError("schema validation requires exactly one schema report")
		}
	default:
		return nil, inputError("unsupported candidate validation kind %q", c.Kind)
	}
	type wire CandidateValidation
	return json.Marshal(wire(c))
}

func decodeTarget(raw []byte) (*ValidationTarget, error) {
	f, err := object(raw, []string{"kind"}, "kind", "catalog", "identity", "schema", "base_uri", "resources", "selection")
	if err != nil {
		return nil, err
	}
	kind, err := stringValue(f["kind"])
	if err != nil {
		return nil, err
	}
	if kind == "field_list" {
		f, err = object(raw, []string{"kind", "catalog"}, "kind", "catalog")
		if err != nil {
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

func (t ValidationTarget) MarshalJSON() ([]byte, error) {
	return targetJSON(t)
}

// targetJSON preserves non-nil empty members so invalid mixed unions cannot be
// erased by omitempty. It checks Go strings before encoding/json can replace
// malformed UTF-8, then delegates all target policy to the existing decoders.
func targetJSON(t ValidationTarget) ([]byte, error) {
	if t.Kind == "field_list" {
		if t.Catalog == nil || t.SchemaTarget != nil {
			return nil, inputError("field_list requires only a field catalog")
		}
		c := *t.Catalog
		for _, names := range [][]string{c.Fields, c.OptionalFields, {c.Identity, c.Version}} {
			for _, name := range names {
				if !utf8.ValidString(name) {
					return nil, inputError("invalid UTF-8 catalog string")
				}
			}
		}
		c.Fields = append([]string{}, c.Fields...)
		c.OptionalFields = append([]string{}, c.OptionalFields...)
		return json.Marshal(struct {
			Kind    string                  `json:"kind"`
			Catalog validation.FieldCatalog `json:"catalog"`
		}{t.Kind, c})
	}
	if (t.Kind != "json_schema" && t.Kind != "ocsf") || t.SchemaTarget == nil || t.Catalog != nil || t.Kind != t.SchemaTarget.Kind {
		return nil, inputError("schema target kind and payload must agree")
	}
	s := t.SchemaTarget
	for _, name := range []string{s.Kind, s.Identity, s.BaseURI} {
		if !utf8.ValidString(name) {
			return nil, inputError("invalid UTF-8 target string")
		}
	}
	m := map[string]any{"kind": s.Kind}
	if s.Identity != "" {
		m["identity"] = s.Identity
	}
	if s.BaseURI != "" {
		m["base_uri"] = s.BaseURI
	}
	for _, entry := range []struct {
		key string
		raw json.RawMessage
	}{{"schema", s.Schema}, {"catalog", s.Catalog}} {
		if entry.raw != nil {
			if err := checkJSON(entry.raw); err != nil {
				return nil, err
			}
			m[entry.key] = entry.raw
		}
	}
	if s.Resources != nil {
		for key, raw := range s.Resources {
			if !utf8.ValidString(key) {
				return nil, inputError("invalid UTF-8 resource identity")
			}
			if err := checkJSON(raw); err != nil {
				return nil, err
			}
		}
		m["resources"] = s.Resources
	}
	if s.Selection != nil {
		selection := *s.Selection
		for _, names := range [][]string{{selection.Version, selection.Class, selection.Category}, selection.Profiles, selection.Extensions} {
			for _, name := range names {
				if !utf8.ValidString(name) {
					return nil, inputError("invalid UTF-8 selection string")
				}
			}
		}
		selection.Profiles = append([]string{}, selection.Profiles...)
		selection.Extensions = append([]string{}, selection.Extensions...)
		m["selection"] = &selection
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, inputError("validation_target: %v", err)
	}
	return raw, nil
}

// Structural target preparation only: semantic schema compilation remains in
// canonical validation and must complete before any rewrite result is published.
func prepareTarget(t *ValidationTarget) (*ValidationTarget, error) {
	if t == nil {
		return nil, nil
	}
	raw, err := targetJSON(*t)
	if err != nil {
		return nil, err
	}
	return decodeTarget(raw)
}
