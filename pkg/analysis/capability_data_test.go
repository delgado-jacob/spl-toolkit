package analysis

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestEmbeddedCapabilityAssetsAreStructurallyValid(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	if len(records) == 0 || len(cases) == 0 {
		t.Fatalf("embedded capability data must not be empty: records=%d cases=%d", len(records), len(cases))
	}
	var authored struct {
		Records []map[string]json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(embeddedCapabilityLedger, &authored); err != nil {
		t.Fatalf("inspect embedded capability ledger: %v", err)
	}
	for i, record := range authored.Records {
		raw, present := record["grammar_registered"]
		var registered bool
		if !present || json.Unmarshal(raw, &registered) != nil {
			t.Errorf("embedded capability record %d lacks an explicit boolean grammar_registered value", i)
		}
	}
}

func TestDecodeCapabilityAssetsAcceptsOptionalStructuredSemanticExpectations(t *testing.T) {
	ledger, corpus := validCapabilityAssets(t)
	var file map[string]any
	mustUnmarshal(t, corpus, &file)
	cases := file["cases"].([]any)
	observations := cases[0].(map[string]any)["observations"].(map[string]any)
	semantics := observations["semantics"].(map[string]any)
	semantics["scopes"] = []any{}
	semantics["lineage"] = []any{}
	semantics["final_field_state"] = map[string]any{
		"fields":    []any{},
		"removed":   []any{},
		"open":      true,
		"uncertain": false,
	}

	if _, _, err := decodeCapabilityAssets(ledger, mustJSON(t, file)); err != nil {
		t.Fatalf("optional structured semantic expectations were rejected: %v", err)
	}
}

func TestCapabilityFieldStateDistinguishesEqualDisplayIdentities(t *testing.T) {
	var state CapabilityFieldStateExpectation
	data := []byte(`{"fields":[{"name":"actor.name","field_identity":{"kind":"atomic","segments":["actor.name"]},"origin_reference_ids":["reference-1"],"conditional":false},{"name":"actor.name","field_identity":{"kind":"path","segments":["actor","name"]},"origin_reference_ids":["reference-2"],"conditional":false}],"removed":[],"open":false,"uncertain":false}`)
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("decode typed field state: %v", err)
	}
	if err := validateCapabilityFieldState("identity", &state, map[string]int{"reference-1": 0, "reference-2": 1}, nil); err != nil {
		t.Fatalf("distinct identities with equal display names were rejected: %v", err)
	}
	state.Fields = state.Fields[1:]
	state.Removed = []CapabilityFieldRemovalExpectation{{Name: "actor.name", FieldIdentity: &FieldIdentity{Kind: "atomic", Segments: []string{"actor.name"}}}}
	if err := validateCapabilityFieldState("identity", &state, map[string]int{"reference-1": 0, "reference-2": 1}, nil); err != nil {
		t.Fatalf("path field and atomic removal with equal display names were rejected: %v", err)
	}
	state.Removed[0].FieldIdentity = &FieldIdentity{Kind: "path", Segments: []string{"actor", "name"}}
	if err := validateCapabilityFieldState("identity", &state, map[string]int{"reference-1": 0, "reference-2": 1}, nil); err == nil {
		t.Fatal("same exact identity was accepted as both present and removed")
	}
}

func TestCapabilityTransitionIdentityMustMatchOutputReference(t *testing.T) {
	_, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatal(err)
	}
	for _, evidence := range cases {
		if evidence.ID != "spl2.fields.exact-field-list.positive" {
			continue
		}
		cloned := cloneCapabilityEvidence(evidence)
		cloned.Observations.Semantics.Transitions[0].OutputIdentity.Kind = "path"
		if err := validateCapabilityEvidenceObservation(cloned, "semantics"); err == nil {
			t.Fatal("transition and output reference with different exact identities were accepted")
		}
		return
	}
	t.Fatal("missing exact field-list evidence")
}

func TestCapabilityIdentityCollisionEvidenceIsPositive(t *testing.T) {
	_, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatal(err)
	}
	for _, evidence := range cases {
		if evidence.ID != "spl2.field.identity-collision.positive" {
			continue
		}
		if evidence.Classification != CapabilityEvidencePositive || evidence.Observations.Semantics == nil || !evidence.Observations.Semantics.Complete || evidence.Observations.Requirements == nil || !evidence.Observations.Requirements.Complete {
			t.Fatalf("identity collision evidence is not complete positive proof: %+v", evidence)
		}
		if len(evidence.Observations.Requirements.Items) != 3 {
			t.Fatalf("identity collision evidence has %d requirements, want dataset and two fields", len(evidence.Observations.Requirements.Items))
		}
		path := evidence.Observations.Requirements.Items[1].FieldIdentity
		atomic := evidence.Observations.Requirements.Items[2].FieldIdentity
		if path == nil || path.Kind != "path" || !slices.Equal(path.Segments, []string{"actor", "name"}) || atomic == nil || atomic.Kind != "atomic" || !slices.Equal(atomic.Segments, []string{"actor.name"}) {
			t.Fatalf("identity collision evidence lacks separate exact field requirements: %+v", evidence.Observations.Requirements.Items)
		}
		fields := evidence.Observations.Semantics.FinalFieldState.Fields
		if len(fields) != 2 || fields[0].FieldIdentity == nil || fields[0].FieldIdentity.Kind != "atomic" || !slices.Equal(fields[0].OriginReferenceIDs, []string{"ref-2"}) || fields[1].FieldIdentity == nil || fields[1].FieldIdentity.Kind != "path" || !slices.Equal(fields[1].OriginReferenceIDs, []string{"ref-1"}) {
			t.Fatalf("identity collision evidence merged field origins: %+v", fields)
		}
		cloned := cloneCapabilityEvidence(evidence)
		cloned.Observations.Semantics.References[1].FieldIdentity.Segments[0] = "changed"
		cloned.Observations.Semantics.FinalFieldState.Fields[0].FieldIdentity.Segments[0] = "changed"
		cloned.Observations.Requirements.Items[1].FieldIdentity.Segments[0] = "changed"
		if evidence.Observations.Semantics.References[1].FieldIdentity.Segments[0] == "changed" || evidence.Observations.Semantics.FinalFieldState.Fields[0].FieldIdentity.Segments[0] == "changed" || evidence.Observations.Requirements.Items[1].FieldIdentity.Segments[0] == "changed" {
			t.Fatal("capability evidence clone aliases private field identities")
		}
		return
	}
	t.Fatal("missing positive identity-collision evidence")
}

func TestCapabilityIdentityCollisionProofIsDistinctFromMerge(t *testing.T) {
	_, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatal(err)
	}
	for _, evidence := range cases {
		if evidence.ID == "spl2.field.identity-collision.positive" {
			if capabilitySemanticsHaveMergeFacts(evidence.Observations.Semantics) {
				t.Fatal("two exact fields sharing a display name were counted as merge proof")
			}
			if err := validateCapabilitySemanticProof(evidence.Observations.Semantics, capabilityProofIdentity); err != nil {
				t.Fatalf("distinct exact identities did not satisfy their proof category: %v", err)
			}
			cloned := cloneCapabilityEvidence(evidence)
			cloned.Observations.Semantics.FinalFieldState.Fields[1].FieldIdentity = cloneFieldIdentityPointer(cloned.Observations.Semantics.FinalFieldState.Fields[0].FieldIdentity)
			if err := validateCapabilitySemanticProof(cloned.Observations.Semantics, capabilityProofIdentity); err == nil {
				t.Fatal("equal exact identities satisfied collision proof")
			}
			return
		}
	}
	t.Fatal("missing identity collision evidence")
}

func TestCapabilityIdentityCollisionRequirementClaimNeedsDistinctTypedProof(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]map[string]any)
	}{
		{"both identities omitted", func(fields []map[string]any) {
			delete(fields[0], "field_identity")
			delete(fields[1], "field_identity")
		}},
		{"one identity omitted", func(fields []map[string]any) {
			delete(fields[0], "field_identity")
		}},
		{"identities collapsed", func(fields []map[string]any) {
			fields[1]["field_identity"] = fields[0]["field_identity"]
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			corpus := embeddedCapabilityCorpusMap(t)
			var fields []map[string]any
			for _, item := range corpus["cases"].([]any) {
				evidence := item.(map[string]any)
				if evidence["id"] != "spl2.field.identity-collision.positive" {
					continue
				}
				items := evidence["observations"].(map[string]any)["requirements"].(map[string]any)["items"].([]any)
				for _, item := range items {
					requirement := item.(map[string]any)
					if requirement["kind"] == "field" {
						fields = append(fields, requirement)
					}
				}
				break
			}
			if len(fields) != 2 {
				t.Fatalf("identity collision evidence has %d field requirements", len(fields))
			}
			tc.mutate(fields)
			if _, _, err := decodeCapabilityAssets(embeddedCapabilityLedger, mustJSON(t, corpus)); err == nil {
				t.Fatal("supported identity-collision requirements claim accepted non-distinct field evidence")
			}
		})
	}
}

func TestCapabilityRevisionIncludesPrivateRequirementIdentity(t *testing.T) {
	manifest, err := CapabilitiesFor(CapabilityOptions{Language: "spl2", Profile: "splunkd", Version: "current"})
	if err != nil {
		t.Fatal(err)
	}
	baseline := capabilityRevisionTestDigest(t, manifest)
	publicBaseline := string(mustJSON(t, manifest))
	cloned := cloneCapabilityManifest(manifest)
	for i := range cloned.Evidence {
		if cloned.Evidence[i].ID != "spl2.field.identity-collision.positive" {
			continue
		}
		cloned.Evidence[i].Observations.Requirements.Items[1].FieldIdentity.Segments[0] = "changed"
		if capabilityRevisionTestDigest(t, cloned) == baseline {
			t.Fatal("private requirement identity did not affect capability revision")
		}
		if string(mustJSON(t, cloned)) != publicBaseline {
			t.Fatal("private requirement identity changed public capability JSON")
		}
		return
	}
	t.Fatal("missing identity collision evidence")
}

