package analysis

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFieldIdentityKindsKeepPrivateKeysDistinct(t *testing.T) {
	tests := []struct {
		name       string
		identity   fieldIdentity
		publicName string
		exact      bool
	}{
		{name: "atomic dotted", identity: atomicFieldIdentity("actor.name"), publicName: "actor.name", exact: true},
		{name: "structural path", identity: pathFieldIdentity("", []string{"actor", "name"}), publicName: "actor.name", exact: true},
		{name: "qualified path", identity: pathFieldIdentity("left", []string{"actor", "name"}), publicName: "actor.name", exact: true},
		{name: "dynamic access", identity: dynamicFieldIdentity("actor[]"), publicName: "actor[]", exact: false},
	}

	keys := map[fieldIdentityKey]string{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, exact := tc.identity.privateKey()
			if exact != tc.exact || tc.identity.PublicName != tc.publicName {
				t.Fatalf("identity = %+v key=%q exact=%t, want public=%q exact=%t", tc.identity, key, exact, tc.publicName, tc.exact)
			}
			if !exact {
				return
			}
			if previous := keys[key]; previous != "" {
				t.Fatalf("private key %q shared by %s and %s", key, previous, tc.name)
			}
			keys[key] = tc.name
		})
	}
	if len(keys) != 3 {
		t.Fatalf("exact private keys = %v", keys)
	}
}

func TestFieldIdentityPrivateEncodingIsDisjointFromAtomicSpelling(t *testing.T) {
	structural := pathFieldIdentity("left", []string{"actor", "name"})
	structuralKey, exact := structural.privateKey()
	if !exact {
		t.Fatal("structural identity is not exact")
	}
	atomic := atomicFieldIdentity(string(structuralKey))
	atomicKey, exact := atomic.privateKey()
	if !exact {
		t.Fatal("atomic identity is not exact")
	}
	if atomicKey == structuralKey {
		t.Fatalf("atomic spelling collided with structural encoding %q", structuralKey)
	}
	if atomic.PublicName != string(structuralKey) || structural.PublicName != "actor.name" {
		t.Fatalf("private encoding leaked into public names: atomic=%q structural=%q", atomic.PublicName, structural.PublicName)
	}
}

func TestFieldIdentityPublicProjectionKeepsAtomicAndPathSeparate(t *testing.T) {
	query := `FROM main | fields actor.name, 'actor.name'`
	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	reads := []Reference{}
	for _, ref := range result.References {
		if ref.Kind == "field" && ref.Role == "read" && ref.NormalizedName == "actor.name" {
			reads = append(reads, ref)
		}
	}
	if len(reads) != 2 || reads[0].ID != "ref-1" || reads[1].ID != "ref-2" || reads[0].OriginalName != "actor.name" || reads[1].OriginalName != "'actor.name'" {
		t.Fatalf("projection reads = %+v", reads)
	}
	for i, want := range []Location{
		{Start: Position{Offset: 19, Line: 1, Column: 20}, End: Position{Offset: 29, Line: 1, Column: 30}},
		{Start: Position{Offset: 31, Line: 1, Column: 32}, End: Position{Offset: 43, Line: 1, Column: 44}},
	} {
		if reads[i].Location != want {
			t.Fatalf("read %s location = %+v, want %+v", reads[i].ID, reads[i].Location, want)
		}
	}
	if result.Status != Valid || !result.Coverage.SemanticComplete {
		t.Fatalf("analysis = status %q coverage %+v diagnostics %+v", result.Status, result.Coverage, result.Diagnostics)
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeAmbiguousField {
			t.Fatalf("exact identities reported as ambiguous: %+v", diagnostic)
		}
	}
	state := result.Lineage[len(result.Lineage)-1].After
	bindings := []FieldBinding{}
	for _, field := range state.Fields {
		if field.Name == "actor.name" {
			bindings = append(bindings, field)
		}
	}
	if state.Uncertain || len(bindings) != 2 {
		t.Fatalf("projection state = %+v, want two certain actor.name bindings", state)
	}
	if !reflect.DeepEqual(bindings[0].OriginReferenceIDs, []string{"ref-2"}) || !reflect.DeepEqual(bindings[1].OriginReferenceIDs, []string{"ref-1"}) {
		t.Fatalf("projection origins = %+v", bindings)
	}

	// Check the public JSON contract without depending on a private encoded key.
	var projection struct {
		References []struct {
			FieldIdentity *struct {
				Kind     string   `json:"kind"`
				Segments []string `json:"segments"`
			} `json:"field_identity"`
		} `json:"references"`
		Lineage []struct {
			After struct {
				Fields []struct {
					FieldIdentity *struct {
						Kind     string   `json:"kind"`
						Segments []string `json:"segments"`
					} `json:"field_identity"`
				} `json:"fields"`
			} `json:"after"`
		} `json:"lineage"`
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &projection); err != nil {
		t.Fatal(err)
	}
	if len(projection.References) < 3 || projection.References[1].FieldIdentity == nil || projection.References[2].FieldIdentity == nil || projection.References[1].FieldIdentity.Kind != "path" || !reflect.DeepEqual(projection.References[1].FieldIdentity.Segments, []string{"actor", "name"}) || projection.References[2].FieldIdentity.Kind != "atomic" || !reflect.DeepEqual(projection.References[2].FieldIdentity.Segments, []string{"actor.name"}) {
		t.Fatalf("public reference identities = %+v", projection.References)
	}
	publicFields := projection.Lineage[len(projection.Lineage)-1].After.Fields
	if len(publicFields) != 2 || publicFields[0].FieldIdentity == nil || publicFields[1].FieldIdentity == nil || publicFields[0].FieldIdentity.Kind != "atomic" || !reflect.DeepEqual(publicFields[0].FieldIdentity.Segments, []string{"actor.name"}) || publicFields[1].FieldIdentity.Kind != "path" || !reflect.DeepEqual(publicFields[1].FieldIdentity.Segments, []string{"actor", "name"}) {
		t.Fatalf("public binding identities = %+v", publicFields)
	}
}

