package analysis

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestSPL2StructuralRequirementTraceParity(t *testing.T) {
	for _, tc := range []struct {
		name, query, identity, original, stage string
		occurrences                            int
		status                                 Status
	}{
		{"navigation", `FROM main | eval x=actor.name`, "actor.name", "actor.name", "stage-1", 1, Valid},
		{"deep navigation", `FROM main | eval x=actor.user.name`, "actor.user.name", "actor.user.name", "stage-1", 1, Valid},
		{"repeated navigation", `FROM main | eval x=actor.name+actor.name`, "actor.name", "actor.name", "stage-1", 2, Valid},
		{"qualified SQL projection", `SELECT actor.name FROM main AS actor`, "name", "actor.name", "stage-0", 1, Incomplete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document := QueryDocument{Text: tc.query, Language: "spl2"}
			result, trace, err := analyzeRewriteWithTrace(document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != tc.status {
				t.Fatalf("status = %q, want %q: %+v", result.Status, tc.status, result.Diagnostics)
			}

			refs := structuralReadReferences(result, tc.identity)
			if len(refs) != tc.occurrences {
				t.Fatalf("structural references = %+v, want %d", refs, tc.occurrences)
			}
			traceByID := requirementTraceReferencesByID(trace)
			for _, ref := range refs {
				entry, ok := traceByID[ref.ID]
				if !ok {
					t.Fatalf("missing trace counterpart for %+v", ref)
				}
				if ref.OriginalName != tc.original || ref.Binding != "source" || entry.reference.Binding != "source" || !entry.directExternal || entry.conditional {
					t.Errorf("structural parity: public=%+v private=%+v", ref, entry)
				}
				if ref.StageID != tc.stage || ref.ScopeID != "scope-0" || entry.reference.StageID != ref.StageID || entry.reference.ScopeID != ref.ScopeID || entry.reference.Location != ref.Location {
					t.Errorf("structural ownership: public=%+v private=%+v", ref, entry.reference)
				}
			}

			item := requirementItem(result.Requirements, "field", tc.identity, "read")
			if item == nil || item.Necessity != "required" || item.Resolution != "exact" || len(item.Occurrences) != len(refs) {
				t.Fatalf("structural requirement item = %+v", item)
			}
			for i, occurrence := range item.Occurrences {
				if occurrence.ReferenceID != refs[i].ID || occurrence.Binding != "source" {
					t.Errorf("occurrence %d = %+v, want %s source", i, occurrence, refs[i].ID)
				}
			}
			if tc.status == Valid && result.Lineage[len(result.Lineage)-1].After.Uncertain {
				t.Errorf("exact structural identity became uncertain: %+v", result.Lineage[len(result.Lineage)-1])
			}
		})
	}
}

