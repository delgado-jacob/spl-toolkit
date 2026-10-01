package environment

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func invalidSchemaBundle(path string, err error) (*PreparedSchemaBundle, *Report, error) {
	var offset *int
	if path == "" {
		path, offset = inputLocation(err)
	}
	return nil, &Report{SchemaVersion: 1, Status: "invalid", Coverage: []CoverageEntry{}, Diagnostics: []Diagnostic{{Code: "schema_bundle_invalid", Severity: "error", Artifact: "schema_bundle", Path: path, ByteOffset: offset, Message: err.Error()}}}, nil
}

func normalizeBundleProvenance(p Provenance, path string) (Provenance, error) {
	if err := nonblank(p.SourceKind, "source_kind"); err != nil {
		return Provenance{}, at(path+"/source_kind", err)
	}
	if err := nonblank(p.SourceID, "source_id"); err != nil {
		return Provenance{}, at(path+"/source_id", err)
	}
	observed, _, err := normalizeTime(p.ObservedAt, "observed_at")
	if err != nil {
		return Provenance{}, at(path+"/observed_at", err)
	}
	p.ObservedAt = observed
	return p, nil
}

func canonicalBundleJSON(raw []byte) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func normalizeSchemaBundle(input SchemaBundle) (SchemaBundle, map[string]preparedBundleTarget, []CoverageEntry, []Diagnostic, error) {
	if !validateUTF8(reflect.ValueOf(input)) {
		return SchemaBundle{}, nil, nil, nil, fmt.Errorf("schema bundle contains invalid UTF-8")
	}
	if input.SchemaVersion != 1 {
		return SchemaBundle{}, nil, nil, nil, at("/schema_version", fmt.Errorf("schema_version must be integer 1"))
	}
	if err := nonblank(input.BundleID, "bundle_id"); err != nil {
		return SchemaBundle{}, nil, nil, nil, at("/bundle_id", err)
	}
	provenance, err := normalizeBundleProvenance(input.Provenance, "/provenance")
	if err != nil {
		return SchemaBundle{}, nil, nil, nil, err
	}
	out := SchemaBundle{SchemaVersion: 1, BundleID: input.BundleID, Provenance: provenance, Schemas: []SchemaEntry{}, Bindings: []SchemaBinding{}}
	targets := make(map[string]preparedBundleTarget, len(input.Schemas))
	for i, entry := range input.Schemas {
		path := fmt.Sprintf("/schemas/%d", i)
		if err := nonblank(entry.ID, "schema id"); err != nil {
			return SchemaBundle{}, nil, nil, nil, at(path+"/id", err)
		}
		if _, found := targets[entry.ID]; found {
			return SchemaBundle{}, nil, nil, nil, at(path+"/id", fmt.Errorf("duplicate schema id %q", entry.ID))
		}
		entry.Provenance, err = normalizeBundleProvenance(entry.Provenance, path+"/provenance")
		if err != nil {
			return SchemaBundle{}, nil, nil, nil, err
		}
		switch entry.Kind {
		case "field_list":
			if entry.Catalog == nil || entry.Target != nil {
				return SchemaBundle{}, nil, nil, nil, at(path, fmt.Errorf("field_list requires catalog only"))
			}
			catalog, err := validation.DecodeFieldCatalog(entry.Catalog)
			if err != nil {
				return SchemaBundle{}, nil, nil, nil, at(path+"/catalog", err)
			}
			prepared, err := validation.PrepareFieldCatalog(catalog)
			if err != nil {
				return SchemaBundle{}, nil, nil, nil, at(path+"/catalog", err)
			}
			entry.Catalog, err = json.Marshal(catalog)
			if err != nil {
				return SchemaBundle{}, nil, nil, nil, err
			}
			targets[entry.ID] = preparedBundleTarget{field: prepared}
		case "json_schema", "ocsf":
			if entry.Target == nil || entry.Catalog != nil {
				return SchemaBundle{}, nil, nil, nil, at(path, fmt.Errorf("%s requires target only", entry.Kind))
			}
			target, err := validation.DecodeSchemaTarget(entry.Target)
			if err != nil {
				return SchemaBundle{}, nil, nil, nil, at(path+"/target", err)
			}
			if target.Kind != entry.Kind {
				return SchemaBundle{}, nil, nil, nil, at(path+"/target/kind", fmt.Errorf("target kind %q differs from entry kind %q", target.Kind, entry.Kind))
			}
			prepared, err := validation.PrepareSchemaTarget(target)
			if err != nil {
				return SchemaBundle{}, nil, nil, nil, at(path+"/target", err)
			}
			raw, err := json.Marshal(target)
			if err != nil {
				return SchemaBundle{}, nil, nil, nil, err
			}
			entry.Target, err = canonicalBundleJSON(raw)
			if err != nil {
				return SchemaBundle{}, nil, nil, nil, err
			}
			targets[entry.ID] = preparedBundleTarget{schema: prepared}
		default:
			return SchemaBundle{}, nil, nil, nil, at(path+"/kind", fmt.Errorf("unsupported schema kind %q", entry.Kind))
		}
		out.Schemas = append(out.Schemas, entry)
	}
	sort.Slice(out.Schemas, func(i, j int) bool { return out.Schemas[i].ID < out.Schemas[j].ID })
	coverage := []CoverageEntry{}
	diagnostics := []Diagnostic{}
	seen := map[struct{ schemaID, objectID string }]bool{}
	for i, binding := range input.Bindings {
		path := fmt.Sprintf("/bindings/%d", i)
		if err := nonblank(binding.SchemaID, "schema_id"); err != nil {
			return SchemaBundle{}, nil, nil, nil, at(path+"/schema_id", err)
		}
		if _, found := targets[binding.SchemaID]; !found {
			return SchemaBundle{}, nil, nil, nil, at(path+"/schema_id", fmt.Errorf("unknown schema id %q", binding.SchemaID))
		}
		if err := nonblank(binding.ObjectID, "object_id"); err != nil {
			return SchemaBundle{}, nil, nil, nil, at(path+"/object_id", err)
		}
		key := struct{ schemaID, objectID string }{binding.SchemaID, binding.ObjectID}
		if seen[key] {
			return SchemaBundle{}, nil, nil, nil, at(path, fmt.Errorf("duplicate schema-object binding"))
		}
		seen[key] = true
		if !kindSet[binding.Expected.Kind] {
			return SchemaBundle{}, nil, nil, nil, at(path+"/expected/kind", fmt.Errorf("unsupported object kind %q", binding.Expected.Kind))
		}
		if err := nonblank(binding.Expected.Name, "expected name"); err != nil {
			return SchemaBundle{}, nil, nil, nil, at(path+"/expected/name", err)
		}
		if binding.Expected.Kind == "index" || binding.Expected.Kind == "source" || binding.Expected.Kind == "sourcetype" {
			if binding.Expected.Namespace != "" || binding.Expected.App != "" || binding.Expected.Owner != "" {
				return SchemaBundle{}, nil, nil, nil, at(path+"/expected", fmt.Errorf("inapplicable object context"))
			}
		}
		if binding.SourceCoverage != "complete" && binding.SourceCoverage != "partial" {
			return SchemaBundle{}, nil, nil, nil, at(path+"/source_coverage", fmt.Errorf("source_coverage must be complete or partial"))
		}
		if binding.SourceCoverage == "partial" {
			if err := nonblank(binding.Reason, "partial source reason"); err != nil {
				return SchemaBundle{}, nil, nil, nil, at(path+"/reason", err)
			}
		} else if binding.Reason != "" {
			return SchemaBundle{}, nil, nil, nil, at(path+"/reason", fmt.Errorf("complete source cannot have a reason"))
		}
		out.Bindings = append(out.Bindings, binding)
	}
	sort.Slice(out.Bindings, func(i, j int) bool {
		a, b := out.Bindings[i], out.Bindings[j]
		if a.SchemaID != b.SchemaID {
			return a.SchemaID < b.SchemaID
		}
		return a.ObjectID < b.ObjectID
	})
	for _, binding := range out.Bindings {
		coverage = append(coverage, CoverageEntry{Artifact: "schema_bundle", Kind: binding.Expected.Kind, SchemaID: binding.SchemaID, ObjectID: binding.ObjectID, Coverage: binding.SourceCoverage, Reason: binding.Reason})
		if binding.SourceCoverage == "partial" {
			diagnostics = append(diagnostics, Diagnostic{Code: "source_partial", Severity: "warning", Artifact: "schema_bundle", Path: "/bindings", Message: fmt.Sprintf("%s binding for %s is partial: %s", binding.SchemaID, binding.ObjectID, binding.Reason)})
		}
	}
	return out, targets, coverage, diagnostics, nil
}

