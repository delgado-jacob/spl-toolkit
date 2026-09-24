package analysis

import (
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

func TestFieldIdentityPublicCollisionIsExplicitAndConservative(t *testing.T) {
	query := `FROM main | eval atomic='actor.name', structural=actor.name | fields - 'actor.name'`
	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != Incomplete || result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete {
		t.Fatalf("coverage = analysis %q/%+v requirements %+v", result.Status, result.Coverage, result.Requirements.Coverage)
	}

	reads := []Reference{}
	for _, ref := range result.References {
		if ref.Kind == "field" && ref.Role == "read" && ref.NormalizedName == "actor.name" {
			reads = append(reads, ref)
		}
	}
	if len(reads) != 2 || reads[0].OriginalName == reads[1].OriginalName {
		t.Fatalf("colliding reads = %+v", reads)
	}
	wantOrigins := []string{reads[0].ID, reads[1].ID}
	collisionDiagnostics := 0
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeAmbiguousField {
			collisionDiagnostics++
			if diagnostic.Category != "unsupported_semantics" {
				t.Fatalf("ambiguity category = %q", diagnostic.Category)
			}
		}
	}
	if collisionDiagnostics != 1 {
		t.Fatalf("ambiguity diagnostics = %d: %+v", collisionDiagnostics, result.Diagnostics)
	}

	evalState := result.Lineage[1].After
	bindings := []FieldBinding{}
	for _, field := range evalState.Fields {
		if field.Name == "actor.name" {
			bindings = append(bindings, field)
		}
	}
	if !evalState.Uncertain || len(bindings) != 1 || !reflect.DeepEqual(bindings[0].OriginReferenceIDs, wantOrigins) {
		t.Fatalf("collision state = %+v, want one actor.name with origins %v", evalState, wantOrigins)
	}
	for _, lineage := range result.Lineage {
		for _, transition := range lineage.Transitions {
			if transition.Output == "actor.name" {
				t.Fatalf("ambiguous identity produced transition: %+v", transition)
			}
		}
	}
	assertRequirementGap(t, result.Requirements.Gaps, CodeAmbiguousField, wantOrigins, []string{CodeAmbiguousField})
}

func TestFieldIdentityCollisionSurvivesPriorUncertainty(t *testing.T) {
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
			wantOrigins := []string{reads[0].ID, reads[1].ID}
			state := result.Lineage[len(result.Lineage)-1].After
			bindings := []FieldBinding{}
			for _, field := range state.Fields {
				if field.Name == "actor.name" {
					bindings = append(bindings, field)
				}
			}
			if len(bindings) != 1 || !state.Uncertain || !reflect.DeepEqual(bindings[0].OriginReferenceIDs, wantOrigins) {
				t.Fatalf("collision state = %+v, want origins %v", state, wantOrigins)
			}
			ambiguities := 0
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Code == CodeAmbiguousField {
					ambiguities++
					if diagnostic.Category != "unsupported_semantics" {
						t.Fatalf("ambiguity category = %q", diagnostic.Category)
					}
				}
			}
			if ambiguities != 1 {
				t.Fatalf("ambiguity diagnostics = %d: %+v", ambiguities, result.Diagnostics)
			}
			for _, lineage := range result.Lineage {
				for _, transition := range lineage.Transitions {
					if transition.Output == "actor.name" {
						t.Fatalf("ambiguous identity produced transition: %+v", transition)
					}
				}
			}
			assertRequirementGap(t, result.Requirements.Gaps, CodeAmbiguousField, wantOrigins, []string{CodeAmbiguousField})
		})
	}
}

