package validation

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// These are normal compile_version 1 tables, not raw OCSF inheritance inputs.
// Descriptive/browser metadata is intentionally ignored. Every structural link
// is checked before the immutable target is exposed.
type ocsfAttribute struct {
	Type        string   `json:"type"`
	ObjectType  string   `json:"object_type"`
	Requirement string   `json:"requirement"`
	Profiles    []string `json:"profiles"`
	IsArray     bool     `json:"is_array"`
}

func (a *ocsfAttribute) UnmarshalJSON(raw []byte) error {
	type plain ocsfAttribute
	value := struct {
		*plain
		Array json.RawMessage `json:"is_array"`
	}{plain: (*plain)(a)}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	if len(value.Array) > 0 {
		if string(value.Array) != "true" && string(value.Array) != "false" {
			return inputError("compiled is_array must be boolean")
		}
		a.IsArray = string(value.Array) == "true"
	}
	return nil
}

type ocsfNode struct {
	Name         string                     `json:"name"`
	UID          *int64                     `json:"uid"`
	Category     string                     `json:"category"`
	CategoryUID  *int64                     `json:"category_uid"`
	Profiles     []string                   `json:"profiles"`
	Attributes   map[string]ocsfAttribute   `json:"attributes"`
	Constraints  map[string]json.RawMessage `json:"constraints"`
	Extends      json.RawMessage            `json:"extends"`
	Hidden       bool                       `json:"hidden"`
	HiddenFlag   bool                       `json:"hidden?"`
	Abstract     bool                       `json:"abstract"`
	AbstractFlag bool                       `json:"abstract?"`
}

func (n ocsfNode) concrete() bool {
	return !n.Hidden && !n.HiddenFlag && !n.Abstract && !n.AbstractFlag
}

type ocsfExtension struct {
	Name    string `json:"name"`
	UID     *int64 `json:"uid"`
	Version string `json:"version"`
}
type ocsfCatalog struct {
	CompileVersion int                 `json:"compile_version"`
	Version        string              `json:"version"`
	Classes        map[string]ocsfNode `json:"classes"`
	Objects        map[string]ocsfNode `json:"objects"`
	Profiles       map[string]ocsfNode `json:"profiles"`
	Categories     struct {
		Attributes map[string]ocsfNode `json:"attributes"`
	} `json:"categories"`
	Extensions map[string]ocsfExtension `json:"extensions"`
}
type ocsfTarget struct {
	catalog         ocsfCatalog
	selection       *OCSFSelection
	members         []SchemaClass
	metadata        SchemaTargetInfo
	fields          []string
	complete        bool
	unknownProfiles map[string]bool
}

