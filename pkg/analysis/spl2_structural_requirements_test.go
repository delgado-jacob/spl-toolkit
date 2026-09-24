package analysis

import (
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
			item := requirementItem(result.Requirements, "field", "actor.name", "read")
			if item == nil || item.Necessity != "required" || len(item.Occurrences) != 2 {
				t.Fatalf("mixed requirement item = %+v", item)
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
			bindings := map[string]string{quoted.ID: "source", structural.ID: "source"}
			for _, occurrence := range item.Occurrences {
				if occurrence.Binding != bindings[occurrence.ReferenceID] {
					t.Errorf("mixed occurrence = %+v, expected binding %q", occurrence, bindings[occurrence.ReferenceID])
				}
			}
			wantOwners := []string{quoted.ID, structural.ID}
			if structural.Location.Start.Offset < quoted.Location.Start.Offset {
				wantOwners = []string{structural.ID, quoted.ID}
			}
			assertRequirementGap(t, result.Requirements.Gaps, CodeAmbiguousField, wantOwners, []string{CodeAmbiguousField})
			assertTraceDiagnosticOwner(t, trace, result.Diagnostics[0].Message, wantOwners)
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

func TestSPL2PipelineJoinQualifiedRequirementOwnership(t *testing.T) {
	query := `FROM main | join left=L right=R where L.id=R.uid [FROM other | table uid]`
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
