package analysis

import "sort"

type fieldValueState uint8

const (
	fieldValueUnknown fieldValueState = iota
	fieldValueElementSelected
)

type trackedField struct {
	FieldBinding
	identity   fieldIdentity
	owners     []sourceOwner
	source     bool
	valueState fieldValueState
}
type environment struct {
	inputs          []inputFact
	rewrite         *rewriteFlow
	fields          map[fieldIdentityKey]trackedField
	fieldOrder      []fieldIdentityKey
	identities      map[fieldIdentityKey]fieldIdentity
	removed         map[fieldIdentityKey]bool
	removedOrder    []fieldIdentityKey
	ambiguous       map[string]bool
	open, uncertain bool
	requirements    requirementEnvironment
}

func newEnvironment() *environment {
	return newEnvironmentWithRequirementTrace(nil)
}
func newEnvironmentWithRequirementTrace(trace *requirementTrace) *environment {
	return &environment{fields: map[fieldIdentityKey]trackedField{}, identities: map[fieldIdentityKey]fieldIdentity{}, removed: map[fieldIdentityKey]bool{}, ambiguous: map[string]bool{}, open: true, requirements: newRequirementEnvironment(trace)}
}
func copyIDs(ids []string) []string { return append([]string{}, ids...) }
func (e *environment) snapshot() FieldState {
	s := FieldState{Fields: []FieldBinding{}, Removed: []FieldRemoval{}, Open: e.open, Uncertain: e.uncertain}
	for _, key := range e.orderedFieldKeys() {
		f, exists := e.fields[key]
		if !exists {
			continue
		}
		f.OriginReferenceIDs = copyIDs(f.OriginReferenceIDs)
		f.FieldIdentity, _ = f.identity.public()
		s.Fields = append(s.Fields, f.FieldBinding)
	}
	for _, key := range e.orderedRemovedKeys() {
		if !e.removed[key] {
			continue
		}
		identity := e.identities[key]
		name := identity.PublicName
		if name == "" {
			continue
		}
		if _, live := e.fields[key]; live {
			continue
		}
		if publicIdentity, exact := identity.public(); exact {
			s.Removed = append(s.Removed, FieldRemoval{Name: name, FieldIdentity: publicIdentity})
		}
	}
	sort.Slice(s.Fields, func(i, j int) bool {
		if s.Fields[i].Name != s.Fields[j].Name {
			return s.Fields[i].Name < s.Fields[j].Name
		}
		return fieldIdentityLess(s.Fields[i].FieldIdentity, s.Fields[j].FieldIdentity)
	})
	sort.Slice(s.Removed, func(i, j int) bool {
		if s.Removed[i].Name != s.Removed[j].Name {
			return s.Removed[i].Name < s.Removed[j].Name
		}
		return fieldIdentityLess(s.Removed[i].FieldIdentity, s.Removed[j].FieldIdentity)
	})
	return s
}
func (e *environment) clone() *environment {
	return e.cloneWithRequirementTrace(e.requirements.trace)
}

func (e *environment) cloneWithRequirementTrace(trace *requirementTrace) *environment {
	n := newEnvironmentWithRequirementTrace(trace)
	n.inputs = cloneInputFacts(e.inputs)
	n.rewrite = e.rewrite.clone()
	n.open = e.open
	n.uncertain = e.uncertain
	n.requirements = e.requirements.clone()
	n.requirements.trace = trace
	for k, v := range e.fields {
		v.OriginReferenceIDs = copyIDs(v.OriginReferenceIDs)
		v.identity = v.identity.clone()
		v.owners = cloneSourceOwners(v.owners)
		n.fields[k] = v
	}
	n.fieldOrder = append([]fieldIdentityKey{}, e.fieldOrder...)
	for key, identity := range e.identities {
		n.identities[key] = identity.clone()
	}
	for k, v := range e.removed {
		n.removed[k] = v
	}
	n.removedOrder = append([]fieldIdentityKey{}, e.removedOrder...)
	for name, ambiguous := range e.ambiguous {
		n.ambiguous[name] = ambiguous
	}
	return n
}

func (e *environment) forkBranch() *environment {
	if e.requirements.trace == nil {
		return e.clone()
	}
	return e.cloneWithRequirementTrace(e.requirements.trace.forkBranch())
}
func (e *environment) remove(name string) {
	e.removeIdentity(atomicFieldIdentity(name))
}
func (e *environment) install(name string, ids []string, conditional bool) {
	e.installIdentity(atomicFieldIdentity(name), ids, conditional, false)
}