// PrepareSchemaBundle validates a standalone bundle and compiles its targets.
func PrepareSchemaBundle(value SchemaBundle) (*PreparedSchemaBundle, *Report, error) {
	normalized, targets, coverage, diagnostics, err := normalizeSchemaBundle(value)
	if err != nil {
		return invalidSchemaBundle("", err)
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, nil, err
	}
	sum := sha256.Sum256(raw)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	if value.Digest != "" && value.Digest != digest {
		_, report, _ := invalidSchemaBundle("/digest", fmt.Errorf("asserted schema bundle digest does not match computed digest"))
		report.SchemaBundleDigest = digest
		return nil, report, nil
	}
	status := "valid"
	if len(diagnostics) != 0 {
		status = "partial"
	}
	report := &Report{SchemaVersion: 1, Status: status, SchemaBundleDigest: digest, Coverage: coverage, Diagnostics: diagnostics}
	stored := *report
	stored.Coverage = append([]CoverageEntry{}, coverage...)
	stored.Diagnostics = append([]Diagnostic{}, diagnostics...)
	return &PreparedSchemaBundle{bundle: normalized, report: stored, targets: targets}, report, nil
}

func (p *PreparedSchemaBundle) Bundle() SchemaBundle {
	if p == nil {
		return SchemaBundle{}
	}
	raw, _ := json.Marshal(p.bundle)
	var out SchemaBundle
	_ = json.Unmarshal(raw, &out)
	return out
}

func (p *PreparedSchemaBundle) Report() Report {
	if p == nil {
		return Report{}
	}
	out := p.report
	out.Coverage = append([]CoverageEntry{}, out.Coverage...)
	out.Diagnostics = append([]Diagnostic{}, out.Diagnostics...)
	return out
}

func validateSchemaBundleRawShape(raw []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return err
	}
	if root == nil {
		return fmt.Errorf("schema bundle must be an object")
	}
	for _, key := range []string{"schema_version", "bundle_id", "provenance", "schemas", "bindings"} {
		value, found := root[key]
		if !found {
			return at("/"+key, fmt.Errorf("missing property %q", key))
		}
		if bytes.Equal(value, []byte("null")) {
			return at("/"+key, fmt.Errorf("%s must not be null", key))
		}
	}
	for _, key := range []string{"schemas", "bindings"} {
		if len(root[key]) == 0 || root[key][0] != '[' {
			return at("/"+key, fmt.Errorf("%s must be an array", key))
		}
	}
	return nil
}