func TestSPL2StructuralRequirementDynamicFieldHasNoExactIdentity(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `FROM main | where actor[key]=1`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	ref := spl2Ref(t, result, "actor[]", "read")
	if ref.Resolution != "dynamic" || ref.FieldIdentity != nil {
		t.Fatalf("dynamic reference claimed exact identity: %+v", ref)
	}
	item := requirementItem(result.Requirements, "field", "actor[]", "read")
	if item == nil || item.Resolution != "dynamic" || item.FieldIdentity != nil {
		t.Fatalf("dynamic requirement claimed exact identity: %+v", item)
	}
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"field_identity"`) {
		t.Fatalf("dynamic requirement serialized exact identity: %s", encoded)
	}
}

func TestSPL2StructuralAndQuotedDottedRequirementsStayDistinct(t *testing.T) {
	for _, query := range []string{
		`FROM main | eval x='actor.name'+actor.name`,
		`FROM main | eval x=actor.name+'actor.name'`,
	} {
		t.Run(query, func(t *testing.T) {
			result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			items := []RequirementItem{}
			for _, item := range result.Requirements.Items {
				if item.Kind == "field" && item.Identity == "actor.name" && item.Role == "read" {
					items = append(items, item)
				}
			}
			if len(items) != 2 || items[0].ID != "req-2" || items[1].ID != "req-3" || len(items[0].Occurrences) != 1 || len(items[1].Occurrences) != 1 || items[0].Occurrences[0].Location.Start.Offset >= items[1].Occurrences[0].Location.Start.Offset {
				t.Fatalf("mixed exact requirements must retain first occurrence order: %+v", result.Requirements.Items)
			}
			for _, item := range items {
				if item.Necessity != "required" || len(item.Occurrences) != 1 {
					t.Fatalf("mixed requirement item = %+v", item)
				}
				want := FieldIdentity{Kind: "path", Segments: []string{"actor", "name"}}
				if strings.HasPrefix(item.Occurrences[0].OriginalName, "'") {
					want = FieldIdentity{Kind: "atomic", Segments: []string{"actor.name"}}
				}
				if item.FieldIdentity == nil || !reflect.DeepEqual(*item.FieldIdentity, want) {
					t.Fatalf("mixed requirement identity = %+v, want %+v", item, want)
				}
			}

			refs := append([]Reference{}, structuralReadReferences(result, "actor.name")...)
			if len(refs) != 1 {
				t.Fatalf("structural references = %+v", refs)
			}
			structural := refs[0]
			var quoted *Reference
			for i := range result.References {
				ref := &result.References[i]
				if ref.Kind == "field" && ref.Role == "read" && ref.NormalizedName == "actor.name" && strings.HasPrefix(ref.OriginalName, "'") {
					quoted = ref
				}
			}
			if quoted == nil || quoted.Binding != "source" || structural.Binding != "source" {
				t.Fatalf("mixed bindings: quoted=%+v structural=%+v", quoted, structural)
			}
			for _, ref := range []Reference{*quoted, structural} {
				var traced *requirementTraceReference
				for i := range trace.references {
					if trace.references[i].reference.ID == ref.ID {
						traced = &trace.references[i]
						break
					}
				}
				if traced == nil || traced.reference.FieldIdentity == nil || !reflect.DeepEqual(traced.reference.FieldIdentity, ref.FieldIdentity) {
					t.Fatalf("trace lost exact field identity: trace=%+v public=%+v", traced, ref)
				}
			}
			bindings := map[string]string{quoted.ID: "source", structural.ID: "source"}
			for _, item := range items {
				for _, occurrence := range item.Occurrences {
					if occurrence.Binding != bindings[occurrence.ReferenceID] {
						t.Errorf("mixed occurrence = %+v, expected binding %q", occurrence, bindings[occurrence.ReferenceID])
					}
				}
			}
			if result.Status != Valid || !result.Coverage.SemanticComplete {
				t.Fatalf("mixed exact field analysis = status %q coverage %+v diagnostics %+v", result.Status, result.Coverage, result.Diagnostics)
			}
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Code == CodeAmbiguousField {
					t.Fatalf("distinct exact fields reported as ambiguous: %+v", diagnostic)
				}
			}
		})
	}

	result, _, err := analyzeRewriteWithTrace(QueryDocument{Text: `FROM main | eval x='actor.name'`, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	quoted := spl2Ref(t, result, "actor.name", "read")
	if quoted.OriginalName != "'actor.name'" || quoted.Binding != "source" || requirementItem(result.Requirements, "field", "actor.name", "read").Necessity != "required" {
		t.Fatalf("quoted atomic dotted field changed: %+v %+v", quoted, result.Requirements)
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeAmbiguousField {
			t.Fatalf("quoted atomic field acquired ambiguity diagnostic: %+v", diagnostic)
		}
	}
}

func TestSPL2StructuralRequirementAfterAtomicProjectionIsUnavailable(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `FROM main | fields 'actor.name' | where actor.name=1`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	ref := structuralReadReferences(result, "actor.name")
	if len(ref) != 1 || ref[0].Binding != "unavailable" || result.Status != Invalid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete || result.Requirements.QueryStatus != Invalid {
		t.Fatalf("structural read after atomic projection: refs=%+v analysis=%+v requirements=%+v", ref, result, result.Requirements)
	}
	for _, gap := range result.Requirements.Gaps {
		if gap.Code == CodeRequirementIndeterminate {
			t.Fatalf("exact missing field acquired indeterminate requirement gap: %+v", result.Requirements)
		}
	}
}

func TestSPL2StructuralRequirementAfterAtomicAggregateOutputIsUnavailable(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `FROM main | stats count() AS 'actor.name' | where actor.name=1`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	refs := structuralReadReferences(result, "actor.name")
	if len(refs) != 1 || refs[0].Binding != "unavailable" || result.Status != Invalid || !result.Coverage.SemanticComplete || result.Requirements.QueryStatus != Invalid || !result.Requirements.Coverage.Complete {
		t.Fatalf("structural read after atomic aggregate: refs=%+v analysis=%+v requirements=%+v", refs, result, result.Requirements)
	}
}

func TestSPL2StructuralRequirementCollisionDoesNotHideMissingField(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `FROM main | fields actor.name, 'actor.name' | where missing=1`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != Invalid || !result.Coverage.SemanticComplete || result.Requirements.QueryStatus != Invalid || !result.Requirements.Coverage.Complete {
		t.Fatalf("exact missing field status = analysis %s coverage %+v requirements %+v", result.Status, result.Coverage, result.Requirements)
	}
	ref := spl2Ref(t, result, "missing", "read")
	if ref.Binding != "unavailable" {
		t.Fatalf("missing field binding = %+v", ref)
	}
	if requirementItem(result.Requirements, "field", "missing", "read") != nil {
		t.Fatalf("unavailable field became a direct obligation: %+v", result.Requirements.Items)
	}
	foundUnavailable := false
	for _, diagnostic := range result.Requirements.Diagnostics {
		foundUnavailable = foundUnavailable || diagnostic.Code == CodeUnavailableField
	}
	if !foundUnavailable {
		t.Fatalf("embedded requirements lost unavailable-field diagnostic: %+v", result.Requirements.Diagnostics)
	}
	for _, gap := range result.Requirements.Gaps {
		if gap.Code == CodeRequirementIndeterminate {
			t.Fatalf("exact missing field acquired indeterminate requirement gap: %+v", result.Requirements)
		}
	}
}

func TestSPL2StructuralRequirementBranchIdentityNecessity(t *testing.T) {
	query := `FROM main | branch (flag=true) [where actor.name=1], (flag=false) [where 'actor.name'=1]`
	result, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	items := []RequirementItem{}
	for _, item := range result.Requirements.Items {
		if item.Kind == "field" && item.Identity == "actor.name" && item.Role == "read" {
			items = append(items, item)
		}
	}
	if result.Status != Valid || !result.Requirements.Coverage.Complete || len(items) != 2 {
		t.Fatalf("branch identity requirements = %+v", result.Requirements)
	}
	for _, item := range items {
		if item.Necessity != "conditional" || len(item.Occurrences) != 1 {
			t.Errorf("branch identity requirement = %+v", item)
		}
	}
}

func TestSPL2PipelineJoinQualifiedRequirementOwnership(t *testing.T) {
	query := `FROM main | join max=1 max=2 left=L right=R where L.id=R.uid [FROM other | table uid]`
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	traceByID := requirementTraceReferencesByID(trace)
	for _, identity := range []string{"id", "uid"} {
		ref := spl2Ref(t, result, identity, "read")
		ids = append(ids, ref.ID)
		entry := traceByID[ref.ID]
		if ref.Binding != "source" || entry.reference.Binding != "source" || !entry.directExternal || entry.conditional || ref.StageID != "stage-1" || ref.ScopeID != "scope-0" {
			t.Errorf("qualified join parity for %s: public=%+v private=%+v", identity, ref, entry)
		}
		item := requirementItem(result.Requirements, "field", identity, "read")
		if item == nil || item.Necessity != "required" || len(item.Occurrences) == 0 || item.Occurrences[0].ReferenceID != ref.ID || item.Occurrences[0].Binding != "source" {
			t.Errorf("qualified join item for %s: %+v", identity, item)
		}
	}
	assertRequirementGapMessage(t, result.Requirements.Gaps, CodeUnsupportedSemantics, "Join output merge and qualified input binding are unproved", ids, []string{CodeUnsupportedSemantics})
	assertTraceDiagnosticOwner(t, trace, "Join output merge and qualified input binding are unproved", ids)
	wantGapOrder := []string{CodeUnsupportedSemantics + ":" + strings.Join(ids, ",")}
	gotGapOrder := []string{}
	for _, gap := range result.Requirements.Gaps {
		gotGapOrder = append(gotGapOrder, gap.Code+":"+strings.Join(gap.ReferenceIDs, ","))
	}
	if !reflect.DeepEqual(gotGapOrder, wantGapOrder) {
		t.Errorf("join gap order = %v, want %v", gotGapOrder, wantGapOrder)
	}
	foundParentIdentity := map[string]bool{}
	for _, lineage := range result.Lineage {
		for _, field := range lineage.After.Fields {
			if field.Name == "L.id" || field.Name == "R.uid" {
				t.Errorf("qualifier leaked into public field state: %+v", lineage)
			}
			if lineage.StageID == "stage-1" {
				foundParentIdentity[field.Name] = true
			}
		}
	}
	if !foundParentIdentity["id"] || !foundParentIdentity["uid"] {
		t.Fatalf("qualified identities missing from parent field state: %v", foundParentIdentity)
	}
}

func TestSPL2PipelineJoinRejectsUndeclaredQualifierOwnership(t *testing.T) {
	query := `FROM main | join left=L where L.id=R.id [FROM other]`
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var declared *Reference
	for i := range result.References {
		ref := &result.References[i]
		if ref.OriginalName == "L.id" {
			declared = ref
		}
		if ref.OriginalName == "R.id" && ref.Resolution == "exact" && ref.Binding == "source" {
			t.Fatalf("undeclared qualifier gained exact source ownership: %+v", ref)
		}
	}
	if declared == nil || declared.NormalizedName != "id" || declared.Resolution != "exact" || declared.Binding != "source" {
		t.Fatalf("declared qualifier reference = %+v", declared)
	}
	for _, item := range result.Requirements.Items {
		for _, occurrence := range item.Occurrences {
			if occurrence.OriginalName == "R.id" {
				t.Fatalf("undeclared qualifier gained a requirement: %+v", item)
			}
		}
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeAmbiguousField {
			t.Fatalf("undeclared qualifier caused false ambiguity: %+v", diagnostic)
		}
	}
	assertTraceDiagnosticOwner(t, trace, "Join output merge and qualified input binding are unproved", []string{declared.ID})
}

func TestSPL2SQLJoinPredicateDoesNotGainReferences(t *testing.T) {
	query := `SELECT host FROM main AS L JOIN users AS R ON L.id=R.id`
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, ref := range result.References {
		if ref.NormalizedName == "L.id" || ref.NormalizedName == "R.id" {
			t.Fatalf("SQL join predicate gained a reference: %+v", ref)
		}
	}
	for _, item := range result.Requirements.Items {
		if item.Identity == "L.id" || item.Identity == "R.id" {
			t.Fatalf("SQL join predicate gained a requirement: %+v", item)
		}
	}
	for _, diagnostic := range trace.diagnostics {
		if diagnostic.diagnostic.Message == "SQL join field effects are not yet modeled" && len(diagnostic.pendingReferenceIDs) != 0 {
			t.Fatalf("SQL join limitation gained reference owners: %+v", diagnostic)
		}
	}
	session, err := PrepareRewrite(QueryDocument{Text: query, Language: "spl2"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, site := range session.Evidence().Sites {
		if site.Identity.Name != nil && (*site.Identity.Name == "L.id" || *site.Identity.Name == "R.id") {
			t.Fatalf("SQL join predicate gained a rewrite site: %+v", site)
		}
		if reflect.DeepEqual(site.Identity.Path, []string{"L", "id"}) || reflect.DeepEqual(site.Identity.Path, []string{"R", "id"}) {
			t.Fatalf("SQL join predicate gained a path rewrite site: %+v", site)
		}
	}
}

func TestSPL2StructuralRequirementsConditionalPaths(t *testing.T) {
	for _, tc := range []struct {
		name, query string
	}{
		{
			name:  "if",
			query: `FROM synthetic_events | if (synthetic_guard=true) [where synthetic_left>0 AND synthetic_shared>0] else [where synthetic_right>0 AND synthetic_shared>1]`,
		},
		{
			name:  "branch",
			query: `FROM synthetic_events | branch (synthetic_guard=true) [where synthetic_left>0 AND synthetic_shared>0], (synthetic_guard=false) [where synthetic_right>0 AND synthetic_shared>1]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: tc.query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Valid || !result.Requirements.Coverage.Complete {
				t.Fatalf("path requirement coverage = status %s requirements %+v diagnostics %+v", result.Status, result.Requirements, result.Diagnostics)
			}
			trace.assertReferences(result.References)
			for _, identity := range []string{"synthetic_guard", "synthetic_shared"} {
				item := requirementItem(result.Requirements, "field", identity, "read")
				if item == nil || item.Necessity != "required" {
					t.Errorf("%s item = %+v", identity, item)
				}
			}
			for _, identity := range []string{"synthetic_left", "synthetic_right"} {
				item := requirementItem(result.Requirements, "field", identity, "read")
				if item == nil || item.Necessity != "conditional" || len(item.Occurrences) != 1 {
					t.Errorf("%s item = %+v", identity, item)
				}
			}
		})
	}
}