func TestCapabilityStructuredSemanticExpectationsValidation(t *testing.T) {
	valid := validStructuredCapabilityEvidence()
	if err := validateCapabilityEvidenceObservation(valid, "semantics"); err != nil {
		t.Fatalf("valid structured semantic evidence: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*CapabilityEvidence)
	}{
		{
			name: "duplicate scope ID",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Scopes = append(evidence.Observations.Semantics.Scopes, evidence.Observations.Semantics.Scopes[0])
			},
		},
		{
			name: "scope IDs out of canonical order",
			mutate: func(evidence *CapabilityEvidence) {
				scope := evidence.Observations.Semantics.Scopes[0]
				scope.ID = "scope-0"
				evidence.Observations.Semantics.Scopes = append(evidence.Observations.Semantics.Scopes, scope)
			},
		},
		{
			name: "scope names unknown stage",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Scopes[0].StageID = "stage-missing"
			},
		},
		{
			name: "stage names unknown scope",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Stages[0].ScopeID = "scope-missing"
			},
		},
		{
			name: "lineage names unknown scope",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Lineage[0].ScopeID = "scope-missing"
			},
		},
		{
			name: "lineage names unknown stage",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Lineage[0].StageID = "stage-missing"
			},
		},
		{
			name: "lineage omits both states",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Lineage[0].Before = nil
				evidence.Observations.Semantics.Lineage[0].After = nil
			},
		},
		{
			name: "fields out of canonical order",
			mutate: func(evidence *CapabilityEvidence) {
				state := evidence.Observations.Semantics.Lineage[0].After
				state.Fields[0], state.Fields[1] = state.Fields[1], state.Fields[0]
			},
		},
		{
			name: "field names unknown origin",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Lineage[0].After.Fields[0].OriginReferenceIDs = []string{"reference-missing"}
			},
		},
		{
			name: "field origins reverse producer order",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Lineage[0].After.Fields[1].OriginReferenceIDs = []string{"reference-2", "reference-1"}
			},
		},
		{
			name: "transition names unknown input reference",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Transitions[0].InputReferenceIDs = []string{"reference-missing"}
			},
		},
		{
			name: "transition names non-output reference",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Transitions[0].OutputReferenceID = "reference-1"
			},
		},
		{
			name: "transition inputs reverse producer order",
			mutate: func(evidence *CapabilityEvidence) {
				reference := evidence.Observations.Semantics.References[0]
				reference.ID = "reference-3"
				reference.NormalizedName = "other"
				evidence.Observations.Semantics.References = append(evidence.Observations.Semantics.References, reference)
				evidence.Observations.Semantics.Transitions[0].InputReferenceIDs = []string{"reference-3", "reference-1"}
			},
		},
		{
			name: "final state fields are not exact array",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.FinalFieldState.Fields = nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			evidence := cloneCapabilityEvidence(valid)
			tc.mutate(&evidence)
			if err := validateCapabilityEvidenceObservation(evidence, "semantics"); err == nil {
				t.Fatal("malformed structured semantic evidence was accepted")
			}
		})
	}

	t.Run("double-digit scope ordinals follow producer order", func(t *testing.T) {
		evidence := cloneCapabilityEvidence(valid)
		evidence.Observations.Semantics.Stages[0].ScopeID = "scope-9"
		evidence.Observations.Semantics.Scopes[0].ID = "scope-9"
		evidence.Observations.Semantics.Lineage[0].ScopeID = "scope-9"
		scope := evidence.Observations.Semantics.Scopes[0]
		scope.ID = "scope-10"
		evidence.Observations.Semantics.Scopes = append(evidence.Observations.Semantics.Scopes, scope)
		if err := validateCapabilityEvidenceObservation(evidence, "semantics"); err != nil {
			t.Fatalf("numeric scope producer order was rejected: %v", err)
		}
	})
}

func TestMilestone11SemanticClaimsRequireStructuredProofCategories(t *testing.T) {
	tests := []struct {
		name     string
		recordID string
		mutate   func(map[string]any)
	}{
		{
			name:     "guarded branch cannot discard all private proof",
			recordID: "spl2.pipeline.branch.guarded-arms",
			mutate: func(observation map[string]any) {
				delete(observation, "scopes")
				delete(observation, "lineage")
				delete(observation, "final_field_state")
				for _, item := range observation["stages"].([]any) {
					stage := item.(map[string]any)
					delete(stage, "id")
					delete(stage, "scope_id")
				}
				for _, item := range observation["references"].([]any) {
					delete(item.(map[string]any), "id")
				}
				for _, item := range observation["transitions"].([]any) {
					transition := item.(map[string]any)
					delete(transition, "input_reference_ids")
					delete(transition, "output_reference_id")
					delete(transition, "conditional")
				}
			},
		},
		{
			name:     "guarded branch requires merge facts",
			recordID: "spl2.pipeline.branch.guarded-arms",
			mutate: func(observation map[string]any) {
				clearCapabilityMergeFactsMap(observation)
			},
		},
		{
			name:     "local view requires scope facts",
			recordID: "spl2.module.module.local-view",
			mutate: func(observation map[string]any) {
				delete(observation, "scopes")
			},
		},
		{
			name:     "static dataset requires lineage facts",
			recordID: "spl2.dataset.dataset.static-descriptor",
			mutate: func(observation map[string]any) {
				delete(observation, "lineage")
			},
		},
		{
			name:     "structural field path requires origin facts",
			recordID: "spl2.expression.field.structural-path",
			mutate: func(observation map[string]any) {
				clearCapabilityOriginsMap(observation)
			},
		},
		{
			name:     "eval assignment requires expanded transition facts",
			recordID: "spl2.command.eval.exact-assignment",
			mutate: func(observation map[string]any) {
				for _, item := range observation["transitions"].([]any) {
					transition := item.(map[string]any)
					delete(transition, "input_reference_ids")
					delete(transition, "output_reference_id")
					delete(transition, "conditional")
				}
			},
		},
		{
			name:     "static dataset requires final field state",
			recordID: "spl2.dataset.dataset.static-descriptor",
			mutate: func(observation map[string]any) {
				delete(observation, "final_field_state")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ledger capabilityLedgerFile
			mustUnmarshal(t, embeddedCapabilityLedger, &ledger)
			evidenceID := semanticEvidenceIDForRecord(t, ledger, tc.recordID)
			corpus := embeddedCapabilityCorpusMap(t)
			tc.mutate(capabilitySemanticsMap(t, corpus, evidenceID))
			errorOwner := capabilityEvidenceErrorOwner(t, corpus, evidenceID)
			if _, _, err := decodeCapabilityAssets(embeddedCapabilityLedger, mustJSON(t, corpus)); err == nil {
				t.Fatalf("record %q accepted missing structured proof", tc.recordID)
			} else if !strings.Contains(err.Error(), evidenceID) && !strings.Contains(err.Error(), errorOwner) {
				t.Fatalf("record %q mutation failed unrelated evidence: %v", tc.recordID, err)
			}
		})
	}
}

func TestMilestone11MergeOriginsRejectReversedProducerOrder(t *testing.T) {
	var ledger capabilityLedgerFile
	mustUnmarshal(t, embeddedCapabilityLedger, &ledger)
	evidenceID := semanticEvidenceIDForRecord(t, ledger, "spl2.pipeline.branch.guarded-arms")
	corpus := embeddedCapabilityCorpusMap(t)
	observation := capabilitySemanticsMap(t, corpus, evidenceID)
	errorOwner := capabilityEvidenceErrorOwner(t, corpus, evidenceID)
	mutated := false
	for _, item := range observation["final_field_state"].(map[string]any)["fields"].([]any) {
		field := item.(map[string]any)
		origins := field["origin_reference_ids"].([]any)
		if len(origins) < 3 {
			continue
		}
		slices.Reverse(origins)
		mutated = true
		break
	}
	if !mutated {
		t.Fatal("guarded branch evidence has no merged origin chain")
	}
	if _, _, err := decodeCapabilityAssets(embeddedCapabilityLedger, mustJSON(t, corpus)); err == nil {
		t.Fatal("reversed merged producer origins were accepted")
	} else if !strings.Contains(err.Error(), evidenceID) && !strings.Contains(err.Error(), errorOwner) {
		t.Fatalf("reversed merge mutation failed unrelated evidence: %v", err)
	}
}

func TestMilestone11SemanticProofRequirementsCoverExactClaimInventory(t *testing.T) {
	want := []string{
		"spl2.command.bin.span-field",
		"spl2.command.eval.exact-assignment",
		"spl2.command.fields.exact-field-list",
		"spl2.command.from.dataset",
		"spl2.command.if.subpipe",
		"spl2.command.join.qualified-subsearch",
		"spl2.command.mvexpand.limited-field",
		"spl2.command.select.projection",
		"spl2.command.stats.aggregate-call",
		"spl2.command.union.dataset",
		"spl2.command.where.predicate",
		"spl2.dataset.dataset.dynamic-descriptor",
		"spl2.dataset.dataset.parameter",
		"spl2.dataset.dataset.static-descriptor",
		"spl2.expression.field.identity-collision",
		"spl2.expression.field.quoted-dotted-atom",
		"spl2.expression.field.structural-path",
		"spl2.expression.function-call.invalid-selected-arity",
		"spl2.function.abs.one-positional",
		"spl2.function.any.lambda-positional",
		"spl2.function.avg.one-positional",
		"spl2.function.cidrmatch.two-positional",
		"spl2.function.coalesce.two-positional",
		"spl2.function.count.zero-positional",
		"spl2.function.dc.one-positional",
		"spl2.function.distinct_count.one-positional",
		"spl2.function.json.one-positional",
		"spl2.function.json_array_to_mv.one-positional",
		"spl2.function.like.two-positional",
		"spl2.function.lower.one-positional",
		"spl2.function.match.two-positional",
		"spl2.function.max.one-positional",
		"spl2.function.min.one-positional",
		"spl2.function.mvindex.two-positional",
		"spl2.function.round.one-positional",
		"spl2.function.rtrim.one-positional",
		"spl2.function.span.grouping-positional",
		"spl2.function.sqrt.one-positional",
		"spl2.function.stdev.one-positional",
		"spl2.function.strftime.two-positional",
		"spl2.function.sum.one-positional",
		"spl2.function.tonumber.one-positional",
		"spl2.function.values.one-positional",
		"spl2.module.module.declaration-cycle",
		"spl2.module.module.local-scalar-function",
		"spl2.module.module.local-view",
		"spl2.module.module.unresolved-import",
		"spl2.pipeline.branch.guarded-arms",
		"spl2.pipeline.branch.unguarded-arms",
		"spl2.pipeline.join.output-collision",
		"spl2.pipeline.join.qualified-left-subsearch",
		"spl2.pipeline.join.qualified-outer-subsearch",
	}
	got := make([]string, 0, len(milestone11SemanticProofRequirements))
	for recordID, requirement := range milestone11SemanticProofRequirements {
		if requirement == 0 {
			t.Errorf("record %q has no semantic proof requirements", recordID)
		}
		got = append(got, recordID)
	}
	sort.Strings(got)
	if !slices.Equal(got, want) {
		t.Fatalf("semantic proof inventory differs\n got: %v\nwant: %v", got, want)
	}
}