func prepareOCSF(t SchemaTarget) (preparedSchemaTarget, error) {
	t, e := normalizeSchemaTarget(t)
	if e != nil {
		return nil, e
	}
	if t.Kind != "ocsf" {
		return nil, inputError("expected ocsf target")
	}
	p := &ocsfTarget{selection: t.Selection, unknownProfiles: map[string]bool{}}
	if e = json.Unmarshal(t.Catalog, &p.catalog); e != nil {
		return nil, inputError("invalid compiled OCSF catalog: %v", e)
	}
	c := &p.catalog
	if c.CompileVersion != 1 || !validSchemaName(c.Version) || c.Version != t.Selection.Version {
		return nil, inputError("OCSF requires compile_version 1 and the exact selected version")
	}
	if c.Classes == nil || c.Objects == nil || c.Profiles == nil || c.Categories.Attributes == nil || c.Extensions == nil {
		return nil, inputError("OCSF requires compiled class, object, profile, category and extension tables")
	}
	if e = p.validateCatalog(); e != nil {
		return nil, e
	}
	extensionKeys := sortedKeys(c.Extensions)
	if !equalOCSFNames(extensionKeys, t.Selection.Extensions) {
		return nil, inputError("selection extensions must equal the entire compiled extension set")
	}
	s := p.selection
	if s.Class != "" || s.ClassUID != nil {
		key := s.Class
		if s.ClassUID != nil {
			for k, n := range c.Classes {
				if *n.UID == *s.ClassUID {
					key = k
					break
				}
			}
		}
		n, ok := c.Classes[key]
		if !ok || key == "base_event" || !n.concrete() {
			return nil, inputError("selection requires a concrete OCSF class")
		}
		p.members = []SchemaClass{{Key: key, UID: *n.UID}}
		s.Class = key
		s.ClassUID = nil
	} else {
		key := s.Category
		if s.CategoryUID != nil {
			for k, n := range c.Categories.Attributes {
				if *n.UID == *s.CategoryUID {
					key = k
					break
				}
			}
		}
		if _, ok := c.Categories.Attributes[key]; !ok {
			return nil, inputError("unknown OCSF category")
		}
		p.members = []SchemaClass{}
		for _, k := range sortedKeys(c.Classes) {
			n := c.Classes[k]
			if k != "base_event" && n.concrete() && n.Category == key {
				p.members = append(p.members, SchemaClass{Key: k, UID: *n.UID})
			}
		}
		if len(p.members) == 0 {
			return nil, inputError("selection requires a nonempty concrete category")
		}
		s.Category = key
		s.CategoryUID = nil
	}
	for _, profile := range s.Profiles {
		n, ok := c.Profiles[profile]
		if !ok {
			return nil, inputError("unknown OCSF profile %q", profile)
		}
		applicable := false
		for _, m := range p.members {
			if containsOCSF(c.Classes[m.Key].Profiles, profile) {
				applicable = true
				if ocsfHasInheritance(n.Extends) {
					p.unknownProfiles[m.Key] = true
				}
			}
		}
		if !applicable {
			return nil, inputError("OCSF profile %q is inapplicable to the selected members", profile)
		}
	}
	p.metadata = SchemaTargetInfo{Kind: "ocsf", Identity: t.Identity, Version: c.Version, CompileVersion: 1, ResourceURIs: []string{}, Members: p.members, Extensions: []SchemaExtension{}, Limitations: []string{}}
	for _, k := range extensionKeys {
		n := c.Extensions[k]
		p.metadata.Extensions = append(p.metadata.Extensions, SchemaExtension{Key: k, UID: *n.UID, Version: n.Version})
	}
	p.enumerate()
	return p, nil
}
func equalOCSFNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func containsOCSF(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}
func ocsfHasInheritance(raw json.RawMessage) bool {
	return len(raw) > 0 && string(raw) != "null" && string(raw) != "[]" && string(raw) != "\"\""
}
func ocsfIdentity(key, name string) bool {
	if !validSchemaName(key) || !validSchemaName(name) {
		return false
	}
	return key == name || strings.HasSuffix(key, "/"+name)
}
func (p *ocsfTarget) validIdentity(key, name string) bool {
	if !ocsfIdentity(key, name) {
		return false
	}
	if prefix, _, namespaced := strings.Cut(key, "/"); namespaced {
		_, ok := p.catalog.Extensions[prefix]
		return ok
	}
	return true
}
func (p *ocsfTarget) validateProfiles(names []string) error {
	seen := map[string]bool{}
	for _, n := range names {
		if !validSchemaName(n) || seen[n] {
			return inputError("invalid or duplicate compiled profile dependency")
		}
		if _, ok := p.catalog.Profiles[n]; !ok {
			return inputError("unknown compiled profile dependency %q", n)
		}
		seen[n] = true
	}
	return nil
}
func (p *ocsfTarget) validateCatalog() error {
	c := &p.catalog
	for _, table := range []map[string]ocsfNode{c.Classes, c.Categories.Attributes} {
		ids := map[int64]bool{}
		for _, key := range sortedKeys(table) {
			n := table[key]
			if !validSchemaName(key) || n.UID == nil || *n.UID < 0 || ids[*n.UID] {
				return inputError("invalid or ambiguous OCSF UID at %q", key)
			}
			ids[*n.UID] = true
		}
	}
	extensionIDs := map[int64]bool{}
	for _, key := range sortedKeys(c.Extensions) {
		n := c.Extensions[key]
		if !ocsfIdentity(key, n.Name) || n.UID == nil || *n.UID <= 0 || !validSchemaName(n.Version) || extensionIDs[*n.UID] {
			return inputError("invalid OCSF extension identity %q", key)
		}
		extensionIDs[*n.UID] = true
	}
	for _, key := range sortedKeys(c.Profiles) {
		n := c.Profiles[key]
		if !p.validIdentity(key, n.Name) {
			return inputError("invalid profile identity %q", key)
		}
	}
	for _, key := range sortedKeys(c.Categories.Attributes) {
		if *c.Categories.Attributes[key].UID <= 0 {
			return inputError("nonconcrete category identity %q", key)
		}
	}
	for _, table := range []map[string]ocsfNode{c.Classes, c.Objects} {
		for _, key := range sortedKeys(table) {
			n := table[key]
			if !p.validIdentity(key, n.Name) || n.Attributes == nil {
				return inputError("invalid compiled node %q", key)
			}
			if e := p.validateProfiles(n.Profiles); e != nil {
				return e
			}
			for _, name := range sortedKeys(n.Attributes) {
				a := n.Attributes[name]
				if !validSchemaName(name) || !validSchemaName(a.Type) || (a.Requirement != "required" && a.Requirement != "optional" && a.Requirement != "recommended") {
					return inputError("invalid compiled attribute %q/%q", key, name)
				}
				if e := p.validateProfiles(a.Profiles); e != nil {
					return e
				}
				if a.Type == "object_t" {
					if _, ok := c.Objects[a.ObjectType]; !ok {
						return inputError("dangling local object_type %q", a.ObjectType)
					}
				} else if a.ObjectType != "" {
					return inputError("object_type on nonobject attribute %q/%q", key, name)
				}
			}
		}
	}
	for _, key := range sortedKeys(c.Classes) {
		n := c.Classes[key]
		if key == "base_event" {
			if n.Name != "base_event" || *n.UID != 0 || n.Category != "other" || n.CategoryUID == nil || *n.CategoryUID != 0 {
				return inputError("malformed base_event sentinel")
			}
			continue
		}
		category, ok := c.Categories.Attributes[n.Category]
		if *n.UID <= 0 || !ok || n.CategoryUID == nil || *n.CategoryUID != *category.UID {
			return inputError("invalid concrete category link for %q", key)
		}
	}
	return nil
}
func (p *ocsfTarget) universe() analysis.SourceUniverse {
	return analysis.SourceUniverse{Fields: append([]string{}, p.fields...), Complete: p.complete, Resolve: func(n string) analysis.SourceFieldAdmission { return p.project(n).Admission }}
}
func (p *ocsfTarget) info() SchemaTargetInfo {
	v := p.metadata
	s := *p.selection
	s.Profiles = append([]string{}, s.Profiles...)
	s.Extensions = append([]string{}, s.Extensions...)
	v.Selection = &s
	v.Members = append([]SchemaClass{}, v.Members...)
	v.Extensions = append([]SchemaExtension{}, v.Extensions...)
	v.ResourceURIs = append([]string{}, v.ResourceURIs...)
	v.Limitations = append([]string{}, v.Limitations...)
	return v
}