func TestFieldIdentityPriorUncertaintyDoesNotMergeOrigins(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "sql projection", query: `SELECT actor.name, 'actor.name' FROM main`},
		{name: "uncertain environment", query: `FROM main | eval path=actor.name | lookup people path | eval atomic='actor.name'`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: tc.query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			reads := []Reference{}
			for _, ref := range result.References {
				if ref.Kind == "field" && ref.Role == "read" && ref.NormalizedName == "actor.name" && (ref.OriginalName == "actor.name" || ref.OriginalName == "'actor.name'") {
					reads = append(reads, ref)
				}
			}
			if len(reads) != 2 {
				t.Fatalf("colliding exact reads = %+v", reads)
			}
			state := result.Lineage[len(result.Lineage)-1].After
			bindings := []FieldBinding{}
			for _, field := range state.Fields {
				if field.Name == "actor.name" {
					bindings = append(bindings, field)
				}
			}
			if len(bindings) == 0 || !state.Uncertain {
				t.Fatalf("prior uncertainty was lost: %+v", state)
			}
			for _, binding := range bindings {
				if reflect.DeepEqual(binding.OriginReferenceIDs, []string{reads[0].ID, reads[1].ID}) {
					t.Fatalf("distinct origins were merged: %+v", state)
				}
				if binding.FieldIdentity.Kind == "atomic" && !reflect.DeepEqual(binding.FieldIdentity.Segments, []string{"actor.name"}) || binding.FieldIdentity.Kind == "path" && !reflect.DeepEqual(binding.FieldIdentity.Segments, []string{"actor", "name"}) {
					t.Fatalf("binding identity = %+v", binding)
				}
			}
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Code == CodeAmbiguousField {
					t.Fatalf("distinct identities reported as ambiguous: %+v", diagnostic)
				}
			}
		})
	}
}

func TestFieldIdentityProjectionPreservesSourceOrder(t *testing.T) {
	for _, query := range []string{
		`FROM main | fields actor.name, 'actor.name'`,
		`FROM main | fields 'actor.name', actor.name`,
	} {
		t.Run(query, func(t *testing.T) {
			result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			reads := []Reference{}
			for _, ref := range result.References {
				if ref.Kind == "field" && ref.Role == "read" && ref.NormalizedName == "actor.name" {
					reads = append(reads, ref)
				}
			}
			if len(reads) != 2 {
				t.Fatalf("colliding projection reads = %+v", reads)
			}
			state := result.Lineage[len(result.Lineage)-1].After
			bindings := []FieldBinding{}
			for _, field := range state.Fields {
				if field.Name == "actor.name" {
					bindings = append(bindings, field)
				}
			}
			if len(bindings) != 2 || state.Uncertain || result.Status != Valid {
				t.Fatalf("projection state = %+v status=%q", state, result.Status)
			}
			for _, binding := range bindings {
				wantID := reads[0].ID
				if binding.FieldIdentity.Kind == "atomic" && reads[0].OriginalName == "actor.name" || binding.FieldIdentity.Kind == "path" && reads[0].OriginalName != "actor.name" {
					wantID = reads[1].ID
				}
				if !reflect.DeepEqual(binding.OriginReferenceIDs, []string{wantID}) {
					t.Fatalf("projection binding = %+v, want origin %s", binding, wantID)
				}
			}
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Code == CodeAmbiguousField {
					t.Fatalf("exact projection reported ambiguity: %+v", diagnostic)
				}
			}
			projects := 0
			for _, transition := range result.Lineage[len(result.Lineage)-1].Transitions {
				if transition.Operation == "project" && transition.Output == "actor.name" {
					projects++
				}
			}
			if projects != 2 {
				t.Fatalf("projection transitions = %+v", result.Lineage[len(result.Lineage)-1].Transitions)
			}
		})
	}
}