func TestCapabilityRevisionIncludesPrivateSemanticProof(t *testing.T) {
	manifest, err := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	baseline := capabilityRevisionTestDigest(t, manifest)
	publicBaseline := string(mustJSON(t, manifest))

	mutations := []struct {
		name   string
		mutate func(*CapabilitySemanticsObservation) bool
	}{
		{"scopes", func(observation *CapabilitySemanticsObservation) bool {
			observation.Scopes[0].Kind += "-mutated"
			return true
		}},
		{"lineage state", func(observation *CapabilitySemanticsObservation) bool {
			for i := range observation.Lineage {
				if observation.Lineage[i].After != nil {
					observation.Lineage[i].After.Removed = append(observation.Lineage[i].After.Removed, CapabilityFieldRemovalExpectation{Name: "revision-only"})
					return true
				}
			}
			return false
		}},
		{"origins", func(observation *CapabilitySemanticsObservation) bool {
			return mutateCapabilityRevisionField(observation, func(field *CapabilityFieldExpectation) {
				field.OriginReferenceIDs[0] += "-mutated"
			}, func(field CapabilityFieldExpectation) bool { return len(field.OriginReferenceIDs) != 0 })
		}},
		{"conditionality", func(observation *CapabilitySemanticsObservation) bool {
			return mutateCapabilityRevisionField(observation, func(field *CapabilityFieldExpectation) {
				field.Conditional = !field.Conditional
			}, func(CapabilityFieldExpectation) bool { return true })
		}},
		{"openness", func(observation *CapabilitySemanticsObservation) bool {
			observation.Lineage[0].After.Open = !observation.Lineage[0].After.Open
			return true
		}},
		{"uncertainty", func(observation *CapabilitySemanticsObservation) bool {
			observation.Lineage[0].After.Uncertain = !observation.Lineage[0].After.Uncertain
			return true
		}},
		{"final state", func(observation *CapabilitySemanticsObservation) bool {
			observation.FinalFieldState.Removed = append(observation.FinalFieldState.Removed, CapabilityFieldRemovalExpectation{Name: "revision-only"})
			return true
		}},
		{"expanded transition references", func(observation *CapabilitySemanticsObservation) bool {
			for i := range observation.Transitions {
				if len(observation.Transitions[i].InputReferenceIDs) != 0 {
					observation.Transitions[i].InputReferenceIDs[0] += "-mutated"
					return true
				}
			}
			return false
		}},
	}

	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			candidate := cloneCapabilityManifest(manifest)
			observation := capabilityRevisionSemanticObservation(t, &candidate, "spl2.branch.guarded-arms.positive")
			if !tc.mutate(observation) {
				t.Fatal("representative evidence lacks the private fact to mutate")
			}
			if got := capabilityRevisionTestDigest(t, candidate); got == baseline {
				t.Fatalf("%s mutation did not change capability revision payload", tc.name)
			}
			if got := string(mustJSON(t, candidate)); got != publicBaseline {
				t.Fatalf("%s private mutation changed public capability JSON", tc.name)
			}
		})
	}

	const recordID = "spl2.pipeline.branch.guarded-arms"
	requirements := make(map[string]capabilitySemanticProofRequirement, len(milestone11SemanticProofRequirements))
	keys := make([]string, 0, len(milestone11SemanticProofRequirements))
	for id, requirement := range milestone11SemanticProofRequirements {
		requirements[id] = requirement
		keys = append(keys, id)
	}
	mutatedRequirements := make(map[string]capabilitySemanticProofRequirement, len(requirements))
	for id, requirement := range requirements {
		mutatedRequirements[id] = requirement
	}
	mutatedRequirements[recordID] ^= capabilityProofMerge
	if got := capabilityRevisionTestDigestForRequirements(t, manifest, mutatedRequirements); got == baseline {
		t.Error("proof metadata mutation did not change capability revision payload")
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	reorderedRequirements := make(map[string]capabilitySemanticProofRequirement, len(requirements))
	for _, id := range keys {
		reorderedRequirements[id] = requirements[id]
	}
	if got := capabilityRevisionTestDigestForRequirements(t, manifest, reorderedRequirements); got != baseline {
		t.Error("proof metadata map insertion order changed capability revision payload")
	}
}

func capabilityRevisionSemanticObservation(t *testing.T, manifest *CapabilityManifest, evidenceID string) *CapabilitySemanticsObservation {
	t.Helper()
	for i := range manifest.Evidence {
		if manifest.Evidence[i].ID == evidenceID {
			return manifest.Evidence[i].Observations.Semantics
		}
	}
	t.Fatalf("missing evidence %q", evidenceID)
	return nil
}

func mutateCapabilityRevisionField(observation *CapabilitySemanticsObservation, mutate func(*CapabilityFieldExpectation), eligible func(CapabilityFieldExpectation) bool) bool {
	for i := range observation.Lineage {
		for _, state := range []*CapabilityFieldStateExpectation{observation.Lineage[i].Before, observation.Lineage[i].After} {
			if state == nil {
				continue
			}
			for j := range state.Fields {
				if eligible(state.Fields[j]) {
					mutate(&state.Fields[j])
					return true
				}
			}
		}
	}
	return false
}

func capabilityRevisionTestDigest(t *testing.T, manifest CapabilityManifest) [sha256.Size]byte {
	return capabilityRevisionTestDigestForRequirements(t, manifest, milestone11SemanticProofRequirements)
}

func capabilityRevisionTestDigestForRequirements(t *testing.T, manifest CapabilityManifest, requirements map[string]capabilitySemanticProofRequirement) [sha256.Size]byte {
	t.Helper()
	encoded, err := json.Marshal(capabilityRevisionPayloadForRequirements(manifest, requirements))
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(encoded)
}

func TestCapabilityLedgerCoversEveryKindPerLanguage(t *testing.T) {
	records, _, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	for _, language := range []string{"spl", "spl2"} {
		kinds := make(map[string]bool, len(capabilityRecordKinds))
		for _, record := range records {
			if record.Language == language {
				kinds[record.Kind] = true
			}
		}
		for kind := range capabilityRecordKinds {
			if !kinds[kind] {
				t.Errorf("%s ledger has no %s record", language, kind)
			}
		}
	}
}

func TestCapabilityLedgerCoversLegacyInventory(t *testing.T) {
	records, _, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}

	actualNames := func(language, kind string) []string {
		set := map[string]bool{}
		for _, record := range records {
			if record.Language == language && record.Kind == kind {
				set[record.Name] = true
			}
		}
		names := make([]string, 0, len(set))
		for name := range set {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}
	wantSPLCommands := make([]string, 0, len(commands))
	for name := range commands {
		wantSPLCommands = append(wantSPLCommands, name)
	}
	sort.Strings(wantSPLCommands)
	wantSPLFunctions := make([]string, 0, len(functions))
	for name := range functions {
		wantSPLFunctions = append(wantSPLFunctions, name)
	}
	sort.Strings(wantSPLFunctions)
	wantSPL2Commands := make([]string, 0, len(spl2CommandInventory))
	for _, entry := range spl2CommandInventory {
		wantSPL2Commands = append(wantSPL2Commands, entry.name)
	}
	sort.Strings(wantSPL2Commands)
	wantSPL2Functions := make([]string, 0, len(spl2TypedPolicies["splunkd/current"].functions))
	for name := range spl2TypedPolicies["splunkd/current"].functions {
		wantSPL2Functions = append(wantSPL2Functions, name)
	}
	sort.Strings(wantSPL2Functions)

	for _, check := range []struct {
		language string
		kind     string
		want     []string
	}{
		{language: "spl", kind: "command", want: wantSPLCommands},
		{language: "spl", kind: "function", want: wantSPLFunctions},
		{language: "spl2", kind: "command", want: wantSPL2Commands},
		{language: "spl2", kind: "function", want: wantSPL2Functions},
	} {
		if got := actualNames(check.language, check.kind); !slices.Equal(got, check.want) {
			t.Errorf("%s %s names = %v, want %v", check.language, check.kind, got, check.want)
		}
	}

	legacy := []CapabilityManifest{Capabilities()}
	spl2Legacy, err := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatalf("load SPL2 legacy capabilities: %v", err)
	}
	legacy = append(legacy, spl2Legacy)
	for _, manifest := range legacy {
		for kind, capabilities := range map[string][]Capability{
			"command":  manifest.Commands,
			"function": manifest.Functions,
		} {
			for _, capability := range capabilities {
				var grammarRegistered, semanticSupported bool
				var limitations []string
				seenLimitations := map[string]bool{}
				for _, record := range records {
					if record.Language != manifest.Language || record.Kind != kind || record.Name != capability.Name {
						continue
					}
					grammarRegistered = grammarRegistered || record.GrammarRegistered
					semanticSupported = semanticSupported || record.Dimensions.Semantics.State == CapabilitySupported
					for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
						for _, limitation := range dimension.claim.Limitations {
							if !seenLimitations[limitation] {
								seenLimitations[limitation] = true
								limitations = append(limitations, limitation)
							}
						}
					}
				}
				if grammarRegistered != capability.SyntaxSupported || semanticSupported != capability.SemanticSupported {
					t.Errorf("%s %s %s support = grammar:%v semantics:%v, want syntax:%v semantics:%v", manifest.Language, kind, capability.Name, grammarRegistered, semanticSupported, capability.SyntaxSupported, capability.SemanticSupported)
				}
				if !slices.Equal(limitations, capability.Limitations) {
					t.Errorf("%s %s %s limitations = %v, want %v", manifest.Language, kind, capability.Name, limitations, capability.Limitations)
				}
			}
		}
	}
}