func TestSPL2StructuralRequirementsUnionDatasetPaths(t *testing.T) {
	query := `union synthetic_left, $synthetic_right, [FROM synthetic_child | where synthetic_value>0]`
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != Valid || !result.Requirements.Coverage.Complete {
		t.Fatalf("union requirement coverage = status %s requirements %+v diagnostics %+v", result.Status, result.Requirements, result.Diagnostics)
	}
	trace.assertReferences(result.References)
	for _, identity := range []string{"synthetic_left", "$synthetic_right", "synthetic_child"} {
		item := requirementItem(result.Requirements, "dataset", identity, "read")
		if item == nil || item.Necessity != "conditional" || len(item.Occurrences) != 1 {
			t.Errorf("union dataset %s = %+v", identity, item)
		}
	}
	value := requirementItem(result.Requirements, "field", "synthetic_value", "read")
	if value == nil || value.Necessity != "conditional" {
		t.Errorf("union child field = %+v", value)
	}
}

func TestSPL2StructuralRequirementsPipelineJoinSelected(t *testing.T) {
	for _, joinType := range []string{"inner", "left", "outer"} {
		t.Run(joinType, func(t *testing.T) {
			query := `FROM synthetic_left | join type=` + joinType + ` left=L right=R where L.synthetic_id=R.synthetic_uid [FROM synthetic_right | fields synthetic_uid, synthetic_name]`
			result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Valid || !result.Requirements.Coverage.Complete {
				t.Fatalf("join requirement coverage = status %s requirements %+v diagnostics %+v", result.Status, result.Requirements, result.Diagnostics)
			}
			trace.assertReferences(result.References)
			for _, identity := range []string{"synthetic_id", "synthetic_uid"} {
				item := requirementItem(result.Requirements, "field", identity, "read")
				if item == nil || item.Necessity != "required" || item.FieldIdentity == nil || item.FieldIdentity.Kind != "atomic" || !reflect.DeepEqual(item.FieldIdentity.Segments, []string{identity}) {
					t.Errorf("join key %s = %+v", identity, item)
				}
			}
			for _, identity := range []string{"synthetic_left", "synthetic_right"} {
				item := requirementItem(result.Requirements, "dataset", identity, "read")
				if item == nil || item.Necessity != "required" {
					t.Errorf("join dataset %s = %+v", identity, item)
				}
			}
		})
	}
}

