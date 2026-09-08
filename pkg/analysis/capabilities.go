package analysis

import (
	"fmt"
	"sort"
)

// Capabilities returns fresh values generated from the dispatch registries.
func Capabilities() CapabilityManifest {
	m := CapabilityManifest{SchemaVersion: 1, Language: "spl", Profile: "splunkd", Version: "current", Commands: []Capability{}, Functions: []Capability{}}
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