func TestCapabilityGrammarRegistrationDoesNotAddSyntaxCoverage(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	evidenceByID := make(map[string]CapabilityEvidence, len(cases))
	for _, evidence := range cases {
		evidenceByID[evidence.ID] = evidence
	}
	for _, record := range records {
		if record.ID != "spl2.command.spl1.quoted-pipeline" {
			continue
		}
		if !record.GrammarRegistered {
			t.Fatal("SPL2 spl1 grammar registration was not authored")
		}
		if record.Dimensions.Syntax.State != CapabilityUnsupported {
			t.Fatalf("SPL2 spl1 syntax state = %q, want %q", record.Dimensions.Syntax.State, CapabilityUnsupported)
		}
		if len(record.Dimensions.Syntax.EvidenceIDs) == 0 {
			t.Fatal("SPL2 spl1 syntax boundary has no evidence")
		}
		for _, evidenceID := range record.Dimensions.Syntax.EvidenceIDs {
			evidence := evidenceByID[evidenceID]
			if evidence.Classification != CapabilityEvidenceIncomplete || evidence.Observations.Syntax == nil || evidence.Observations.Syntax.Complete {
				t.Fatalf("SPL2 spl1 syntax evidence %q is not an honest incomplete observation: %+v", evidenceID, evidence)
			}
		}
		summary := summarizeCapabilityRecords([]CapabilityRecord{record})
		if summary.Syntax.Covered != 0 || summary.Syntax.Unsupported != 1 {
			t.Fatalf("grammar registration added syntax coverage: %+v", summary.Syntax)
		}
		return
	}
	t.Fatal("SPL2 spl1 capability record is missing")
}

func TestEmbeddedCapabilityReviewedFacts(t *testing.T) {
	_, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	evidenceByID := make(map[string]CapabilityEvidence, len(cases))
	for _, evidence := range cases {
		evidenceByID[evidence.ID] = evidence
	}
	evidence := func(id string) CapabilityEvidence {
		t.Helper()
		got, ok := evidenceByID[id]
		if !ok {
			t.Fatalf("capability evidence %q is missing", id)
		}
		return got
	}

	for _, id := range []string{
		"spl.splunkd.baseline.rewrite",
		"spl2.splunkd.baseline.rewrite",
	} {
		got := evidence(id)
		if got.Classification != CapabilityEvidenceNegative || got.Observations.SafeRewriting == nil {
			t.Fatalf("%s is not an authored negative rewrite boundary: %+v", id, got)
		}
		if got.Observations.SafeRewriting.Status != Invalid {
			t.Errorf("%s rewrite status = %q, want %q", id, got.Observations.SafeRewriting.Status, Invalid)
		}
		if len(got.Observations.SafeRewriting.ChangeReasons) != 0 || len(got.Observations.SafeRewriting.RuleEvaluationReasons) != 0 {
			t.Errorf("%s records nonexistent rewrite or rule evaluations: %+v", id, got.Observations.SafeRewriting)
		}
	}
	if got := evidence("spl.splunkd.baseline.rewrite").Observations.SafeRewriting.CoverageReasons; !slices.Equal(got, []string{"SPL_SYNTAX_ERROR", "post_verification_failed"}) {
		t.Errorf("SPL rewrite coverage reasons = %v", got)
	}
	if got := evidence("spl2.splunkd.baseline.rewrite").Observations.SafeRewriting.CoverageReasons; !slices.Equal(got, []string{"SPL_PROFILE_MISMATCH", "SPL_UNSUPPORTED_SEMANTICS", "post_verification_failed"}) {
		t.Errorf("SPL2 rewrite coverage reasons = %v", got)
	}

	for _, id := range []string{
		"spl2.decrypt.profile-mismatch.incomplete",
		"spl2.fillnull.field-list.incomplete",
		"spl2.ocsf.profile-mismatch.incomplete",
		"spl2.route.profile-mismatch.incomplete",
		"spl2.timewrap.span.incomplete",
	} {
		got := evidence(id)
		if got.Observations.Semantics == nil || got.Observations.Semantics.Status != Invalid {
			t.Errorf("%s semantic status is not invalid: %+v", id, got.Observations.Semantics)
		}
	}
	for _, replacement := range []struct {
		positive string
		boundary string
	}{
		{positive: "spl2.if.subpipe.positive", boundary: "spl2.branch.unguarded-arms.incomplete"},
		{positive: "spl2.join.qualified-subsearch.positive", boundary: "spl2.join.output-collision.incomplete"},
	} {
		positive := evidence(replacement.positive)
		if positive.Observations.Semantics == nil || positive.Observations.Semantics.Status != Valid || !positive.Observations.Semantics.Complete {
			t.Errorf("%s is not complete valid replacement evidence: %+v", replacement.positive, positive.Observations.Semantics)
		}
		boundary := evidence(replacement.boundary)
		if boundary.Observations.Semantics == nil || boundary.Observations.Semantics.Status != Incomplete || boundary.Observations.Semantics.Complete {
			t.Errorf("%s is not held-boundary replacement evidence: %+v", replacement.boundary, boundary.Observations.Semantics)
		}
	}

	macro := evidence("spl.macro.exact-invocation.incomplete")
	if got := macro.Observations.Semantics.Stages; len(got) != 1 || got[0].Command != "search" || got[0].SemanticComplete {
		t.Errorf("macro invocation stages = %+v, want incomplete search stage", got)
	}

	for _, check := range []struct {
		id          string
		startOffset int
		startColumn int
		endOffset   int
		endColumn   int
	}{
		{id: "spl.splunkd.baseline.negative", startOffset: 13, startColumn: 14, endOffset: 13, endColumn: 14},
		{id: "spl2.splunkd.baseline.negative", startOffset: 12, startColumn: 13, endOffset: 19, endColumn: 20},
	} {
		got := evidence(check.id)
		for surface, diagnostics := range map[string][]CapabilityDiagnosticExpectation{
			"syntax":    got.Observations.Syntax.Diagnostics,
			"semantics": got.Observations.Semantics.Diagnostics,
		} {
			if len(diagnostics) != 1 {
				t.Errorf("%s %s diagnostics = %+v, want one", check.id, surface, diagnostics)
				continue
			}
			location := diagnostics[0].Location
			if location.Start.Offset != check.startOffset || location.Start.Line != 1 || location.Start.Column != check.startColumn ||
				location.End.Offset != check.endOffset || location.End.Line != 1 || location.End.Column != check.endColumn {
				t.Errorf("%s %s diagnostic location = %+v", check.id, surface, location)
			}
		}
	}

	timewrap := evidence("spl2.timewrap.span.positive")
	if timewrap.Document.Text != "FROM main | timechart count() | timewrap 2day" {
		t.Fatalf("timewrap positive witness = %q", timewrap.Document.Text)
	}
	result, err := Analyze(timewrap.Document)
	if err != nil {
		t.Fatalf("analyze timewrap positive witness: %v", err)
	}
	if !result.Coverage.SyntaxComplete || result.Status == Invalid {
		t.Fatalf("timewrap positive witness status = %q, syntax complete = %v", result.Status, result.Coverage.SyntaxComplete)
	}
}

func TestSPL2Milestone11CapabilityClaimsDoNotBroadenLintingOrRewriting(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatal(err)
	}
	changed := map[string]bool{
		"spl2.command.bin.span-field": true, "spl2.command.eval.exact-assignment": true,
		"spl2.command.fields.exact-field-list": true, "spl2.command.from.dataset": true,
		"spl2.command.if.subpipe": true, "spl2.command.join.qualified-subsearch": true,
		"spl2.command.mvexpand.limited-field": true, "spl2.command.select.projection": true,
		"spl2.command.stats.aggregate-call": true, "spl2.command.union.dataset": true,
		"spl2.command.where.predicate":            true,
		"spl2.dataset.dataset.dynamic-descriptor": true, "spl2.dataset.dataset.parameter": true,
		"spl2.dataset.dataset.static-descriptor":   true,
		"spl2.expression.field.identity-collision": true, "spl2.expression.field.quoted-dotted-atom": true,
		"spl2.expression.field.structural-path": true, "spl2.expression.function-call.invalid-selected-arity": true,
		"spl2.pipeline.branch.guarded-arms": true, "spl2.pipeline.branch.unguarded-arms": true,
		"spl2.pipeline.join.output-collision": true, "spl2.pipeline.join.qualified-left-subsearch": true,
		"spl2.pipeline.join.qualified-outer-subsearch": true,
		"spl2.module.module.declaration-cycle":         true, "spl2.module.module.local-scalar-function": true,
		"spl2.module.module.local-view": true, "spl2.module.module.unresolved-import": true,
	}
	for _, name := range []string{"abs", "avg", "coalesce", "count", "dc", "distinct_count", "lower", "match", "max", "min", "round", "rtrim", "sum", "tonumber", "values"} {
		form := "one-positional"
		if name == "coalesce" || name == "match" {
			form = "two-positional"
		} else if name == "count" {
			form = "zero-positional"
		}
		changed["spl2.function."+name+"."+form] = true
	}
	for _, id := range []string{
		"spl2.function.any.lambda-positional", "spl2.function.cidrmatch.two-positional",
		"spl2.function.json.one-positional", "spl2.function.json_array_to_mv.one-positional",
		"spl2.function.like.two-positional", "spl2.function.mvindex.two-positional",
		"spl2.function.span.grouping-positional", "spl2.function.sqrt.one-positional",
		"spl2.function.stdev.one-positional", "spl2.function.strftime.two-positional",
	} {
		changed[id] = true
	}

	for _, record := range records {
		if !changed[record.ID] {
			continue
		}
		for dimension, claim := range map[string]CapabilityClaim{
			"linting":        record.Dimensions.Linting,
			"safe_rewriting": record.Dimensions.SafeRewriting,
		} {
			if claim.State != CapabilityUnassessed || len(claim.EvidenceIDs) != 0 || len(claim.Limitations) != 0 {
				t.Errorf("%s %s broadened without independent evidence: %+v", record.ID, dimension, claim)
			}
		}
		delete(changed, record.ID)
	}
	if len(changed) != 0 {
		t.Fatalf("Milestone 11 capability records missing from ledger: %v", changed)
	}

	frozen := map[string]bool{
		"spl2.bin.span-field.incomplete": true, "spl2.mvexpand.limited-field.incomplete": true,
		"spl2.union.dataset.incomplete": true,
	}
	for _, evidence := range cases {
		if frozen[evidence.ID] {
			t.Errorf("frozen evidence %q was retained instead of replaced", evidence.ID)
		}
	}
}