func TestFieldIdentitySnapshotKeepsConditionalityAndOriginsSeparate(t *testing.T) {
	environment := newEnvironment()
	atomic := atomicFieldIdentity("actor.name")
	structural := pathFieldIdentity("", []string{"actor", "name"})
	environment.installIdentity(atomic, []string{"ref-atomic"}, false, false)
	environment.installIdentity(structural, []string{"ref-path"}, true, false)

	state := environment.snapshot()
	want := FieldState{
		Fields: []FieldBinding{
			{Name: "actor.name", FieldIdentity: FieldIdentity{Kind: "atomic", Segments: []string{"actor.name"}}, OriginReferenceIDs: []string{"ref-atomic"}},
			{Name: "actor.name", FieldIdentity: FieldIdentity{Kind: "path", Segments: []string{"actor", "name"}}, OriginReferenceIDs: []string{"ref-path"}, Conditional: true},
		},
		Removed: []FieldRemoval{},
		Open:    true,
	}
	if !reflect.DeepEqual(state, want) {
		t.Fatalf("identity snapshot = %+v, want %+v", state, want)
	}
}

func TestFieldIdentityReinstalledBindingAppearsOnceInReport(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `FROM main | fields actor.name, 'actor.name' | fields - 'actor.name' | eval 'actor.name'=2`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	state := result.Lineage[len(result.Lineage)-1].After
	if len(state.Fields) != 2 || state.Fields[0].FieldIdentity.Kind != "atomic" || state.Fields[1].FieldIdentity.Kind != "path" {
		t.Fatalf("reinstalled binding duplicated or lost typed peer: %+v", state)
	}
}

func TestFieldIdentityRepeatedRemovalAppearsOnceInReport(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `FROM main | fields actor.name, 'actor.name' | fields - 'actor.name' | eval 'actor.name'=2 | fields - 'actor.name'`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	state := result.Lineage[len(result.Lineage)-1].After
	if len(state.Fields) != 1 || state.Fields[0].FieldIdentity.Kind != "path" || len(state.Removed) != 1 || state.Removed[0].FieldIdentity.Kind != "atomic" {
		t.Fatalf("repeated removal duplicated tombstone or lost typed peer: %+v", state)
	}
}

func TestFieldIdentityDynamicNavigationKeepsStablePublicSpelling(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		original   string
		normalized string
	}{
		{name: "lambda local", query: `FROM main | eval result=filter(items,$it.amount>0)`, original: `$it.amount`, normalized: `$it.amount`},
		{name: "dynamic index", query: `FROM main | where actor[key]=1`, original: `actor[key]`, normalized: `actor[]`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: tc.query, Language: "spl2"})
			if err != nil {
				t.Fatal(err)
			}
			for _, ref := range result.References {
				if ref.OriginalName != tc.original {
					continue
				}
				if ref.NormalizedName != tc.normalized || ref.Resolution != "dynamic" || ref.Binding != "indeterminate" || ref.FieldIdentity != nil {
					t.Fatalf("dynamic reference = %+v, want normalized=%q dynamic/indeterminate", ref, tc.normalized)
				}
				for _, lineage := range result.Lineage {
					for _, field := range lineage.After.Fields {
						if field.Name == tc.normalized {
							t.Fatalf("dynamic identity became an exact field binding: %+v", lineage)
						}
					}
					for _, transition := range lineage.Transitions {
						if transition.Output == tc.normalized {
							t.Fatalf("dynamic identity produced an exact transition: %+v", transition)
						}
					}
				}
				return
			}
			t.Fatalf("missing dynamic reference %q in %+v", tc.original, result.References)
		})
	}
}