func TestSPL2StructuralRequirementsJoinSideBindingIgnoresPublicRefinement(t *testing.T) {
	query := `FROM main | join type=inner left=L right=R where L.id=R.uid [FROM other]`
	plain := spl2AnalyzeTest(t, query)
	refined, err := AnalyzeWithSourceUniverse(QueryDocument{Text: query, Language: "spl2"}, SourceUniverse{
		Fields:   []string{"id", "uid"},
		Complete: true,
		Resolve:  func(string) SourceFieldAdmission { return SourceFieldAdmitted },
	})
	if err != nil {
		t.Fatal(err)
	}
	if plain.Status != Valid || !plain.Requirements.Coverage.Complete {
		t.Fatalf("plain join requirements = %+v", plain)
	}
	if refined.Result.Status != Valid || !refined.Result.Coverage.SemanticComplete || !refined.Result.Requirements.Coverage.Complete {
		t.Fatalf("refined join requirements = %+v", refined.Result)
	}
	for _, identity := range []string{"id", "uid"} {
		plainItem := requirementItem(plain.Requirements, "field", identity, "read")
		refinedItem := requirementItem(refined.Result.Requirements, "field", identity, "read")
		if plainItem == nil || refinedItem == nil || plainItem.Necessity != "required" || refinedItem.Necessity != "required" || len(plainItem.Occurrences) != 1 || len(refinedItem.Occurrences) != 1 {
			t.Fatalf("join requirement %s changed under refinement: plain=%+v refined=%+v", identity, plainItem, refinedItem)
		}
	}
}