func TestMilestone10CapabilityClaimsStayBounded(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	recordByID := make(map[string]CapabilityRecord, len(records))
	for _, record := range records {
		recordByID[record.ID] = record
	}
	evidenceByID := make(map[string]CapabilityEvidence, len(cases))
	for _, evidence := range cases {
		evidenceByID[evidence.ID] = evidence
	}

	supportedIDs := []string{
		"spl.command.bin.exact-field",
		"spl.command.bucket.exact-field",
		"spl.command.fillnull.exact-fields",
		"spl.command.mvexpand.exact-field",
		"spl.command.regex.exact-field",
		"spl.command.rex.named-captures",
		"spl.command.spath.explicit-output",
		"spl.command.tstats.exact-model-dataset",
		"spl.function.earliest.one-positional",
		"spl.function.latest.one-positional",
		"spl.function.like.two-positional",
		"spl.function.mvfind.two-positional",
		"spl.function.mvindex.two-positional",
		"spl.function.now.zero-positional",
		"spl.function.null.zero-positional",
		"spl.function.relative_time.two-positional",
		"spl.function.stdev.one-positional",
		"spl.function.strftime.two-positional",
		"spl.function.true.zero-positional",
	}
	for _, id := range supportedIDs {
		record, ok := recordByID[id]
		if !ok {
			t.Errorf("Milestone 10 record %q is missing", id)
			continue
		}
		for name, claim := range map[string]CapabilityClaim{
			"syntax": record.Dimensions.Syntax, "semantics": record.Dimensions.Semantics, "requirements": record.Dimensions.Requirements,
		} {
			if claim.State != CapabilitySupported || len(claim.EvidenceIDs) != 1 {
				t.Errorf("%s %s claim = %+v, want one supported witness", id, name, claim)
				continue
			}
			if name != "requirements" {
				continue
			}
			evidence := evidenceByID[claim.EvidenceIDs[0]]
			observation := evidence.Observations.Requirements
			if observation == nil || !observation.Complete || len(observation.GapCodes) != 0 {
				t.Errorf("%s supported requirements observation = %+v, want complete evidence with no gaps", id, observation)
				continue
			}
			actual, err := Requirements(evidence.Document)
			if err != nil {
				t.Errorf("%s requirements execution: %v", id, err)
				continue
			}
			actualItems := make([]CapabilityRequirementExpectation, 0, len(actual.Items))
			for _, item := range actual.Items {
				actualItems = append(actualItems, CapabilityRequirementExpectation{
					Kind: item.Kind, Identity: item.Identity, Role: item.Role, Necessity: item.Necessity, Resolution: item.Resolution,
				})
			}
			if !actual.Coverage.Complete || !slices.Equal(observation.Items, actualItems) || len(actual.Gaps) != 0 {
				t.Errorf("%s supported requirements observation = %+v, actual = %+v; want exact item equality and no gaps", id, observation, actual)
			}
		}
	}
	for id, text := range map[string]string{
		"spl.now.zero-positional.positive":  "| eval out=now()",
		"spl.null.zero-positional.positive": "| eval out=null()",
		"spl.true.zero-positional.positive": "| eval out=true()",
	} {
		evidence := evidenceByID[id]
		if evidence.Document.Text != text {
			t.Errorf("%s document = %q, want isolated witness %q", id, evidence.Document.Text, text)
		}
		if got := evidence.Observations.Requirements; got == nil || !got.Complete || len(got.Items) != 0 || len(got.GapCodes) != 0 {
			t.Errorf("%s requirements = %+v, want an exact empty item and gap set", id, got)
		}
	}

	inline := recordByID["spl.command.tstats.inline-macro"]
	if inline.Dimensions.Syntax.State != CapabilitySupported || inline.Dimensions.Semantics.State != CapabilityUnsupported || inline.Dimensions.Requirements.State != CapabilityUnsupported {
		t.Errorf("inline tstats macro claims = %+v, want supported syntax with unsupported semantics and requirements", inline.Dimensions)
	}
	for _, id := range []string{
		"spl.command.append.append-subsearch",
		"spl.command.appendpipe.appendpipe-subsearch",
		"spl.command.join.field-subsearch",
		"spl.command.macro.exact-invocation",
	} {
		if got := recordByID[id].Dimensions.Requirements.State; got != CapabilityUnsupported {
			t.Errorf("%s requirements state = %q, want %q", id, got, CapabilityUnsupported)
		}
	}
	appendpipe := evidenceByID["spl.appendpipe.appendpipe-subsearch.incomplete"]
	wantAppendpipeItems := []CapabilityRequirementExpectation{
		{Kind: "index", Identity: "main", Role: "read", Necessity: "required", Resolution: "exact"},
		{Kind: "field", Identity: "child", Role: "filter", Necessity: "required", Resolution: "exact"},
	}
	if got := appendpipe.Observations.Requirements; got == nil || !slices.Equal(got.Items, wantAppendpipeItems) || !slices.Equal(got.GapCodes, []string{"SPL_UNSUPPORTED_SEMANTICS"}) {
		t.Errorf("appendpipe retained requirements = %+v, want exact parent and child items plus merge gap", got)
	}

	tstats := evidenceByID["spl.tstats.exact-model-dataset.positive"]
	if got := tstats.Observations.Semantics; got == nil || !got.Complete ||
		!slices.Contains(got.Dependencies, CapabilityDependencyExpectation{Kind: "data_model", Name: "Authentication"}) ||
		!slices.Contains(got.Dependencies, CapabilityDependencyExpectation{Kind: "dataset", Name: "Authentication.Authentication"}) ||
		!slices.ContainsFunc(got.Transitions, func(transition CapabilityTransitionExpectation) bool {
			return transition.Operation == "aggregate" && transition.Output == "total"
		}) ||
		!slices.ContainsFunc(got.Transitions, func(transition CapabilityTransitionExpectation) bool {
			return transition.Operation == "aggregate" && transition.Output == "count"
		}) {
		t.Errorf("exact tstats semantic witness lacks reviewed source or aggregate facts: %+v", got)
	}

	spl2Revision, err := capabilityRevisionFor(CapabilityOptions{Language: "spl2", Profile: "splunkd", Version: "current"})
	if err != nil {
		t.Fatal(err)
	}
	const wantSPL2Revision = "sha256:ee254f612293152dc2f6220ead4048c7abc0971382e5de91e3fb94872080b471"
	if spl2Revision != wantSPL2Revision {
		t.Errorf("SPL2 capability revision = %q, want %q", spl2Revision, wantSPL2Revision)
	}
}

func TestCapabilityEvidenceIsReferenced(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	references := make(map[string]int, len(cases))
	for _, record := range records {
		for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
			for _, evidenceID := range dimension.claim.EvidenceIDs {
				references[evidenceID]++
			}
		}
	}
	for _, evidence := range cases {
		if references[evidence.ID] == 0 {
			t.Errorf("evidence %q is not referenced", evidence.ID)
		}
	}
}

func TestDecodeCapabilityAssetsRejectsMalformedInput(t *testing.T) {
	ledger, corpus := validCapabilityAssets(t)

	tests := []struct {
		name   string
		mutate func([]byte, []byte) ([]byte, []byte)
	}{
		{
			name: "unknown ledger field",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return replaceJSON(t, ledger, "schema_version", 1, "unknown", true), corpus
			},
		},
		{
			name: "unknown corpus field",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return ledger, replaceJSON(t, corpus, "schema_version", 1, "unknown", true)
			},
		},
		{
			name: "trailing ledger JSON",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return append(ledger, []byte(` {}`)...), corpus
			},
		},
		{
			name: "trailing corpus JSON",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return ledger, append(corpus, []byte(` {}`)...)
			},
		},
		{
			name: "wrong ledger schema version",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return replaceJSON(t, ledger, "schema_version", 2), corpus
			},
		},
		{
			name: "wrong corpus schema version",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return ledger, replaceJSON(t, corpus, "schema_version", 2)
			},
		},
		{
			name: "duplicate record IDs",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityLedgerFile
				mustUnmarshal(t, ledger, &file)
				file.Records = append(file.Records, file.Records[0])
				return mustJSON(t, file), corpus
			},
		},
		{
			name: "duplicate evidence IDs",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases = append(file.Cases, file.Cases[0])
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "missing record ID",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityLedgerFile
				mustUnmarshal(t, ledger, &file)
				file.Records[0].ID = ""
				return mustJSON(t, file), corpus
			},
		},
		{
			name: "missing evidence ID",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].ID = ""
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "unknown record language",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) { record.Language = "sql" })
			},
		},
		{
			name: "unknown record profile",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) { record.Profile = "cloud" })
			},
		},
		{
			name: "unknown record kind",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) { record.Kind = "clause" })
			},
		},
		{
			name: "invalid grammar registration type",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file map[string]any
				mustUnmarshal(t, ledger, &file)
				records := file["records"].([]any)
				records[0].(map[string]any)["grammar_registered"] = "yes"
				return mustJSON(t, file), corpus
			},
		},
		{
			name: "unknown state",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) {
					record.Dimensions.Syntax.State = "unknown"
				})
			},
		},
		{
			name: "unknown classification",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].Classification = "unknown"
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "unknown record source family",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) {
					record.Provenance.SourceFamily = "blog"
				})
			},
		},
		{
			name: "unknown evidence source family",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].Provenance.SourceFamily = "blog"
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "missing dimensions object",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return deleteNestedJSONField(t, ledger, "records", 0, "dimensions"), corpus
			},
		},
		{
			name: "missing fixed dimension",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return deleteRecordDimension(t, ledger, "safe_rewriting"), corpus
			},
		},
		{
			name: "evidence selector mismatch",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].Document.Language = "spl2"
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "dangling evidence link",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) {
					record.Dimensions.Syntax.EvidenceIDs = []string{"missing"}
				})
			},
		},
		{
			name: "unreferenced evidence",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				extra := cloneCapabilityEvidence(file.Cases[0])
				extra.ID = "unused"
				extra.Document.SourceID = "capability:unused"
				file.Cases = append(file.Cases, extra)
				return ledger, mustJSON(t, file)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			badLedger, badCorpus := tc.mutate(append([]byte(nil), ledger...), append([]byte(nil), corpus...))
			if _, _, err := decodeCapabilityAssets(badLedger, badCorpus); err == nil {
				t.Fatal("malformed capability assets were accepted")
			}
		})
	}
}