func (e *environment) installIdentity(identity fieldIdentity, ids []string, conditional, source bool) (bool, []string) {
	key, exact := identity.privateKey()
	if !exact {
		return false, nil
	}
	if !source {
		e.rewriteInvalidate(identity.PublicName)
	}
	collision, origins := e.registerIdentityField(trackedField{FieldBinding: FieldBinding{Name: identity.PublicName, OriginReferenceIDs: copyIDs(ids), Conditional: conditional}, identity: identity.clone(), owners: traceSourceOwners(e.requirements.trace, ids), source: source})
	delete(e.removed, key)
	return collision, origins
}

func (e *environment) registerIdentityField(field trackedField) (bool, []string) {
	key, exact := field.identity.privateKey()
	if !exact {
		field.identity = atomicFieldIdentity(field.Name)
		key, _ = field.identity.privateKey()
	}
	if _, exists := e.fields[key]; !exists {
		e.fieldOrder = append(e.fieldOrder, key)
	}
	field.OriginReferenceIDs = copyIDs(field.OriginReferenceIDs)
	field.identity = field.identity.clone()
	field.owners = cloneSourceOwners(field.owners)
	field.FieldIdentity, _ = field.identity.public()
	e.fields[key] = field
	e.identities[key] = field.identity.clone()
	return e.detectCollision(field.Name)
}

func (e *environment) removeIdentity(identity fieldIdentity) {
	key, exact := identity.privateKey()
	if !exact {
		return
	}
	e.rewriteInvalidate(identity.PublicName)
	e.identities[key] = identity.clone()
	delete(e.fields, key)
	if !e.removed[key] {
		e.removedOrder = append(e.removedOrder, key)
	}
	e.removed[key] = true
}

func (e *environment) field(identity fieldIdentity) (trackedField, bool) {
	key, exact := identity.privateKey()
	if !exact {
		return trackedField{}, false
	}
	field, known := e.fields[key]
	return field, known
}

// selectElement records a private value-shape fact without changing public
// field presence, provenance, identity, or requirement evidence.
func (e *environment) selectElement(identity fieldIdentity) bool {
	key, exact := identity.privateKey()
	if !exact {
		return false
	}
	field, known := e.fields[key]
	if !known {
		return false
	}
	field.valueState = fieldValueElementSelected
	e.fields[key] = field
	return true
}

func (e *environment) hasOtherIdentity(identity fieldIdentity) bool {
	key, exact := identity.privateKey()
	if !exact {
		return false
	}
	for candidateKey, field := range e.fields {
		if candidateKey != key && field.Name == identity.PublicName {
			return true
		}
	}
	return false
}

func (e *environment) detectCollision(publicName string) (bool, []string) {
	origins := []string{}
	for _, key := range e.orderedFieldKeys() {
		field, exists := e.fields[key]
		if !exists || field.Name != publicName {
			continue
		}
		origins = uniqueIDs(origins, field.OriginReferenceIDs)
	}
	return false, origins
}

func fieldIdentityLess(a, b FieldIdentity) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	if a.Qualifier != b.Qualifier {
		return a.Qualifier < b.Qualifier
	}
	for i := 0; i < len(a.Segments) && i < len(b.Segments); i++ {
		if a.Segments[i] != b.Segments[i] {
			return a.Segments[i] < b.Segments[i]
		}
	}
	return len(a.Segments) < len(b.Segments)
}

func (e *environment) orderedFieldKeys() []fieldIdentityKey {
	return orderedIdentityKeys(e.fieldOrder, e.fields)
}

func (e *environment) orderedRemovedKeys() []fieldIdentityKey {
	return orderedIdentityKeys(e.removedOrder, e.removed)
}

func orderedIdentityKeys[T any](order []fieldIdentityKey, values map[fieldIdentityKey]T) []fieldIdentityKey {
	out := make([]fieldIdentityKey, 0, len(order))
	seen := map[fieldIdentityKey]bool{}
	for _, key := range order {
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, key)
	}
	missing := []string{}
	for key := range values {
		if !seen[key] {
			missing = append(missing, string(key))
		}
	}
	sort.Strings(missing)
	for _, key := range missing {
		out = append(out, fieldIdentityKey(key))
	}
	return out
}

func uniqueIDs(groups ...[]string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, g := range groups {
		for _, id := range g {
			if !seen[id] {
				out = append(out, id)
				seen[id] = true
			}
		}
	}
	return out
}