// Enumeration is bounded and conservative. Generic objects, arrays, recursion,
// and unresolved profile inheritance always leave a partial name universe.
func (p *ocsfTarget) enumerate() {
	names := map[string]bool{}
	limits := map[string]bool{"array_traversal": true}
	p.complete = true
	work := 0
	var walk func(ocsfNode, string, map[string]bool, int, string)
	walk = func(n ocsfNode, prefix string, active map[string]bool, depth int, class string) {
		for _, raw := range n.Constraints {
			var names []string
			if json.Unmarshal(raw, &names) != nil || len(names) == 0 {
				p.complete = false
				limits["ocsf_constraint"] = true
			}
		}
		for operator := range n.Constraints {
			if operator != "at_least_one" && operator != "just_one" {
				p.complete = false
				limits["ocsf_constraint"] = true
			}
		}
		for _, key := range sortedKeys(n.Attributes) {
			work++
			if work > schemaEnumerationBudget || depth > schemaPathSegmentBudget {
				p.complete = false
				limits["enumeration_budget"] = true
				return
			}
			a := n.Attributes[key]
			admitted, uncertain, _ := p.profileAdmission(a, class)
			if uncertain {
				limits["ocsf_profile_requirement"] = true
			}
			if admitted == analysis.SourceFieldProhibited {
				continue
			}
			name := prefix + key
			names[name] = true
			if a.IsArray {
				p.complete = false
				continue
			}
			if a.Type == "json_t" || (a.Type == "object_t" && a.ObjectType == "object") {
				p.complete = false
				continue
			}
			if a.Type == "object_t" {
				if active[a.ObjectType] {
					p.complete = false
					limits["recursive_ref"] = true
					continue
				}
				active[a.ObjectType] = true
				walk(p.catalog.Objects[a.ObjectType], name+".", active, depth+1, class)
				delete(active, a.ObjectType)
			}
		}
	}
	for _, m := range p.members {
		if p.unknownProfiles[m.Key] {
			p.complete = false
			limits["ocsf_profile_inheritance"] = true
		}
		walk(p.catalog.Classes[m.Key], "", map[string]bool{}, 0, m.Key)
	}
	p.fields = []string{}
	for _, name := range sortedKeys(names) {
		if p.project(name).Admission != analysis.SourceFieldProhibited {
			p.fields = append(p.fields, name)
		}
	}
	if !p.complete {
		limits["partial_name_universe"] = true
	}
	// Capability metadata describes constructs present, even if a query never uses them.
	for _, table := range []map[string]ocsfNode{p.catalog.Classes, p.catalog.Objects} {
		for _, n := range table {
			for key := range n.Constraints {
				if key != "at_least_one" && key != "just_one" {
					limits["ocsf_constraint"] = true
				}
			}
		}
	}
	p.metadata.Limitations = sortedKeys(limits)
	sort.Strings(p.fields)
}