func TestDecodeCapabilityAssetsRejectsMalformedObservations(t *testing.T) {
	ledger, corpus := validCapabilityAssets(t)
	tests := []struct {
		name    string
		mutate  func(*CapabilityEvidence)
		wantErr bool
	}{
		{name: "valid representative"},
		{
			name: "positive syntax marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Complete = false
			},
			wantErr: true,
		},
		{
			name: "syntax diagnostic without code",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Code = ""
			},
			wantErr: true,
		},
		{
			name: "semantics with unknown status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Status = "mystery"
			},
			wantErr: true,
		},
		{
			name: "complete semantic and requirements observations with invalid overall status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Status = Invalid
				evidence.Observations.Requirements.QueryStatus = Invalid
			},
		},
		{
			name: "positive semantics marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Complete = false
			},
			wantErr: true,
		},
		{
			name: "semantics without typed facts",
			mutate: func(evidence *CapabilityEvidence) {
				observation := evidence.Observations.Semantics
				observation.Stages = nil
				observation.References = nil
				observation.Dependencies = nil
				observation.Transitions = nil
				observation.Diagnostics = nil
			},
			wantErr: true,
		},
		{
			name: "semantics stage without command",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Stages[0].Command = ""
			},
			wantErr: true,
		},
		{
			name: "positive semantics with incomplete stage",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Stages[0].SemanticComplete = false
			},
			wantErr: true,
		},
		{
			name: "semantics reference without binding",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.References[0].Binding = ""
			},
			wantErr: true,
		},
		{
			name: "semantics reference without location",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.References[0].Location = Location{}
			},
			wantErr: true,
		},
		{
			name: "semantics dependency without name",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Dependencies[0].Name = ""
			},
			wantErr: true,
		},
		{
			name: "semantics transition without output",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Transitions[0].Output = ""
			},
			wantErr: true,
		},
		{
			name: "semantics diagnostic without location",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Diagnostics[0].Location = Location{}
			},
			wantErr: true,
		},
		{
			name: "requirements with unknown status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.QueryStatus = "mystery"
			},
			wantErr: true,
		},
		{
			name: "positive requirements marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.Complete = false
			},
			wantErr: true,
		},
		{
			name: "complete requirements with exact empty collections",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.Items = []CapabilityRequirementExpectation{}
				evidence.Observations.Requirements.GapCodes = []string{}
			},
		},
		{
			name: "requirements missing exact item collection",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.Items = nil
			},
			wantErr: true,
		},
		{
			name: "requirements missing exact gap collection",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.GapCodes = nil
			},
			wantErr: true,
		},
		{
			name: "incomplete requirements without typed content",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Classification = CapabilityEvidenceIncomplete
				evidence.Observations.Requirements.Complete = false
				evidence.Observations.Requirements.QueryStatus = Incomplete
				evidence.Observations.Requirements.Items = []CapabilityRequirementExpectation{}
				evidence.Observations.Requirements.GapCodes = []string{}
			},
			wantErr: true,
		},
		{
			name: "requirement item without identity",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.Items[0].Identity = ""
			},
			wantErr: true,
		},
		{
			name: "blank requirement gap code",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.GapCodes = []string{" "}
			},
			wantErr: true,
		},
		{
			name: "complete requirements with a gap",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.GapCodes = []string{"TEST"}
			},
			wantErr: true,
		},
		{
			name: "lint diagnostic without category",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Linting.Diagnostics[0].Category = ""
			},
			wantErr: true,
		},
		{
			name: "lint diagnostic without location",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Linting.Diagnostics[0].Location = Location{}
			},
			wantErr: true,
		},
		{
			name: "positive linting observation without diagnostics",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Linting.Diagnostics = nil
			},
		},
		{
			name: "diagnostic with unknown severity",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Severity = "mystery"
			},
			wantErr: true,
		},
		{
			name: "diagnostic location beyond document",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Location.End = Position{Offset: 99, Line: 1, Column: 100}
			},
			wantErr: true,
		},
		{
			name: "diagnostic location with incoherent column",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Location.Start.Column = 2
			},
			wantErr: true,
		},
		{
			name: "diagnostic location with reversed range",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Location = Location{
					Start: Position{Offset: 6, Line: 1, Column: 7},
					End:   Position{Offset: 0, Line: 1, Column: 1},
				}
			},
			wantErr: true,
		},
		{
			name: "canonical zero length locations in empty document",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Document.Text = ""
				location := Location{
					Start: Position{Offset: 0, Line: 1, Column: 1},
					End:   Position{Offset: 0, Line: 1, Column: 1},
				}
				for i := range evidence.Observations.Syntax.Diagnostics {
					evidence.Observations.Syntax.Diagnostics[i].Location = location
				}
				for i := range evidence.Observations.Semantics.Diagnostics {
					evidence.Observations.Semantics.Diagnostics[i].Location = location
				}
				for i := range evidence.Observations.Semantics.References {
					evidence.Observations.Semantics.References[i].Location = location
				}
				for i := range evidence.Observations.Linting.Diagnostics {
					evidence.Observations.Linting.Diagnostics[i].Location = location
				}
			},
		},
		{
			name: "safe rewrite with unknown status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.Status = "mystery"
			},
			wantErr: true,
		},
		{
			name: "positive safe rewrite marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.RewriteComplete = false
			},
			wantErr: true,
		},
		{
			name: "complete rewrite observation with incomplete overall status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.Status = Incomplete
			},
		},
		{
			name: "safe rewrite without text",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.Text = ""
			},
			wantErr: true,
		},
		{
			name: "safe rewrite with blank reason",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.CoverageReasons = []string{" "}
			},
			wantErr: true,
		},
		{
			name: "committed rewrite differs from candidate",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.CandidateText = "search index=other"
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidateCorpus := append([]byte(nil), corpus...)
			if tc.mutate != nil {
				candidateCorpus = mutateCapabilityEvidence(t, candidateCorpus, tc.mutate)
			}
			_, _, err := decodeCapabilityAssets(ledger, candidateCorpus)
			if (err != nil) != tc.wantErr {
				t.Fatalf("decodeCapabilityAssets() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestValidateCapabilityClaims(t *testing.T) {
	record := validCapabilityRecord()
	evidence := map[string]CapabilityEvidence{
		"positive":   validCapabilityEvidence("positive", CapabilityEvidencePositive),
		"incomplete": validCapabilityEvidence("incomplete", CapabilityEvidenceIncomplete),
		"negative":   validCapabilityEvidence("negative", CapabilityEvidenceNegative),
	}

	tests := []struct {
		name      string
		dimension string
		claim     CapabilityClaim
		mutate    func(map[string]CapabilityEvidence)
		wantErr   bool
	}{
		{name: "supported", dimension: "syntax", claim: CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"positive"}}},
		{name: "supported without evidence", dimension: "syntax", claim: CapabilityClaim{State: CapabilitySupported}, wantErr: true},
		{name: "supported by negative evidence", dimension: "syntax", claim: CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}}, wantErr: true},
		{name: "partial", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"positive", "incomplete"}, Limitations: []string{"dynamic forms remain incomplete"}}},
		{name: "partial without positive", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"incomplete"}, Limitations: []string{"limited"}}, wantErr: true},
		{name: "partial without incomplete", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"positive"}, Limitations: []string{"limited"}}, wantErr: true},
		{name: "partial without limitation", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"positive", "incomplete"}}, wantErr: true},
		{name: "unsupported by negative", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"negative"}, Limitations: []string{"not modeled"}}},
		{name: "unsupported by incomplete boundary", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"incomplete"}, Limitations: []string{"not modeled"}}},
		{name: "unsupported by positive", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"positive"}, Limitations: []string{"not modeled"}}, wantErr: true},
		{name: "unsupported without limitation", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"negative"}}, wantErr: true},
		{name: "not applicable", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityNotApplicable, Limitations: []string{"rewriting does not apply"}}},
		{name: "not applicable without reason", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityNotApplicable}, wantErr: true},
		{name: "not applicable with evidence", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityNotApplicable, EvidenceIDs: []string{"positive"}, Limitations: []string{"not applicable"}}, wantErr: true},
		{name: "unassessed", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityUnassessed}},
		{name: "unassessed with evidence", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityUnassessed, EvidenceIDs: []string{"positive"}}, wantErr: true},
		{name: "unassessed with limitation", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityUnassessed, Limitations: []string{"unknown"}}, wantErr: true},
		{name: "supported linting by negative case with exact observation", dimension: "linting", claim: CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}}},
		{
			name:      "supported linting by negative case without exact observation",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting = nil
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "supported linting by negative case without diagnostics",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics = nil
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without code",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Code = ""
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without category",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Category = ""
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without severity",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Severity = ""
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without complete location",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Location = Location{}
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "evidence observation belongs to another dimension",
			dimension: "syntax",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"positive"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["positive"]
				item.Observations.Syntax = nil
				evidence["positive"] = item
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cases := cloneEvidenceMap(evidence)
			if tc.mutate != nil {
				tc.mutate(cases)
			}
			err := validateCapabilityClaim(tc.dimension, record, tc.claim, cases)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateCapabilityClaim() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestCapabilitySummaryUsesStrictDenominator(t *testing.T) {
	states := []CapabilityState{
		CapabilitySupported,
		CapabilityPartial,
		CapabilityUnsupported,
		CapabilityNotApplicable,
		CapabilityUnassessed,
	}
	records := make([]CapabilityRecord, 0, len(states))
	for i, state := range states {
		record := validCapabilityRecord()
		record.ID = fmt.Sprintf("record-%d", i)
		claim := CapabilityClaim{State: state}
		record.Dimensions = CapabilityDimensions{
			Syntax:        claim,
			Semantics:     claim,
			Requirements:  claim,
			Linting:       claim,
			SafeRewriting: claim,
		}
		records = append(records, record)
	}

	summary := summarizeCapabilityRecords(records)
	for name, counts := range map[string]CapabilityStateCounts{
		"syntax":         summary.Syntax,
		"semantics":      summary.Semantics,
		"requirements":   summary.Requirements,
		"linting":        summary.Linting,
		"safe_rewriting": summary.SafeRewriting,
	} {
		if counts.Applicable != counts.Supported+counts.Partial+counts.Unsupported+counts.Unassessed {
			t.Errorf("%s applicable denominator is not strict: %+v", name, counts)
		}
		if counts.Covered != counts.Supported {
			t.Errorf("%s covered count differs from supported: %+v", name, counts)
		}
		if len(records) != counts.Applicable+counts.NotApplicable {
			t.Errorf("%s record count identity failed: records=%d counts=%+v", name, len(records), counts)
		}
		if counts.Supported != 1 || counts.Partial != 1 || counts.Unsupported != 1 || counts.NotApplicable != 1 || counts.Unassessed != 1 {
			t.Errorf("%s state counts changed: %+v", name, counts)
		}
	}
}

func TestDecodeCapabilityAssetsRequiresEveryKindPerLanguage(t *testing.T) {
	ledgerJSON, corpusJSON := validCapabilityAssets(t)
	if _, _, err := decodeCapabilityAssets(ledgerJSON, corpusJSON); err != nil {
		t.Fatalf("valid capability assets failed: %v", err)
	}

	for _, language := range []string{"spl", "spl2"} {
		t.Run(language, func(t *testing.T) {
			var ledger capabilityLedgerFile
			mustUnmarshal(t, ledgerJSON, &ledger)
			for i, record := range ledger.Records {
				if record.Language == language && record.Kind == "command" {
					ledger.Records = append(ledger.Records[:i], ledger.Records[i+1:]...)
					break
				}
			}
			if _, _, err := decodeCapabilityAssets(mustJSON(t, ledger), corpusJSON); err == nil {
				t.Fatalf("ledger missing the command kind for %s was accepted", language)
			}
		})
	}

	t.Run("empty bundle", func(t *testing.T) {
		ledger := mustJSON(t, capabilityLedgerFile{SchemaVersion: 1, Records: []CapabilityRecord{}})
		corpus := mustJSON(t, capabilityCorpusFile{SchemaVersion: 1, Cases: []CapabilityEvidence{}})
		if _, _, err := decodeCapabilityAssets(ledger, corpus); err == nil {
			t.Fatal("empty capability bundle was accepted")
		}
	})

	for _, language := range []string{"spl", "spl2"} {
		t.Run(language+" only", func(t *testing.T) {
			var fullLedger capabilityLedgerFile
			mustUnmarshal(t, ledgerJSON, &fullLedger)
			var records []CapabilityRecord
			for _, record := range fullLedger.Records {
				if record.Language == language {
					records = append(records, record)
				}
			}
			var fullCorpus capabilityCorpusFile
			mustUnmarshal(t, corpusJSON, &fullCorpus)
			var cases []CapabilityEvidence
			for _, evidence := range fullCorpus.Cases {
				if evidence.Document.Language == language {
					cases = append(cases, evidence)
				}
			}
			ledger := mustJSON(t, capabilityLedgerFile{SchemaVersion: 1, Records: records})
			corpus := mustJSON(t, capabilityCorpusFile{SchemaVersion: 1, Cases: cases})
			if _, _, err := decodeCapabilityAssets(ledger, corpus); err == nil {
				t.Fatalf("%s-only capability bundle was accepted", language)
			}
		})
	}
}

func TestCapabilityDataCanonicalOrder(t *testing.T) {
	recordOrder := []struct {
		name        string
		left, right CapabilityRecord
	}{
		{
			name:  "language precedes every lower key",
			left:  CapabilityRecord{Language: "spl", Profile: "z", Kind: "z", Name: "z", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl2", Profile: "a", Kind: "a", Name: "a", Form: "a", ID: "a"},
		},
		{
			name:  "profile precedes kind name form and ID",
			left:  CapabilityRecord{Language: "spl", Profile: "a", Kind: "z", Name: "z", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "b", Kind: "a", Name: "a", Form: "a", ID: "a"},
		},
		{
			name:  "kind precedes name form and ID",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "z", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "function", Name: "a", Form: "a", ID: "a"},
		},
		{
			name:  "name precedes form and ID",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "alpha", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "beta", Form: "a", ID: "a"},
		},
		{
			name:  "form precedes ID",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "alpha", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "beta", ID: "a"},
		},
		{
			name:  "ID is the final key",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "basic", ID: "a"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "basic", ID: "b"},
		},
	}
	for _, tc := range recordOrder {
		t.Run(tc.name, func(t *testing.T) {
			if compareCapabilityRecords(tc.left, tc.right) >= 0 || compareCapabilityRecords(tc.right, tc.left) <= 0 {
				t.Fatalf("record comparison ignored %s", tc.name)
			}
		})
	}

	evidenceOrder := []struct {
		name        string
		left, right CapabilityEvidence
	}{
		{
			name:  "document language precedes profile and ID",
			left:  CapabilityEvidence{ID: "z", Document: QueryDocument{Language: "spl", Profile: "z"}},
			right: CapabilityEvidence{ID: "a", Document: QueryDocument{Language: "spl2", Profile: "a"}},
		},
		{
			name:  "document profile precedes ID",
			left:  CapabilityEvidence{ID: "z", Document: QueryDocument{Language: "spl", Profile: "a"}},
			right: CapabilityEvidence{ID: "a", Document: QueryDocument{Language: "spl", Profile: "b"}},
		},
		{
			name:  "evidence ID is the final key",
			left:  CapabilityEvidence{ID: "a", Document: QueryDocument{Language: "spl", Profile: "splunkd"}},
			right: CapabilityEvidence{ID: "b", Document: QueryDocument{Language: "spl", Profile: "splunkd"}},
		},
	}
	for _, tc := range evidenceOrder {
		t.Run(tc.name, func(t *testing.T) {
			if compareCapabilityEvidence(tc.left, tc.right) >= 0 || compareCapabilityEvidence(tc.right, tc.left) <= 0 {
				t.Fatalf("evidence comparison ignored %s", tc.name)
			}
		})
	}

	ledgerJSON, corpusJSON := validCapabilityAssets(t)
	if _, _, err := decodeCapabilityAssets(ledgerJSON, corpusJSON); err != nil {
		t.Fatalf("canonical capability data failed: %v", err)
	}
	var ledger capabilityLedgerFile
	mustUnmarshal(t, ledgerJSON, &ledger)
	ledger.Records[0], ledger.Records[1] = ledger.Records[1], ledger.Records[0]
	if _, _, err := decodeCapabilityAssets(mustJSON(t, ledger), corpusJSON); err == nil {
		t.Fatal("out-of-order capability records were accepted")
	}
	var corpus capabilityCorpusFile
	mustUnmarshal(t, corpusJSON, &corpus)
	corpus.Cases[0], corpus.Cases[1] = corpus.Cases[1], corpus.Cases[0]
	if _, _, err := decodeCapabilityAssets(ledgerJSON, mustJSON(t, corpus)); err == nil {
		t.Fatal("out-of-order capability evidence was accepted")
	}
}