func TestSPLStructuralRemediationDoesNotChangeAtomicDottedField(t *testing.T) {
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: `search 'actor.name'=value`}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ref := spl2Ref(t, result, "actor.name", "filter")
	entry := requirementTraceReferencesByID(trace)[ref.ID]
	if ref.Binding != "source" || entry.reference.Binding != "source" || !entry.directExternal || entry.conditional {
		t.Fatalf("SPL atomic dotted field changed: public=%+v private=%+v", ref, entry)
	}
}

func structuralReadReferences(result *Result, identity string) []Reference {
	refs := []Reference{}
	for _, ref := range result.References {
		if ref.Kind == "field" && ref.Role == "read" && ref.NormalizedName == identity && !strings.HasPrefix(ref.OriginalName, "'") {
			refs = append(refs, ref)
		}
	}
	sort.SliceStable(refs, func(i, j int) bool { return refs[i].Location.Start.Offset < refs[j].Location.Start.Offset })
	return refs
}

func requirementTraceReferencesByID(trace *requirementTrace) map[string]requirementTraceReference {
	entries := map[string]requirementTraceReference{}
	for _, entry := range trace.references {
		entries[entry.reference.ID] = entry
	}
	return entries
}

func requirementItem(set RequirementSet, kind, identity, role string) *RequirementItem {
	for i := range set.Items {
		item := &set.Items[i]
		if item.Kind == kind && item.Identity == identity && item.Role == role {
			return item
		}
	}
	return nil
}