func TestFieldIdentityQualifiedJoinReferencesKeepQualifier(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `FROM main | join type=inner left=L right=R where L.id=R.uid [FROM other | table uid]`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		original string
		identity FieldIdentity
	}{
		{original: "L.id", identity: FieldIdentity{Kind: "path", Segments: []string{"id"}, Qualifier: "L"}},
		{original: "R.uid", identity: FieldIdentity{Kind: "path", Segments: []string{"uid"}, Qualifier: "R"}},
	} {
		found := false
		for _, ref := range result.References {
			if ref.Kind != "field" || ref.OriginalName != tc.original {
				continue
			}
			found = true
			if ref.FieldIdentity == nil || !reflect.DeepEqual(*ref.FieldIdentity, tc.identity) {
				t.Fatalf("qualified join reference = %+v, want %+v", ref, tc.identity)
			}
		}
		if !found {
			t.Fatalf("missing qualified join reference %q in %+v", tc.original, result.References)
		}
	}
}

func TestFieldIdentityExactRemovalKeepsOtherPrivateIdentityAvailable(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		readOriginal string
		removedKind  string
	}{
		{name: "atomic removal keeps structural", query: `FROM main | fields - 'actor.name' | eval x=actor.name`, readOriginal: "actor.name", removedKind: "atomic"},
		{name: "structural removal keeps atomic", query: `FROM main | fields - actor.name | eval x='actor.name'`, readOriginal: "'actor.name'", removedKind: "path"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: tc.query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			var read *Reference
			for i := range result.References {
				candidate := &result.References[i]
				if candidate.Kind == "field" && candidate.Role == "read" && candidate.OriginalName == tc.readOriginal {
					read = candidate
				}
			}
			if read == nil || read.Binding != "source" {
				t.Fatalf("surviving private identity read = %+v", read)
			}
			entry := requirementTraceReferencesByID(trace)[read.ID]
			if entry.reference.Binding != "source" || !entry.directExternal || entry.conditional {
				t.Fatalf("surviving requirement identity = %+v", entry)
			}
			state := result.Lineage[len(result.Lineage)-1].After
			removals := publicRemovalIdentities(t, state)
			if len(removals) != 1 || removals[0].Name != "actor.name" || removals[0].FieldIdentity.Kind != tc.removedKind {
				t.Fatalf("typed removal = %+v, want %s actor.name", removals, tc.removedKind)
			}
			count := 0
			for _, field := range state.Fields {
				if field.Name == "actor.name" {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("surviving public field count = %d: %+v", count, state)
			}
		})
	}
}

func TestFieldIdentityWildcardRemovalUsesMatchedPrivateIdentity(t *testing.T) {
	query := `FROM main | eval seed=actor.name | fields - 'actor.*' | eval after=actor.name`
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	reads := []Reference{}
	for _, ref := range result.References {
		if ref.Kind == "field" && ref.Role == "read" && ref.OriginalName == "actor.name" {
			reads = append(reads, ref)
		}
	}
	if len(reads) != 2 || reads[1].Binding != "indeterminate" {
		t.Fatalf("structural reads after wildcard removal = %+v", reads)
	}
	entry := requirementTraceReferencesByID(trace)[reads[1].ID]
	if entry.reference.Binding != reads[1].Binding || entry.directExternal || !entry.conditional {
		t.Fatalf("requirement read after wildcard removal = %+v", entry)
	}
	state := result.Lineage[len(result.Lineage)-1].After
	for _, field := range state.Fields {
		if field.Name == "actor.name" {
			t.Fatalf("wildcard-removed structural field remained live: %+v", state)
		}
	}
	foundRemoved := false
	for _, removed := range state.Removed {
		foundRemoved = foundRemoved || removed.Name == "actor.name"
	}
	if !foundRemoved {
		t.Fatalf("structural removal missing from public state: %+v", state)
	}
}