func TestCapabilityDataClonesAreDeep(t *testing.T) {
	record := validCapabilityRecord()
	record.GrammarRegistered = true
	record.Dimensions.Syntax.Limitations = []string{}
	recordClone := cloneCapabilityRecord(record)
	if !recordClone.GrammarRegistered {
		t.Fatal("record clone dropped grammar registration")
	}
	if recordClone.Dimensions.Syntax.Limitations == nil {
		t.Fatal("record clone changed an authored empty slice to nil")
	}
	recordClone.Dimensions.Syntax.EvidenceIDs[0] = "mutated"
	if record.Dimensions.Syntax.EvidenceIDs[0] == "mutated" {
		t.Fatal("record clone aliases evidence IDs")
	}

	evidence := validCapabilityEvidence("positive", CapabilityEvidencePositive)
	evidence.Observations.Semantics = validStructuredCapabilityEvidence().Observations.Semantics
	evidence.Observations.SafeRewriting.ChangeReasons = []string{}
	evidenceClone := cloneCapabilityEvidence(evidence)
	if evidenceClone.Observations.SafeRewriting.ChangeReasons == nil {
		t.Fatal("evidence clone changed an authored empty slice to nil")
	}
	evidenceClone.Observations.Semantics.Stages[0].Command = "mutated"
	evidenceClone.Observations.Semantics.Transitions[0].InputReferenceIDs[0] = "mutated"
	evidenceClone.Observations.Semantics.Lineage[0].After.Fields[0].OriginReferenceIDs[0] = "mutated"
	evidenceClone.Observations.Semantics.FinalFieldState.Removed[0].Name = "mutated"
	evidenceClone.Observations.Requirements.Items[0].Identity = "mutated"
	evidenceClone.RewriteRequest[0] = '['
	if evidence.Observations.Semantics.Stages[0].Command == "mutated" ||
		evidence.Observations.Semantics.Transitions[0].InputReferenceIDs[0] == "mutated" ||
		evidence.Observations.Semantics.Lineage[0].After.Fields[0].OriginReferenceIDs[0] == "mutated" ||
		evidence.Observations.Semantics.FinalFieldState.Removed[0].Name == "mutated" ||
		evidence.Observations.Requirements.Items[0].Identity == "mutated" ||
		evidence.RewriteRequest[0] == '[' {
		t.Fatal("evidence clone aliases nested authored data")
	}
}

func validCapabilityAssets(t *testing.T) ([]byte, []byte) {
	t.Helper()
	return mustJSON(t, capabilityLedgerFile{SchemaVersion: 1, Records: validCapabilityRecords()}),
		mustJSON(t, capabilityCorpusFile{SchemaVersion: 1, Cases: []CapabilityEvidence{
			validCapabilityEvidenceForLanguage("positive-spl", CapabilityEvidencePositive, "spl"),
			validCapabilityEvidenceForLanguage("positive-spl2", CapabilityEvidencePositive, "spl2"),
		}})
}

func embeddedCapabilityCorpusMap(t *testing.T) map[string]any {
	t.Helper()
	var corpus map[string]any
	mustUnmarshal(t, embeddedCapabilityCorpus, &corpus)
	return corpus
}

func semanticEvidenceIDForRecord(t *testing.T, ledger capabilityLedgerFile, recordID string) string {
	t.Helper()
	for _, record := range ledger.Records {
		if record.ID != recordID {
			continue
		}
		if len(record.Dimensions.Semantics.EvidenceIDs) != 1 {
			t.Fatalf("record %q has %d semantic evidence IDs, want 1", recordID, len(record.Dimensions.Semantics.EvidenceIDs))
		}
		return record.Dimensions.Semantics.EvidenceIDs[0]
	}
	t.Fatalf("missing capability record %q", recordID)
	return ""
}

func capabilitySemanticsMap(t *testing.T, corpus map[string]any, evidenceID string) map[string]any {
	t.Helper()
	for _, item := range corpus["cases"].([]any) {
		evidence := item.(map[string]any)
		if evidence["id"] == evidenceID {
			return evidence["observations"].(map[string]any)["semantics"].(map[string]any)
		}
	}
	t.Fatalf("missing semantic evidence %q", evidenceID)
	return nil
}

func capabilityEvidenceErrorOwner(t *testing.T, corpus map[string]any, evidenceID string) string {
	t.Helper()
	for i, item := range corpus["cases"].([]any) {
		if item.(map[string]any)["id"] == evidenceID {
			return fmt.Sprintf("evidence case %d", i)
		}
	}
	t.Fatalf("missing capability evidence %q", evidenceID)
	return ""
}

