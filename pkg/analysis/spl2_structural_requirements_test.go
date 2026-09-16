package analysis

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

const typedNavigationDiagnostic = "Typed navigation or alias binding is not represented by the string-only source universe"

func TestSPL2StructuralReferenceBindingCopiesOrigins(t *testing.T) {
	document := QueryDocument{Text: "actor.name", Language: "spl2", Profile: "splunkd", Version: "current"}
	result := newResult(document)
	result.Stages = []Stage{{ID: "stage-0", ScopeID: "scope-0", SemanticComplete: true}}
	result.rewrite = &RewriteSession{sites: []*rewriteSite{}}
	trace := newRequirementTrace()
	stage := &spl2SemanticStage{semanticStage: &semanticStage{result: result, env: newEnvironmentWithRequirementTrace(trace)}, aliases: map[string]bool{}}
	operand := locatedOperand{
		Name:       "actor.name",
		Location:   Location{Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 10, Line: 1, Column: 11}},
		Resolution: "exact",
		Sound:      true,
		rewrite:    rewriteOwner{role: "navigation", identity: RewriteIdentity{Path: []string{"actor", "name"}}},
	}
	id := stage.referenceAt(operand.Location, operand.Name, "field", "read", operand.Resolution)
	result.References[len(result.References)-1].OriginReferenceIDs = []string{"pending-origin"}
	stage.bindStructuralFieldReference(id, operand)

	public := &result.References[len(result.References)-1]
	private := trace.reference(id)
	if public.Binding != "indeterminate" || private.reference.Binding != "indeterminate" || private.directExternal || !private.conditional || !reflect.DeepEqual(private.reference.OriginReferenceIDs, []string{"pending-origin"}) {
		t.Fatalf("structural origin binding: public=%+v private=%+v", public, private)
	}
	if len(result.rewrite.sites) != 1 || result.rewrite.sites[0].binding != "indeterminate" || !reflect.DeepEqual(result.rewrite.sites[0].inputs, []string{"pending-origin"}) {
		t.Fatalf("rewrite origin binding: %+v", result.rewrite.sites)
	}
	public.OriginReferenceIDs[0] = "changed-public"
	if !reflect.DeepEqual(private.reference.OriginReferenceIDs, []string{"pending-origin"}) || !reflect.DeepEqual(result.rewrite.sites[0].inputs, []string{"pending-origin"}) {
		t.Fatalf("public origins alias private consumers: private=%v rewrite=%v", private.reference.OriginReferenceIDs, result.rewrite.sites[0].inputs)
	}
	private.reference.OriginReferenceIDs[0] = "changed-private"
	if !reflect.DeepEqual(result.rewrite.sites[0].inputs, []string{"pending-origin"}) {
		t.Fatalf("trace origins alias rewrite inputs: %v", result.rewrite.sites[0].inputs)
	}
}

