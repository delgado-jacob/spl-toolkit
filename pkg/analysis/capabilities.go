package analysis

import (
	"fmt"
	"sort"
)

// Capabilities returns fresh values generated from the dispatch registries.
func Capabilities() CapabilityManifest {
	m := CapabilityManifest{Rewrite: rewriteCapabilities("spl"), SchemaVersion: 1, Language: "spl", Profile: "splunkd", Version: "current", Commands: []Capability{}, Functions: []Capability{}}
	for name, c := range commands {
		m.Commands = append(m.Commands, Capability{Name: name, SyntaxSupported: true, SemanticSupported: c.handle != nil, Limitations: []string{c.limitation}})
	}
	for name, f := range functions {
		limit := fmt.Sprintf("%d to %d arguments", f.min, f.max)
		if f.max < 0 {
			limit = fmt.Sprintf("at least %d arguments", f.min)
		}
		if name == "case" {
			limit += " in condition/value pairs"
		}
		if f.aggregate {
			limit += "; aggregate context only"
		} else {
			limit += "; expression context only"
		}
		if f.dynamic {
			limit = "Dynamic query semantics are unresolved."
		}
		m.Functions = append(m.Functions, Capability{Name: name, SyntaxSupported: true, SemanticSupported: !f.dynamic, Limitations: []string{limit}})
	}
	sort.Slice(m.Commands, func(i, j int) bool { return m.Commands[i].Name < m.Commands[j].Name })
	sort.Slice(m.Functions, func(i, j int) bool { return m.Functions[i].Name < m.Functions[j].Name })
	return m
}

// CapabilitiesFor validates selectors and returns a fresh available manifest.
func CapabilitiesFor(options CapabilityOptions) (CapabilityManifest, error) {
	normalized, err := normalizeSelectors(options)
	if err != nil {
		return CapabilityManifest{}, err
	}
	if normalized.Language == "spl2" {
		return spl2Capabilities(), nil
	}
	return Capabilities(), nil
}

// These are canonical render roles, not promises of runtime equivalence. Binding,
// linked effects and whole-candidate correspondence still govern each edit.
func rewriteCapabilities(language string) *RewriteCapabilityManifest {
	m := &RewriteCapabilityManifest{SchemaVersion: 1, Forms: []RewriteCapabilityForm{}}
	add := func(kind, role string, supported bool, limitations ...string) {
		m.Forms = append(m.Forms, RewriteCapabilityForm{Kind: kind, Role: role, IdentityForms: []string{"atom"}, Supported: supported, Limitations: append([]string{}, limitations...)})
	}
	for _, role := range []string{"expression_atom", "search_field", "selector_atom", "rename_input", "null_test", "lookup_local", "lookup_dual"} {
		add("field", role, true, "Requires an exact source binding; explicit aliases, wildcard selectors and indeterminate operands remain fixed.")
	}
	add("field", "implicit_output", language == "spl", "Only canonical implicit labels and all required consumers may change together; SPL2 changed labels remain unproved.")
	m.Forms = append(m.Forms, RewriteCapabilityForm{Kind: "field", Role: "navigation", IdentityForms: []string{"path"}, Supported: false, Limitations: []string{"Navigated and alias-qualified identities have no canonical source binding proof."}})
	for _, kind := range []string{"index", "source", "sourcetype"} {
		add(kind, "search_value", true, "Typed exact static search values only; patterns, interpolation and uncertain decoding remain held.")
	}
	for _, kind := range []string{"lookup", "dataset"} {
		add(kind, "catalog_atom", true, "Named external dependency, independent of columns and source aliases.")
	}
	if language == "spl2" {
		add("index", "metric_value", true, "Only intact command-owned bare-index equality with a proven static value.")
		add("data_model", "quoted_catalog_atom", true, "Existing atomic tstats datamodel_name token; affected effects remain subject to proof.")
	} else {
		add("data_model", "catalog_atom", true, "Exact model token.")
		add("data_model", "catalog_component", true, "Original qualified owner, prefix and quote boundaries are retained.")
		add("dataset", "catalog_component", true, "Dataset component keeps its model prefix; linked model normalization is explicit.")
		add("dataset", "qualified_dataset", true, "Compatible overlapping proposals co-render; conflicts and backslashes remain refused.")
	}
	return m
}