func visitCapabilityFieldStateMaps(observation map[string]any, visit func(map[string]any)) {
	if lineage, ok := observation["lineage"].([]any); ok {
		for _, item := range lineage {
			entry := item.(map[string]any)
			for _, key := range []string{"before", "after"} {
				if state, ok := entry[key].(map[string]any); ok {
					visit(state)
				}
			}
		}
	}
	if state, ok := observation["final_field_state"].(map[string]any); ok {
		visit(state)
	}
}

func clearCapabilityOriginsMap(observation map[string]any) {
	visitCapabilityFieldStateMaps(observation, func(state map[string]any) {
		for _, item := range state["fields"].([]any) {
			item.(map[string]any)["origin_reference_ids"] = []any{}
		}
	})
}

func clearCapabilityMergeFactsMap(observation map[string]any) {
	visitCapabilityFieldStateMaps(observation, func(state map[string]any) {
		state["uncertain"] = false
		for _, item := range state["fields"].([]any) {
			field := item.(map[string]any)
			field["conditional"] = false
			origins := field["origin_reference_ids"].([]any)
			if len(origins) > 1 {
				field["origin_reference_ids"] = origins[:1]
			}
		}
	})
	for _, item := range observation["transitions"].([]any) {
		item.(map[string]any)["conditional"] = false
	}
}

func validCapabilityRecords() []CapabilityRecord {
	kinds := []string{
		"annotation",
		"command",
		"dataset",
		"expression",
		"function",
		"lexical_form",
		"macro",
		"module",
		"namespace",
		"pipeline",
		"profile_form",
		"subsearch",
		"variable",
	}
	records := make([]CapabilityRecord, 0, len(kinds)*2)
	for _, language := range []string{"spl", "spl2"} {
		for _, kind := range kinds {
			record := validCapabilityRecord()
			record.ID = language + "-" + kind
			record.Language = language
			record.Kind = kind
			record.Name = kind
			record.Form = "default"
			record.Dimensions = claimsReferencing("positive-" + language)
			records = append(records, record)
		}
	}
	return records
}

func validCapabilityRecord() CapabilityRecord {
	return CapabilityRecord{
		ID:       "record",
		Language: "spl",
		Profile:  "splunkd",
		Kind:     "command",
		Name:     "search",
		Form:     "basic",
		Provenance: CapabilityProvenance{
			SourceFamily: "toolkit",
			Reference:    "pkg/analysis",
			Note:         "exercised by the analysis corpus",
		},
		Dimensions: claimsReferencing("positive"),
	}
}

func claimsReferencing(id string) CapabilityDimensions {
	claim := CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{id}}
	return CapabilityDimensions{
		Syntax:        claim,
		Semantics:     claim,
		Requirements:  claim,
		Linting:       claim,
		SafeRewriting: claim,
	}
}

func validCapabilityEvidence(id string, classification CapabilityEvidenceClassification) CapabilityEvidence {
	return validCapabilityEvidenceForLanguage(id, classification, "spl")
}

func validStructuredCapabilityEvidence() CapabilityEvidence {
	evidence := validCapabilityEvidence("structured", CapabilityEvidencePositive)
	semantics := evidence.Observations.Semantics
	semantics.Stages[0].ID = "stage-1"
	semantics.Stages[0].ScopeID = "scope-1"
	semantics.Scopes = []CapabilityScopeExpectation{{
		ID:      "scope-1",
		Kind:    "pipeline",
		StageID: "stage-1",
		Location: Location{
			Start: Position{Offset: 0, Line: 1, Column: 1},
			End:   Position{Offset: 6, Line: 1, Column: 7},
		},
	}}
	semantics.References = []CapabilityReferenceExpectation{
		{ID: "reference-1", NormalizedName: "main", Kind: "index", Role: "read", Resolution: "exact", Binding: "source", Location: semantics.Diagnostics[0].Location},
		{ID: "reference-2", NormalizedName: "result", Kind: "field", Role: "output", Resolution: "exact", Binding: "not_applicable", Location: semantics.Diagnostics[0].Location},
	}
	before := &CapabilityFieldStateExpectation{
		Fields:    []CapabilityFieldExpectation{{Name: "alpha", OriginReferenceIDs: []string{"reference-1"}}},
		Removed:   []CapabilityFieldRemovalExpectation{},
		Open:      true,
		Uncertain: false,
	}
	after := &CapabilityFieldStateExpectation{
		Fields: []CapabilityFieldExpectation{
			{Name: "alpha", OriginReferenceIDs: []string{"reference-1"}},
			{Name: "beta", OriginReferenceIDs: []string{"reference-1", "reference-2"}, Conditional: true},
		},
		Removed:   []CapabilityFieldRemovalExpectation{{Name: "discarded"}},
		Open:      false,
		Uncertain: true,
	}
	semantics.Lineage = []CapabilityLineageExpectation{{StageID: "stage-1", ScopeID: "scope-1", Before: before, After: after}}
	semantics.Transitions = []CapabilityTransitionExpectation{{
		Operation:         "eval",
		Output:            "beta",
		InputReferenceIDs: []string{"reference-1"},
		OutputReferenceID: "reference-2",
		Conditional:       true,
	}}
	semantics.FinalFieldState = cloneCapabilityFieldStateExpectation(after)
	return evidence
}

func validCapabilityEvidenceForLanguage(id string, classification CapabilityEvidenceClassification, language string) CapabilityEvidence {
	diagnostic := CapabilityDiagnosticExpectation{
		Code:     "TEST",
		Category: "test",
		Severity: "warning",
		Location: Location{
			Start: Position{Offset: 0, Line: 1, Column: 1},
			End:   Position{Offset: 6, Line: 1, Column: 7},
		},
	}
	gapCodes := []string{"TEST"}
	if classification == CapabilityEvidencePositive {
		gapCodes = []string{}
	}
	return CapabilityEvidence{
		ID:             id,
		Classification: classification,
		Document: QueryDocument{
			Text:     "search index=main",
			Language: language,
			Profile:  "splunkd",
			Version:  "current",
			SourceID: "capability:" + id,
		},
		Observations: CapabilityEvidenceObservations{
			Syntax: &CapabilitySyntaxObservation{Complete: classification == CapabilityEvidencePositive, Diagnostics: []CapabilityDiagnosticExpectation{diagnostic}},
			Semantics: &CapabilitySemanticsObservation{
				Status:       classificationStatus(classification),
				Complete:     classification == CapabilityEvidencePositive,
				Stages:       []CapabilityStageExpectation{{Command: "search", SemanticComplete: classification == CapabilityEvidencePositive}},
				References:   []CapabilityReferenceExpectation{{NormalizedName: "main", Kind: "index", Role: "read", Resolution: "exact", Binding: "source", Location: diagnostic.Location}},
				Dependencies: []CapabilityDependencyExpectation{{Kind: "index", Name: "main"}},
				Transitions:  []CapabilityTransitionExpectation{{Operation: "read", Output: "main"}},
				Diagnostics:  []CapabilityDiagnosticExpectation{diagnostic},
			},
			Requirements: &CapabilityRequirementsObservation{
				QueryStatus: classificationStatus(classification),
				Complete:    classification == CapabilityEvidencePositive,
				Items:       []CapabilityRequirementExpectation{{Kind: "index", Identity: "main", Role: "read", Necessity: "required", Resolution: "exact"}},
				GapCodes:    gapCodes,
			},
			Linting:       &CapabilityLintingObservation{Diagnostics: []CapabilityDiagnosticExpectation{diagnostic}},
			SafeRewriting: &CapabilityRewriteObservation{Status: classificationStatus(classification), Committed: true, RewriteComplete: classification == CapabilityEvidencePositive, Text: "search index=main", CandidateText: "search index=main", CoverageReasons: []string{"TEST"}, ChangeReasons: []string{"TEST"}, RuleEvaluationReasons: []string{"TEST"}},
		},
		RewriteRequest: json.RawMessage(`{"schema_version":1}`),
		Provenance: CapabilityProvenance{
			SourceFamily: "toolkit",
			Reference:    "pkg/analysis",
			Note:         "expected observation",
		},
	}
}

func classificationStatus(classification CapabilityEvidenceClassification) Status {
	if classification == CapabilityEvidencePositive {
		return Valid
	}
	if classification == CapabilityEvidenceIncomplete {
		return Incomplete
	}
	return Invalid
}

func cloneEvidenceMap(source map[string]CapabilityEvidence) map[string]CapabilityEvidence {
	cloned := make(map[string]CapabilityEvidence, len(source))
	for id, evidence := range source {
		cloned[id] = cloneCapabilityEvidence(evidence)
	}
	return cloned
}

func mutateRecord(t *testing.T, ledger, corpus []byte, mutate func(*CapabilityRecord)) ([]byte, []byte) {
	t.Helper()
	var file capabilityLedgerFile
	mustUnmarshal(t, ledger, &file)
	mutate(&file.Records[0])
	return mustJSON(t, file), corpus
}

func mutateCapabilityEvidence(t *testing.T, corpus []byte, mutate func(*CapabilityEvidence)) []byte {
	t.Helper()
	var file capabilityCorpusFile
	mustUnmarshal(t, corpus, &file)
	mutate(&file.Cases[0])
	return mustJSON(t, file)
}

func replaceJSON(t *testing.T, raw []byte, key string, value any, additions ...any) []byte {
	t.Helper()
	var object map[string]any
	mustUnmarshal(t, raw, &object)
	object[key] = value
	for i := 0; i < len(additions); i += 2 {
		object[additions[i].(string)] = additions[i+1]
	}
	return mustJSON(t, object)
}

func deleteNestedJSONField(t *testing.T, raw []byte, arrayKey string, index int, field string) []byte {
	t.Helper()
	var object map[string]any
	mustUnmarshal(t, raw, &object)
	record := object[arrayKey].([]any)[index].(map[string]any)
	delete(record, field)
	return mustJSON(t, object)
}

func deleteRecordDimension(t *testing.T, raw []byte, dimension string) []byte {
	t.Helper()
	var object map[string]any
	mustUnmarshal(t, raw, &object)
	record := object["records"].([]any)[0].(map[string]any)
	dimensions := record["dimensions"].(map[string]any)
	delete(dimensions, dimension)
	return mustJSON(t, object)
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustUnmarshal(t *testing.T, raw []byte, value any) {
	t.Helper()
	if err := json.Unmarshal(raw, value); err != nil {
		t.Fatal(err)
	}
}