func TestFieldIdentityWildcardRemovalRetainsBothTransitions(t *testing.T) {
	query := `FROM main | eval structural=actor.name, atomic='actor.name' | fields - 'actor.*'`
	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	state := result.Lineage[len(result.Lineage)-1].After
	if len(state.Removed) != 2 {
		t.Fatalf("wildcard removal state = %+v, want two exact actor.name removals", state)
	}
	reads := []Reference{}
	for _, ref := range result.References {
		if ref.Kind == "field" && ref.Role == "read" && ref.NormalizedName == "actor.name" {
			reads = append(reads, ref)
		}
	}
	if len(reads) != 2 {
		t.Fatalf("colliding wildcard inputs = %+v", reads)
	}
	removals := 0
	for _, transition := range result.Lineage[len(result.Lineage)-1].Transitions {
		if transition.Operation == "remove" && transition.Output == "actor.name" {
			removals++
		}
	}
	if removals != 2 {
		t.Fatalf("wildcard emitted %d removals: %+v", removals, result.Lineage[len(result.Lineage)-1].Transitions)
	}
	publicRemovals := publicRemovalIdentities(t, state)
	if len(publicRemovals) != 2 || publicRemovals[0].FieldIdentity.Kind != "atomic" || publicRemovals[1].FieldIdentity.Kind != "path" {
		t.Fatalf("wildcard typed removals = %+v", publicRemovals)
	}
	publicTransitions := publicTransitionIdentities(t, result.Lineage[len(result.Lineage)-1].Transitions)
	removedKinds := map[string]bool{}
	for _, transition := range publicTransitions {
		if transition.Operation == "remove" && transition.Output == "actor.name" {
			removedKinds[transition.OutputIdentity.Kind] = true
		}
	}
	if !removedKinds["atomic"] || !removedKinds["path"] {
		t.Fatalf("wildcard typed transitions = %+v", publicTransitions)
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeAmbiguousField {
			t.Fatalf("exact wildcard identities reported as ambiguous: %+v", diagnostic)
		}
	}
}

func TestFieldIdentityWildcardSelectionHasTypedTransitions(t *testing.T) {
	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: `FROM main | eval structural=actor.name, atomic='actor.name' | fields 'actor.*'`, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	transitions := publicTransitionIdentities(t, result.Lineage[len(result.Lineage)-1].Transitions)
	kinds := map[string]bool{}
	for _, transition := range transitions {
		if transition.Operation == "project" && transition.Output == "actor.name" {
			kinds[transition.OutputIdentity.Kind] = true
		}
	}
	if !kinds["atomic"] || !kinds["path"] {
		t.Fatalf("wildcard typed selection transitions = %+v", transitions)
	}
}

func TestFieldIdentityAggregateAliasHasTypedTransition(t *testing.T) {
	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: `FROM main | stats count() AS 'actor.name' BY actor.name`, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	transitions := publicTransitionIdentities(t, result.Lineage[len(result.Lineage)-1].Transitions)
	kinds := map[string]bool{}
	for _, transition := range transitions {
		if transition.Output == "actor.name" {
			kinds[transition.Operation+":"+transition.OutputIdentity.Kind] = true
		}
	}
	if !kinds["project:path"] || !kinds["aggregate:atomic"] {
		t.Fatalf("typed aggregate group and alias transitions = %+v", transitions)
	}
}

func TestFieldIdentityUnprovedGroupDoesNotClaimExactOutput(t *testing.T) {
	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: `FROM [{known:1}] | stats count() AS n BY missing`, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, transition := range result.Lineage[len(result.Lineage)-1].Transitions {
		if transition.Operation == "project" && transition.Output == "missing" {
			if transition.OutputIdentity != nil {
				t.Fatalf("unproved group claimed exact output: %+v", transition)
			}
			return
		}
	}
	t.Fatal("missing group transition was unexpectedly removed")
}

func publicRemovalIdentities(t *testing.T, state FieldState) []struct {
	Name          string        `json:"name"`
	FieldIdentity FieldIdentity `json:"field_identity"`
} {
	t.Helper()
	var public struct {
		Removed []struct {
			Name          string        `json:"name"`
			FieldIdentity FieldIdentity `json:"field_identity"`
		} `json:"removed"`
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &public); err != nil {
		t.Fatalf("public removals lack typed entries: %v", err)
	}
	return public.Removed
}

func publicTransitionIdentities(t *testing.T, transitions []Transition) []struct {
	Operation      string        `json:"operation"`
	Output         string        `json:"output"`
	OutputIdentity FieldIdentity `json:"output_identity"`
} {
	t.Helper()
	var public []struct {
		Operation      string        `json:"operation"`
		Output         string        `json:"output"`
		OutputIdentity FieldIdentity `json:"output_identity"`
	}
	encoded, err := json.Marshal(transitions)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &public); err != nil {
		t.Fatal(err)
	}
	return public
}