func TestFieldIdentityProjectionCollisionPreservesSourceOrder(t *testing.T) {
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
			wantOrigins := []string{reads[0].ID, reads[1].ID}
			state := result.Lineage[len(result.Lineage)-1].After
			bindings := []FieldBinding{}
			for _, field := range state.Fields {
				if field.Name == "actor.name" {
					bindings = append(bindings, field)
				}
			}
			if len(bindings) != 1 || !state.Uncertain || !reflect.DeepEqual(bindings[0].OriginReferenceIDs, wantOrigins) {
				t.Fatalf("projection state = %+v, want one uncertain actor.name with origins %v", state, wantOrigins)
			}
			ambiguities := 0
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Code != CodeAmbiguousField {
					continue
				}
				ambiguities++
				if diagnostic.Category != "unsupported_semantics" {
					t.Fatalf("ambiguity category = %q", diagnostic.Category)
				}
			}
			if ambiguities != 1 {
				t.Fatalf("ambiguity diagnostics = %d: %+v", ambiguities, result.Diagnostics)
			}
			for _, transition := range result.Lineage[len(result.Lineage)-1].Transitions {
				if transition.Operation == "project" && transition.Output == "actor.name" {
					t.Fatalf("ambiguous projection emitted an identity-specific transition: %+v", transition)
				}
			}
			assertRequirementGap(t, result.Requirements.Gaps, CodeAmbiguousField, wantOrigins, []string{CodeAmbiguousField})
			if result.Requirements.Coverage.Complete {
				t.Fatalf("ambiguous projection requirements are complete: %+v", result.Requirements.Coverage)
			}
		})
	}
}

func TestFieldIdentitySnapshotCombinesConditionalityAndOrigins(t *testing.T) {
	environment := newEnvironment()
	atomic := atomicFieldIdentity("actor.name")
	structural := pathFieldIdentity("", []string{"actor", "name"})
	environment.installIdentity(atomic, []string{"ref-atomic"}, false, false)
	environment.installIdentity(structural, []string{"ref-path"}, true, false)

	state := environment.snapshot()
	want := FieldState{
		Fields:    []FieldBinding{{Name: "actor.name", OriginReferenceIDs: []string{"ref-atomic", "ref-path"}, Conditional: true}},
		Removed:   []string{},
		Open:      true,
		Uncertain: true,
	}
	if !reflect.DeepEqual(state, want) {
		t.Fatalf("collision snapshot = %+v, want %+v", state, want)
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
				if ref.NormalizedName != tc.normalized || ref.Resolution != "dynamic" || ref.Binding != "indeterminate" {
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

func TestFieldIdentityExactRemovalKeepsOtherPrivateIdentityAvailable(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		readOriginal string
	}{
		{name: "atomic removal keeps structural", query: `FROM main | fields - 'actor.name' | eval x=actor.name`, readOriginal: "actor.name"},
		{name: "structural removal keeps atomic", query: `FROM main | fields - actor.name | eval x='actor.name'`, readOriginal: "'actor.name'"},
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
			count := 0
			for _, field := range state.Fields {
				if field.Name == "actor.name" {
					count++
				}
			}
			for _, removed := range state.Removed {
				if removed == "actor.name" {
					t.Fatalf("live public field also reported removed: %+v", state)
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
		foundRemoved = foundRemoved || removed == "actor.name"
	}
	if !foundRemoved {
		t.Fatalf("structural removal missing from public state: %+v", state)
	}
}

func TestFieldIdentityWildcardRemovalCollapsesPublicCollision(t *testing.T) {
	query := `FROM main | eval structural=actor.name, atomic='actor.name' | fields - 'actor.*'`
	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	state := result.Lineage[len(result.Lineage)-1].After
	if !reflect.DeepEqual(state.Removed, []string{"actor.name"}) {
		t.Fatalf("wildcard removal state = %+v, want one conservative actor.name removal", state)
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
	wantOrigins := []string{reads[0].ID, reads[1].ID}
	removals := 0
	for _, transition := range result.Lineage[len(result.Lineage)-1].Transitions {
		if transition.Operation == "remove" && transition.Output == "actor.name" {
			removals++
		}
	}
	if removals != 0 {
		t.Fatalf("ambiguous wildcard emitted %d identity-specific removals: %+v", removals, result.Lineage[len(result.Lineage)-1].Transitions)
	}
	ambiguities := 0
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeAmbiguousField {
			ambiguities++
			if diagnostic.Category != "unsupported_semantics" {
				t.Fatalf("ambiguity category = %q", diagnostic.Category)
			}
		}
	}
	if ambiguities != 1 || result.Requirements.Coverage.Complete {
		t.Fatalf("collision evidence = diagnostics %+v requirements %+v", result.Diagnostics, result.Requirements.Coverage)
	}
	assertRequirementGap(t, result.Requirements.Gaps, CodeAmbiguousField, wantOrigins, []string{CodeAmbiguousField})
}