func assertRequirementGap(t *testing.T, gaps []RequirementGap, code string, referenceIDs, diagnosticCodes []string) {
	t.Helper()
	for _, gap := range gaps {
		if gap.Code == code && reflect.DeepEqual(gap.ReferenceIDs, referenceIDs) && reflect.DeepEqual(gap.DiagnosticCodes, diagnosticCodes) {
			return
		}
	}
	t.Errorf("missing gap code=%s refs=%v diagnostics=%v in %+v", code, referenceIDs, diagnosticCodes, gaps)
}

func assertRequirementGapMessage(t *testing.T, gaps []RequirementGap, code, message string, referenceIDs, diagnosticCodes []string) {
	t.Helper()
	for _, gap := range gaps {
		if gap.Code == code && gap.Message == message && reflect.DeepEqual(gap.ReferenceIDs, referenceIDs) && reflect.DeepEqual(gap.DiagnosticCodes, diagnosticCodes) {
			return
		}
	}
	t.Errorf("missing gap code=%s message=%q refs=%v diagnostics=%v in %+v", code, message, referenceIDs, diagnosticCodes, gaps)
}

func assertTraceDiagnosticOwner(t *testing.T, trace *requirementTrace, message string, referenceIDs []string) {
	t.Helper()
	for _, diagnostic := range trace.diagnostics {
		if diagnostic.diagnostic.Message == message && reflect.DeepEqual(diagnostic.pendingReferenceIDs, referenceIDs) {
			return
		}
	}
	t.Errorf("missing trace diagnostic %q owned by %v in %+v", message, referenceIDs, trace.diagnostics)
}