func TestSPL2StructuralRequirementTraceParity(t *testing.T) {
	for _, tc := range []struct {
		name, query, identity, stage string
		occurrences                  int
	}{
		{"navigation", `FROM main | eval x=actor.name`, "actor.name", "stage-1", 1},
		{"deep navigation", `FROM main | eval x=actor.user.name`, "actor.user.name", "stage-1", 1},
		{"repeated navigation", `FROM main | eval x=actor.name+actor.name`, "actor.name", "stage-1", 2},
		{"qualified SQL projection", `SELECT actor.name FROM main AS actor`, "actor.name", "stage-0", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document := QueryDocument{Text: tc.query, Language: "spl2"}
			result, trace, err := analyzeRewriteWithTrace(document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Incomplete || result.Requirements.QueryStatus != Incomplete {
				t.Fatalf("status = analysis %q requirements %q, want incomplete/incomplete", result.Status, result.Requirements.QueryStatus)
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
				if ref.Binding != "indeterminate" || entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional {
					t.Errorf("structural parity: public=%+v private=%+v", ref, entry)
				}
				if ref.StageID != tc.stage || ref.ScopeID != "scope-0" || entry.reference.StageID != ref.StageID || entry.reference.ScopeID != ref.ScopeID || entry.reference.Location != ref.Location {
					t.Errorf("structural ownership: public=%+v private=%+v", ref, entry.reference)
				}
				if !reflect.DeepEqual(entry.reference.OriginReferenceIDs, ref.OriginReferenceIDs) {
					t.Errorf("structural origins differ: public=%v private=%v", ref.OriginReferenceIDs, entry.reference.OriginReferenceIDs)
				}
				assertRequirementGap(t, result.Requirements.Gaps, CodeRequirementIndeterminate, []string{ref.ID}, []string{})
				assertRequirementGapMessage(t, result.Requirements.Gaps, CodeUnsupportedSemantics, typedNavigationDiagnostic, []string{ref.ID}, []string{CodeUnsupportedSemantics})
				assertTraceDiagnosticOwner(t, trace, typedNavigationDiagnostic, []string{ref.ID})
			}

			item := requirementItem(result.Requirements, "field", tc.identity, "read")
			if item == nil || item.Necessity != "conditional" || item.Resolution != "exact" || len(item.Occurrences) != len(refs) {
				t.Fatalf("structural requirement item = %+v", item)
			}
			for i, occurrence := range item.Occurrences {
				if occurrence.ReferenceID != refs[i].ID || occurrence.Binding != "indeterminate" {
					t.Errorf("occurrence %d = %+v, want %s indeterminate", i, occurrence, refs[i].ID)
				}
			}
			gotGapOrder := []string{}
			for _, gap := range result.Requirements.Gaps {
				if gap.Code == CodeRequirementIndeterminate || gap.Code == CodeUnsupportedSemantics && gap.Message == typedNavigationDiagnostic {
					gotGapOrder = append(gotGapOrder, gap.Code+":"+strings.Join(gap.ReferenceIDs, ","))
				}
			}
			wantGapOrder := []string{}
			for _, ref := range refs {
				wantGapOrder = append(wantGapOrder, CodeRequirementIndeterminate+":"+ref.ID, CodeUnsupportedSemantics+":"+ref.ID)
			}
			if !reflect.DeepEqual(gotGapOrder, wantGapOrder) {
				t.Errorf("structural gap order = %v, want %v", gotGapOrder, wantGapOrder)
			}
			for _, lineage := range result.Lineage {
				for _, field := range lineage.After.Fields {
					if field.Name == tc.identity {
						t.Errorf("structural identity persisted in the ordinary field map: %+v", lineage)
					}
				}
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
			if quoted == nil || quoted.Binding != "source" || structural.Binding != "indeterminate" {
				t.Fatalf("mixed bindings: quoted=%+v structural=%+v", quoted, structural)
			}
			bindings := map[string]string{quoted.ID: "source", structural.ID: "indeterminate"}
			for _, occurrence := range item.Occurrences {
				if occurrence.Binding != bindings[occurrence.ReferenceID] {
					t.Errorf("mixed occurrence = %+v, expected binding %q", occurrence, bindings[occurrence.ReferenceID])
				}
			}
			assertRequirementGap(t, result.Requirements.Gaps, CodeRequirementIndeterminate, []string{structural.ID}, []string{})
			assertRequirementGapMessage(t, result.Requirements.Gaps, CodeUnsupportedSemantics, typedNavigationDiagnostic, []string{structural.ID}, []string{CodeUnsupportedSemantics})
			assertTraceDiagnosticOwner(t, trace, typedNavigationDiagnostic, []string{structural.ID})
			for _, gap := range result.Requirements.Gaps {
				if reflect.DeepEqual(gap.ReferenceIDs, []string{quoted.ID}) {
					t.Errorf("quoted atomic field acquired a structural gap: %+v", gap)
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
		if diagnostic.Message == typedNavigationDiagnostic {
			t.Fatalf("quoted atomic field acquired typed-navigation diagnostic: %+v", diagnostic)
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
	for _, name := range []string{"L.id", "R.uid"} {
		ref := spl2Ref(t, result, name, "read")
		ids = append(ids, ref.ID)
		entry := traceByID[ref.ID]
		if ref.Binding != "indeterminate" || entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional || ref.StageID != "stage-1" || ref.ScopeID != "scope-0" {
			t.Errorf("qualified join parity for %s: public=%+v private=%+v", name, ref, entry)
		}
		item := requirementItem(result.Requirements, "field", name, "read")
		if item == nil || item.Necessity != "conditional" || len(item.Occurrences) != 1 || item.Occurrences[0].Binding != "indeterminate" {
			t.Errorf("qualified join item for %s: %+v", name, item)
		}
		assertRequirementGap(t, result.Requirements.Gaps, CodeRequirementIndeterminate, []string{ref.ID}, []string{})
	}
	assertRequirementGapMessage(t, result.Requirements.Gaps, CodeUnsupportedSemantics, "Join output merge and qualified input binding are unproved", ids, []string{CodeUnsupportedSemantics})
	assertTraceDiagnosticOwner(t, trace, "Join output merge and qualified input binding are unproved", ids)
	wantGapOrder := []string{CodeRequirementIndeterminate + ":" + ids[0], CodeRequirementIndeterminate + ":" + ids[1], CodeUnsupportedSemantics + ":" + strings.Join(ids, ",")}
	gotGapOrder := []string{}
	for _, gap := range result.Requirements.Gaps {
		gotGapOrder = append(gotGapOrder, gap.Code+":"+strings.Join(gap.ReferenceIDs, ","))
	}
	if !reflect.DeepEqual(gotGapOrder, wantGapOrder) {
		t.Errorf("join gap order = %v, want %v", gotGapOrder, wantGapOrder)
	}
	for _, lineage := range result.Lineage {
		for _, field := range lineage.After.Fields {
			if field.Name == "L.id" || field.Name == "R.uid" {
				t.Errorf("qualified join operand persisted in ordinary fields: %+v", lineage)
			}
		}
	}
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
