package validation

import (
	"bytes"
	"sort"
	"strings"
	"unicode/utf8"
)

// DecodeFieldCatalog accepts a concrete string array or the strict catalog object.
func DecodeFieldCatalog(data []byte) (FieldCatalog, error) {
	var catalog FieldCatalog
	if err := checkJSON(data); err != nil {
		return catalog, err
	}
	data = bytes.TrimSpace(data)
	if data[0] == '[' {
		fields, err := decodeNames(data)
		if err != nil {
			return catalog, err
		}
		catalog.Fields = fields
	} else {
		values, err := object(data, []string{"fields"}, "fields", "optional_fields", "identity", "version")
		if err != nil {
			return catalog, err
		}
		catalog.Fields, err = decodeNames(values["fields"])
		if err != nil {
			return FieldCatalog{}, err
		}
		if raw, ok := values["optional_fields"]; ok {
			catalog.OptionalFields, err = decodeNames(raw)
			if err != nil {
				return FieldCatalog{}, err
			}
		}
		if raw, ok := values["identity"]; ok {
			catalog.Identity, err = stringValue(raw)
			if err != nil {
				return FieldCatalog{}, err
			}
		}
		if raw, ok := values["version"]; ok {
			catalog.Version, err = stringValue(raw)
			if err != nil {
				return FieldCatalog{}, err
			}
		}
	}
	return normalizeCatalog(catalog)
}
func decodeNames(data []byte) ([]string, error) {
	values, err := array(data)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(values))
	for _, raw := range values {
		name, err := stringValue(raw)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}
func normalizeCatalog(catalog FieldCatalog) (FieldCatalog, error) {
	catalog.Fields = append([]string{}, catalog.Fields...)
	catalog.OptionalFields = append([]string{}, catalog.OptionalFields...)
	seen := map[string]bool{}
	for _, names := range [][]string{catalog.Fields, catalog.OptionalFields} {
		for _, name := range names {
			if !utf8.ValidString(name) || strings.TrimSpace(name) == "" {
				return FieldCatalog{}, inputError("field name must be nonempty, non-whitespace UTF-8")
			}
			if seen[name] {
				return FieldCatalog{}, inputError("duplicate or overlapping field name %q", name)
			}
			seen[name] = true
		}
		sort.Strings(names)
	}
	if !utf8.ValidString(catalog.Identity) || !utf8.ValidString(catalog.Version) {
		return FieldCatalog{}, inputError("catalog metadata must be valid UTF-8")
	}
	return catalog, nil
}
