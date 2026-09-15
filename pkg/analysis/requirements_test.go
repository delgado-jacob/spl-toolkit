package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestProjectRequirementsGroupsAndOrdersOccurrences(t *testing.T) {
	document := QueryDocument{Text: "host host host", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "query.spl"}
	trace := newRequirementTrace()
	for _, reference := range []Reference{
		{ID: "ref-0", OriginalName: "host", NormalizedName: "host", Kind: "field", Role: "read", StageID: "stage-0", ScopeID: "scope-0", Location: testRequirementLocation(0, 4), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
		{ID: "ref-1", OriginalName: "host", NormalizedName: "host", Kind: "field", Role: "read", StageID: "stage-1", ScopeID: "scope-0", Location: testRequirementLocation(5, 9), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
		{ID: "ref-2", OriginalName: "host", NormalizedName: "host", Kind: "field", Role: "filter", StageID: "stage-2", ScopeID: "scope-0", Location: testRequirementLocation(10, 14), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
	} {
		trace.recordReference(reference, true, false, trace.nextEvent())
	}

	got, err := projectRequirements(document, trace)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items = %+v, want two grouped requirements", got.Items)
	}
	if got.Items[0].ID != "req-1" || got.Items[0].Role != "read" || len(got.Items[0].Occurrences) != 2 {
		t.Fatalf("first item = %+v, want grouped reads", got.Items[0])
	}
	if got.Items[1].ID != "req-2" || got.Items[1].Role != "filter" || len(got.Items[1].Occurrences) != 1 {
		t.Fatalf("second item = %+v, want filter", got.Items[1])
	}
	if got.Items[0].Occurrences[0].ReferenceID != "ref-0" || got.Items[0].Occurrences[1].ReferenceID != "ref-1" {
		t.Fatalf("occurrences are not in canonical reference order: %+v", got.Items[0].Occurrences)
	}
	if got.Query.SourceID != document.SourceID || got.Query.QueryDigest != queryDigest(document.Text) {
		t.Fatalf("query identity = %+v", got.Query)
	}
	if got.SchemaVersion != 1 || got.QueryStatus != Valid || !got.Coverage.Complete {
		t.Fatalf("set metadata = %+v", got)
	}
	if got.Coverage.Reasons == nil || got.Gaps == nil || got.Diagnostics == nil {
		t.Fatalf("empty collections must be arrays: %+v", got)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	if _, found := wire["document"]; found {
		t.Fatalf("requirement set leaked full query text: %s", encoded)
	}
}

func testRequirementLocation(start, end int) Location {
	return Location{Start: Position{Offset: start, Line: 1, Column: start + 1}, End: Position{Offset: end, Line: 1, Column: end + 1}}
}

func TestProjectRequirementsFieldPolicy(t *testing.T) {
	for _, tc := range []struct {
		name        string
		role        string
		binding     string
		resolution  string
		direct      bool
		conditional bool
		wantItem    bool
		wantNeed    string
		wantGap     string
	}{
		{name: "exact source", role: "read", binding: "source", resolution: "exact", direct: true, wantItem: true, wantNeed: "required"},
		{name: "indeterminate source", role: "read", binding: "indeterminate", resolution: "exact", conditional: true, wantItem: true, wantNeed: "conditional", wantGap: CodeRequirementIndeterminate},
		{name: "wildcard", role: "read", binding: "indeterminate", resolution: "wildcard", conditional: true, wantItem: true, wantNeed: "conditional", wantGap: CodeRequirementDynamic},
		{name: "dynamic", role: "read", binding: "indeterminate", resolution: "dynamic", conditional: true, wantItem: true, wantNeed: "conditional", wantGap: CodeRequirementDynamic},
		{name: "derived", role: "read", binding: "derived", resolution: "exact"},
		{name: "create", role: "create", binding: "definition", resolution: "exact"},
		{name: "output", role: "output", binding: "definition", resolution: "exact"},
		{name: "rename source", role: "read", binding: "source", resolution: "exact", direct: true, wantItem: true, wantNeed: "required"},
		{name: "rename target", role: "rename", binding: "definition", resolution: "exact"},
		{name: "remove", role: "remove", binding: "not_applicable", resolution: "exact"},
		{name: "null test", role: "null_test", binding: "source", resolution: "exact"},
		{name: "unavailable local", role: "read", binding: "unavailable", resolution: "exact"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace := newRequirementTrace()
			reference := Reference{
				ID:                 "ref-0",
				OriginalName:       "host",
				NormalizedName:     "host",
				Kind:               "field",
				Role:               tc.role,
				StageID:            "stage-0",
				ScopeID:            "scope-0",
				Location:           testRequirementLocation(0, 4),
				Resolution:         tc.resolution,
				Binding:            tc.binding,
				OriginReferenceIDs: []string{},
			}
			trace.recordReference(reference, tc.direct, tc.conditional, trace.nextEvent())
			got, err := projectRequirements(QueryDocument{Text: "host", Language: "spl", Profile: "splunkd", Version: "current"}, trace)
			if err != nil {
				t.Fatal(err)
			}
			if (len(got.Items) == 1) != tc.wantItem {
				t.Fatalf("items = %+v, want item %t", got.Items, tc.wantItem)
			}
			if tc.wantItem {
				item := got.Items[0]
				if item.Necessity != tc.wantNeed || item.Resolution != tc.resolution || len(item.Occurrences) != 1 || item.Occurrences[0].Binding != tc.binding {
					t.Fatalf("item = %+v", item)
				}
			}
			if tc.wantGap == "" {
				if len(got.Gaps) != 0 || !got.Coverage.Complete {
					t.Fatalf("unexpected gap or incomplete coverage: %+v", got)
				}
			} else {
				if len(got.Gaps) != 1 || got.Gaps[0].Code != tc.wantGap || !reflect.DeepEqual(got.Gaps[0].ReferenceIDs, []string{"ref-0"}) {
					t.Fatalf("gaps = %+v, want %s owned by ref-0", got.Gaps, tc.wantGap)
				}
				if got.Coverage.Complete || !reflect.DeepEqual(got.Coverage.Reasons, []string{tc.wantGap}) {
					t.Fatalf("coverage = %+v", got.Coverage)
				}
			}
		})
	}
}

func TestProjectRequirementsKnowledgePolicy(t *testing.T) {
	for i, kind := range []string{"index", "source", "sourcetype", "dataset", "data_model", "lookup", "macro"} {
		t.Run(kind+" exact", func(t *testing.T) {
			trace := newRequirementTrace()
			reference := testRequirementReference("ref-0", kind, kind+"-name", "read", "not_applicable", "exact", i)
			trace.recordReference(reference, true, false, trace.nextEvent())
			got := mustProjectRequirements(t, trace)
			if len(got.Items) != 1 || got.Items[0].Kind != kind || got.Items[0].Identity != kind+"-name" || got.Items[0].Necessity != "required" || len(got.Gaps) != 0 {
				t.Fatalf("set = %+v", got)
			}
		})
	}

	t.Run("owned wildcard diagnostic", func(t *testing.T) {
		trace := newRequirementTrace()
		reference := testRequirementReference("ref-0", "index", "ma*", "read", "not_applicable", "wildcard", 0)
		trace.recordReference(reference, false, true, trace.nextEvent())
		trace.recordDiagnostic(Diagnostic{Code: CodeUnresolvedWildcard, Severity: "warning", Category: "unsupported_semantics", Message: "wildcard index membership is unresolved", Location: reference.Location, StageID: reference.StageID, ScopeID: reference.ScopeID}, true, []string{"ref-0"}, trace.nextEvent())
		got := mustProjectRequirements(t, trace)
		if len(got.Items) != 1 || got.Items[0].Necessity != "conditional" || got.Items[0].Resolution != "wildcard" {
			t.Fatalf("items = %+v", got.Items)
		}
		if len(got.Gaps) != 1 || got.Gaps[0].Code != CodeUnresolvedWildcard || !reflect.DeepEqual(got.Gaps[0].DiagnosticCodes, []string{CodeUnresolvedWildcard}) {
			t.Fatalf("gaps = %+v", got.Gaps)
		}
	})

	t.Run("defensible dynamic identity", func(t *testing.T) {
		trace := newRequirementTrace()
		reference := testRequirementReference("ref-0", "lookup", "$lookup$", "read", "not_applicable", "dynamic", 0)
		trace.recordReference(reference, false, true, trace.nextEvent())
		got := mustProjectRequirements(t, trace)
		if len(got.Items) != 1 || got.Items[0].Identity != "$lookup$" || len(got.Gaps) != 1 || got.Gaps[0].Code != CodeRequirementDynamic {
			t.Fatalf("set = %+v", got)
		}
	})

	t.Run("gap-only dynamic identity", func(t *testing.T) {
		trace := newRequirementTrace()
		reference := testRequirementReference("ref-0", "lookup", "", "read", "not_applicable", "dynamic", 0)
		trace.recordReference(reference, false, true, trace.nextEvent())
		got := mustProjectRequirements(t, trace)
		if len(got.Items) != 0 || len(got.Gaps) != 1 || got.Gaps[0].Code != CodeRequirementDynamic || !reflect.DeepEqual(got.Gaps[0].ReferenceIDs, []string{"ref-0"}) {
			t.Fatalf("set = %+v", got)
		}
	})

	t.Run("exact macro with expansion gap", func(t *testing.T) {
		trace := newRequirementTrace()
		reference := testRequirementReference("ref-0", "macro", "expand", "read", "not_applicable", "exact", 0)
		trace.recordReference(reference, true, false, trace.nextEvent())
		trace.recordDiagnostic(Diagnostic{Code: CodeDynamicReference, Severity: "warning", Category: "unsupported_semantics", Message: "macro expansion is unresolved", Location: reference.Location, StageID: reference.StageID, ScopeID: reference.ScopeID}, true, []string{"ref-0"}, trace.nextEvent())
		got := mustProjectRequirements(t, trace)
		if len(got.Items) != 1 || got.Items[0].Necessity != "required" || len(got.Gaps) != 1 || got.Gaps[0].Code != CodeDynamicReference {
			t.Fatalf("set = %+v", got)
		}
	})
}

func testRequirementReference(id, kind, identity, role, binding, resolution string, offset int) Reference {
	return Reference{
		ID:                 id,
		OriginalName:       identity,
		NormalizedName:     identity,
		Kind:               kind,
		Role:               role,
		StageID:            "stage-0",
		ScopeID:            "scope-0",
		Location:           testRequirementLocation(offset, offset+len(identity)),
		Resolution:         resolution,
		Binding:            binding,
		OriginReferenceIDs: []string{},
	}
}

func mustProjectRequirements(t *testing.T, trace *requirementTrace) RequirementSet {
	t.Helper()
	got, err := projectRequirements(QueryDocument{Text: "query", Language: "spl", Profile: "splunkd", Version: "current"}, trace)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestProjectRequirementsGapOrderingAndDeduplication(t *testing.T) {
	trace := newRequirementTrace()
	indeterminate := testRequirementReference("ref-0", "field", "host", "read", "indeterminate", "exact", 0)
	dynamic := testRequirementReference("ref-1", "lookup", "$lookup$", "read", "not_applicable", "dynamic", 5)
	trace.recordReference(indeterminate, false, true, trace.nextEvent())
	trace.recordReference(dynamic, false, true, trace.nextEvent())
	first := Diagnostic{Code: CodeUnsupportedSemantics, Severity: "warning", Category: "unsupported_semantics", Message: "first deterministic message", Location: indeterminate.Location, StageID: "stage-0", ScopeID: "scope-0"}
	duplicate := first
	duplicate.Message = "later duplicate message"
	trace.recordDiagnostic(first, true, []string{"ref-0"}, trace.nextEvent())
	trace.recordDiagnostic(duplicate, true, []string{"ref-0"}, trace.nextEvent())
	trace.recordDiagnostic(first, true, []string{"ref-1"}, trace.nextEvent())

	got := mustProjectRequirements(t, trace)
	wantCodes := []string{CodeRequirementIndeterminate, CodeRequirementDynamic, CodeUnsupportedSemantics, CodeUnsupportedSemantics}
	if len(got.Gaps) != len(wantCodes) {
		t.Fatalf("gaps = %+v, want %d", got.Gaps, len(wantCodes))
	}
	for i, code := range wantCodes {
		if got.Gaps[i].Code != code {
			t.Fatalf("gap order = %+v", got.Gaps)
		}
	}
	if got.Gaps[2].Message != first.Message || !reflect.DeepEqual(got.Gaps[2].ReferenceIDs, []string{"ref-0"}) || !reflect.DeepEqual(got.Gaps[3].ReferenceIDs, []string{"ref-1"}) {
		t.Fatalf("diagnostic gap deduplication changed first evidence: %+v", got.Gaps)
	}
	wantReasons := []string{CodeRequirementIndeterminate, CodeRequirementDynamic, CodeUnsupportedSemantics}
	if !reflect.DeepEqual(got.Coverage.Reasons, wantReasons) {
		t.Fatalf("coverage reasons = %v, want %v", got.Coverage.Reasons, wantReasons)
	}

	t.Run("unexplained incomplete coverage", func(t *testing.T) {
		trace := newRequirementTrace()
		trace.semanticComplete = false
		got := mustProjectRequirements(t, trace)
		if len(got.Gaps) != 1 || got.Gaps[0].Code != CodeRequirementCoverageIncomplete || got.Gaps[0].ReferenceIDs == nil || got.Gaps[0].DiagnosticCodes == nil {
			t.Fatalf("fallback gap = %+v", got.Gaps)
		}
		if got.Coverage.Complete || !reflect.DeepEqual(got.Coverage.Reasons, []string{CodeRequirementCoverageIncomplete}) {
			t.Fatalf("fallback coverage = %+v", got.Coverage)
		}
	})
}

func TestRequirementProjectionIndexUsesExactEvidenceKeys(t *testing.T) {
	trace := newRequirementTrace()
	trace.recordDiagnostic(Diagnostic{Code: CodeUnresolvedWildcard}, true, []string{"ref-0", "ref-2"}, trace.nextEvent())
	trace.recordDiagnostic(Diagnostic{Code: CodeUnresolvedWildcard}, false, []string{"ref-1"}, trace.nextEvent())
	trace.recordDiagnostic(Diagnostic{Code: CodeDynamicReference}, true, []string{"ref-0"}, trace.nextEvent())

	index := newRequirementProjectionIndex(trace)
	for _, tc := range []struct {
		referenceID string
		code        string
		want        bool
	}{
		{"ref-0", CodeUnresolvedWildcard, true},
		{"ref-2", CodeUnresolvedWildcard, true},
		{"ref-1", CodeUnresolvedWildcard, false},
		{"ref-0", CodeDynamicReference, true},
		{"ref-2", CodeDynamicReference, false},
	} {
		if got := index.hasOwnedDiagnostic(tc.referenceID, tc.code); got != tc.want {
			t.Errorf("owned diagnostic (%q, %q) = %t, want %t", tc.referenceID, tc.code, got, tc.want)
		}
	}

	set := RequirementSet{
		Coverage: RequirementCoverage{Complete: true, Reasons: []string{}},
		Gaps:     []RequirementGap{},
	}
	first := RequirementGap{Code: "gap-a", Message: "first message", ReferenceIDs: []string{"ref-0", "ref-1"}, DiagnosticCodes: []string{"diag-0", "diag-1"}}
	index.appendGap(&set, first)
	duplicate := first
	duplicate.Message = "later message"
	index.appendGap(&set, duplicate)
	reversedReferences := first
	reversedReferences.ReferenceIDs = []string{"ref-1", "ref-0"}
	index.appendGap(&set, reversedReferences)
	reversedDiagnostics := first
	reversedDiagnostics.DiagnosticCodes = []string{"diag-1", "diag-0"}
	index.appendGap(&set, reversedDiagnostics)
	index.appendGap(&set, RequirementGap{Code: "gap-b", Message: "second reason", ReferenceIDs: []string{}, DiagnosticCodes: []string{}})
	index.appendGap(&set, RequirementGap{Code: "gap-b", Message: "empty reference identity", ReferenceIDs: []string{""}, DiagnosticCodes: []string{}})

	if len(set.Gaps) != 5 {
		t.Fatalf("gaps = %+v, want five exact evidence keys", set.Gaps)
	}
	if set.Gaps[0].Message != first.Message {
		t.Fatalf("first gap message = %q, want %q", set.Gaps[0].Message, first.Message)
	}
	if got, want := set.Coverage.Reasons, []string{"gap-a", "gap-b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("coverage reasons = %v, want %v", got, want)
	}
}

func TestProjectRequirementsStatusIndependentFromCoverage(t *testing.T) {
	const query = "search host=* | fields - host | table host"
	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := projectRequirements(QueryDocument{Text: query, Language: "spl", Profile: "splunkd", Version: "current"}, trace)
	if err != nil {
		t.Fatal(err)
	}
	if got.QueryStatus != Invalid {
		t.Fatalf("query status = %q, want invalid", got.QueryStatus)
	}
	if !got.Coverage.Complete || len(got.Gaps) != 0 {
		t.Fatalf("local invalidity changed requirement coverage: %+v", got)
	}
	for _, item := range got.Items {
		for _, occurrence := range item.Occurrences {
			if occurrence.Binding == "unavailable" {
				t.Fatalf("unavailable local read became an external obligation: %+v", item)
			}
		}
	}
	foundDiagnostic := false
	for _, diagnostic := range got.Diagnostics {
		if diagnostic.Code == CodeUnavailableField {
			foundDiagnostic = true
		}
	}
	if !foundDiagnostic {
		t.Fatalf("query-only diagnostics omitted local invalidity: %+v", got.Diagnostics)
	}
}

func TestAnalyzeAlwaysEmbedsRequirementSet(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func() (*Result, error)
	}{
		{"plain SPL", func() (*Result, error) { return Analyze(QueryDocument{}) }},
		{"plain SPL2", func() (*Result, error) { return Analyze(QueryDocument{Language: "spl2"}) }},
		{"field refinement", func() (*Result, error) {
			result, err := AnalyzeWithSourceFields(QueryDocument{}, []string{})
			if err != nil {
				return nil, err
			}
			return result.Result, nil
		}},
		{"source-universe refinement", func() (*Result, error) {
			result, err := AnalyzeWithSourceUniverse(QueryDocument{}, SourceUniverse{Fields: []string{}})
			if err != nil {
				return nil, err
			}
			return result.Result, nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.run()
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(result.Requirements)
			if err != nil {
				t.Fatal(err)
			}
			if result.Requirements.SchemaVersion != 1 || result.Requirements.Items == nil || result.Requirements.Gaps == nil || result.Requirements.Diagnostics == nil || result.Requirements.Coverage.Reasons == nil {
				t.Fatalf("embedded requirement set is not initialized: %s", encoded)
			}
		})
	}
}

func TestAnalyzeSPLNullInspectionIsNotARequirement(t *testing.T) {
	for _, query := range []string{
		"| where isnull(missing)",
		"| eval present=isnotnull(missing)",
	} {
		t.Run(query, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: query})
			if err != nil {
				t.Fatal(err)
			}
			var found *Reference
			for i := range result.References {
				if result.References[i].NormalizedName == "missing" {
					found = &result.References[i]
					break
				}
			}
			if found == nil || found.Role != "null_test" || found.Binding != "source" {
				t.Fatalf("canonical null inspection = %+v", found)
			}
			if len(result.Requirements.Items) != 0 {
				t.Fatalf("null inspection became an external requirement: %+v", result.Requirements.Items)
			}
		})
	}
}

func TestAnalyzeSPLNullInspectionKeepsIneligibleReads(t *testing.T) {
	for _, tc := range []struct {
		query    string
		wantRead bool
	}{
		{"| where isnull(missing, 1)", true},
		{"| where isnull(missing + other)", true},
		{"| where isnull((missing))", true},
		{"| stats isnull(missing)", true},
		{"| where abs(value:isnull(missing))=1", true},
		{"| where unknown(isnull(missing))=1", true},
		{"| where searchmatch(isnull(missing))=1", true},
		{"| where isnull(missing", false},
		{"| where isnull(missing @)", false},
		{"| where isnull(missing[0])", false},
	} {
		t.Run(tc.query, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: tc.query})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, reference := range result.References {
				if reference.NormalizedName != "missing" {
					continue
				}
				found = true
				if reference.Role != "read" {
					t.Fatalf("ineligible null inspection role = %q, want read: %+v", reference.Role, reference)
				}
			}
			if tc.wantRead && !found {
				t.Fatal("ineligible null inspection lost its conservative field read")
			}
		})
	}
}

func TestRequirementsMatchesEmbeddedAndIsDetached(t *testing.T) {
	document := QueryDocument{Text: "search host=* | `expand()` | table h*", SourceID: "queries/example.spl"}
	result, err := Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	standalone, err := Requirements(document)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*standalone, result.Requirements) {
		t.Fatalf("standalone requirements differ from embedded set:\nstandalone: %+v\n  embedded: %+v", *standalone, result.Requirements)
	}
	if len(standalone.Coverage.Reasons) == 0 || len(standalone.Items) == 0 || len(standalone.Items[0].Occurrences) == 0 || len(standalone.Gaps) == 0 || len(standalone.Diagnostics) == 0 {
		t.Fatalf("detachment witness lacks nested values: %+v", *standalone)
	}
	standalone.Coverage.Reasons[0] = "changed"
	standalone.Items[0].Occurrences[0].ReferenceID = "changed"
	standalone.Gaps[0].ReferenceIDs[0] = "changed"
	standalone.Gaps[0].DiagnosticCodes[0] = "changed"
	standalone.Diagnostics[0].Code = "changed"
	if reflect.DeepEqual(*standalone, result.Requirements) {
		t.Fatal("standalone requirement set aliases the embedded analysis result")
	}
	repeated, err := Requirements(document)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*repeated, result.Requirements) {
		t.Fatalf("standalone mutation leaked into a repeated call:\n repeated: %+v\n embedded: %+v", *repeated, result.Requirements)
	}

	for _, c := range loadRequirementsCorpus(t) {
		t.Run(c.ID, func(t *testing.T) {
			result, err := Analyze(c.Document)
			if err != nil {
				t.Fatal(err)
			}
			standalone, err := Requirements(c.Document)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*standalone, result.Requirements) {
				t.Fatalf("standalone requirements differ from embedded set: %+v", *standalone)
			}
			embeddedBefore, err := json.Marshal(result.Requirements)
			if err != nil {
				t.Fatal(err)
			}
			standalone.Coverage.Reasons = append(standalone.Coverage.Reasons, "changed")
			standalone.Items = append(standalone.Items, RequirementItem{ID: "changed", Occurrences: []RequirementOccurrence{}})
			standalone.Gaps = append(standalone.Gaps, RequirementGap{Code: "changed", ReferenceIDs: []string{}, DiagnosticCodes: []string{}})
			standalone.Diagnostics = append(standalone.Diagnostics, Diagnostic{Code: "changed"})
			if len(c.Expected.Items) > 0 && len(c.Expected.Items[0].Occurrences) > 0 {
				standalone.Items[0].Occurrences[0].ReferenceID = "changed"
			}
			if len(c.Expected.Gaps) > 0 {
				standalone.Gaps[0].Code = "changed"
				if len(c.Expected.Gaps[0].ReferenceIDs) > 0 {
					standalone.Gaps[0].ReferenceIDs[0] = "changed"
				}
				if len(c.Expected.Gaps[0].DiagnosticCodes) > 0 {
					standalone.Gaps[0].DiagnosticCodes[0] = "changed"
				}
			}
			if len(c.Expected.Diagnostics) > 0 {
				standalone.Diagnostics[0].Code = "changed"
			}
			embeddedAfter, err := json.Marshal(result.Requirements)
			if err != nil {
				t.Fatal(err)
			}
			if string(embeddedAfter) != string(embeddedBefore) {
				t.Fatalf("standalone mutation changed embedded set:\n before: %s\n  after: %s", embeddedBefore, embeddedAfter)
			}
		})
	}

	invalid := QueryDocument{Text: "search host=*", Language: "SPL"}
	_, analyzeErr := Analyze(invalid)
	invalidSet, requirementsErr := Requirements(invalid)
	if analyzeErr == nil || requirementsErr == nil || requirementsErr.Error() != analyzeErr.Error() || invalidSet != nil {
		t.Fatalf("Requirements error = (%+v, %v), Analyze error = %v", invalidSet, requirementsErr, analyzeErr)
	}
}

func TestRequirementTypesJSONShape(t *testing.T) {
	set := RequirementSet{
		SchemaVersion: 1,
		Query: RequirementQueryIdentity{
			SourceID:    "queries/example.spl",
			Language:    "spl",
			Profile:     "splunkd",
			Version:     "current",
			QueryDigest: "sha256:query",
		},
		CapabilityRevision: "sha256:capability",
		QueryStatus:        Incomplete,
		Coverage: RequirementCoverage{
			Complete: false,
			Reasons:  []string{"SPL_REQUIREMENT_DYNAMIC"},
		},
		Items: []RequirementItem{{
			ID:         "req-1",
			Kind:       "field",
			Identity:   "host",
			Role:       "read",
			Necessity:  "required",
			Origin:     "direct",
			Resolution: "exact",
			Occurrences: []RequirementOccurrence{{
				ReferenceID:  "ref-1",
				OriginalName: "Host",
				Binding:      "source",
				StageID:      "stage-0",
				ScopeID:      "scope-0",
				Location: Location{
					Start: Position{Offset: 7, Line: 1, Column: 8},
					End:   Position{Offset: 11, Line: 1, Column: 12},
				},
			}},
		}},
		Gaps: []RequirementGap{{
			Code:            "SPL_REQUIREMENT_DYNAMIC",
			Message:         "dynamic requirement",
			ReferenceIDs:    []string{"ref-1"},
			DiagnosticCodes: []string{"SPL_DYNAMIC_REFERENCE"},
		}},
		Diagnostics: []Diagnostic{{
			Code:     "SPL_DYNAMIC_REFERENCE",
			Severity: "warning",
			Category: "semantic",
			Message:  "dynamic reference",
			Location: Location{
				Start: Position{Offset: 7, Line: 1, Column: 8},
				End:   Position{Offset: 11, Line: 1, Column: 12},
			},
			StageID: "stage-0",
			ScopeID: "scope-0",
		}},
	}

	encoded, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	var got any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	const expected = `{"schema_version":1,"query":{"source_id":"queries/example.spl","language":"spl","profile":"splunkd","version":"current","query_digest":"sha256:query"},"capability_revision":"sha256:capability","query_status":"incomplete","coverage":{"complete":false,"reasons":["SPL_REQUIREMENT_DYNAMIC"]},"items":[{"id":"req-1","kind":"field","identity":"host","role":"read","necessity":"required","origin":"direct","resolution":"exact","occurrences":[{"reference_id":"ref-1","original_name":"Host","binding":"source","stage_id":"stage-0","scope_id":"scope-0","location":{"start":{"offset":7,"line":1,"column":8},"end":{"offset":11,"line":1,"column":12}}}]}],"gaps":[{"code":"SPL_REQUIREMENT_DYNAMIC","message":"dynamic requirement","reference_ids":["ref-1"],"diagnostic_codes":["SPL_DYNAMIC_REFERENCE"]}],"diagnostics":[{"code":"SPL_DYNAMIC_REFERENCE","severity":"warning","category":"semantic","message":"dynamic reference","location":{"start":{"offset":7,"line":1,"column":8},"end":{"offset":11,"line":1,"column":12}},"stage_id":"stage-0","scope_id":"scope-0"}]}`
	var want any
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("requirement JSON shape mismatch:\n got: %s\nwant: %s", encoded, expected)
	}
}

func TestQueryDigestExactBytes(t *testing.T) {
	for _, text := range []string{"", "café 😀", "a\nb", "a\r\nb"} {
		sum := sha256.Sum256([]byte(text))
		want := "sha256:" + hex.EncodeToString(sum[:])
		if got := queryDigest(text); got != want {
			t.Errorf("queryDigest(%q) = %q, want %q", text, got, want)
		}
	}

	left := QueryDocument{Text: "search café=*", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "left.spl"}
	right := QueryDocument{Text: left.Text, Language: "spl2", Profile: "unused", Version: "future", SourceID: "right.spl"}
	if queryDigest(left.Text) != queryDigest(right.Text) {
		t.Fatal("selector-only changes affected the query digest")
	}
}

func TestCapabilityRevisionUsesNormalizedTypedManifest(t *testing.T) {
	defaultDocument := QueryDocument{}
	explicitDocument := QueryDocument{Language: "spl", Profile: "splunkd", Version: "current"}

	defaultRevision, err := capabilityRevision(defaultDocument)
	if err != nil {
		t.Fatal(err)
	}
	explicitRevision, err := capabilityRevision(explicitDocument)
	if err != nil {
		t.Fatal(err)
	}
	if defaultRevision != explicitRevision {
		t.Fatalf("default revision %q differs from normalized explicit revision %q", defaultRevision, explicitRevision)
	}
	repeatedRevision, err := capabilityRevision(defaultDocument)
	if err != nil {
		t.Fatal(err)
	}
	if repeatedRevision != defaultRevision {
		t.Fatalf("repeated revision %q differs from first revision %q", repeatedRevision, defaultRevision)
	}

	spl2Document := QueryDocument{Language: "spl2", Profile: "splunkd", Version: "current"}
	spl2Revision, err := capabilityRevision(spl2Document)
	if err != nil {
		t.Fatal(err)
	}
	if spl2Revision == defaultRevision {
		t.Fatalf("SPL and SPL2 capability revisions are both %q", defaultRevision)
	}

	for _, document := range []QueryDocument{defaultDocument, explicitDocument, spl2Document} {
		manifest, err := CapabilitiesFor(CapabilityOptions{
			Language: document.Language,
			Profile:  document.Profile,
			Version:  document.Version,
		})
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(encoded)
		want := "sha256:" + hex.EncodeToString(sum[:])
		got, err := capabilityRevision(document)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("capabilityRevision(%+v) = %q, want %q", document, got, want)
		}
	}

	if got, err := capabilityRevision(QueryDocument{Language: "SPL"}); err == nil || got != "" {
		t.Fatalf("unsupported selectors returned revision %q and error %v", got, err)
	}
}

func TestCloneRequirementSetOwnsNestedSlices(t *testing.T) {
	source := RequirementSet{
		Coverage: RequirementCoverage{Reasons: []string{"reason"}},
		Items: []RequirementItem{{
			ID:          "req-1",
			Occurrences: []RequirementOccurrence{{ReferenceID: "ref-1"}},
		}},
		Gaps: []RequirementGap{{
			Code:            "gap",
			ReferenceIDs:    []string{"ref-1"},
			DiagnosticCodes: []string{"diagnostic"},
		}},
		Diagnostics: []Diagnostic{{Code: "diagnostic"}},
	}
	before, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}

	cloned := cloneRequirementSet(source)
	cloned.Coverage.Reasons[0] = "changed reason"
	cloned.Items[0].ID = "req-changed"
	cloned.Items[0].Occurrences[0].ReferenceID = "ref-changed"
	cloned.Gaps[0].Code = "changed gap"
	cloned.Gaps[0].ReferenceIDs[0] = "ref-changed"
	cloned.Gaps[0].DiagnosticCodes[0] = "changed diagnostic link"
	cloned.Diagnostics[0].Code = "changed diagnostic"

	after, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("clone mutations changed source:\n before: %s\n  after: %s", before, after)
	}
}

func TestRequirementSetEmptyCollectionsAreArrays(t *testing.T) {
	set := RequirementSet{
		Coverage:    RequirementCoverage{Reasons: []string{}},
		Items:       []RequirementItem{},
		Gaps:        []RequirementGap{},
		Diagnostics: []Diagnostic{},
	}
	encoded, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	coverage, ok := decoded["coverage"].(map[string]any)
	if !ok {
		t.Fatalf("coverage is %T, want object", decoded["coverage"])
	}
	for name, value := range map[string]any{
		"coverage.reasons": coverage["reasons"],
		"items":            decoded["items"],
		"gaps":             decoded["gaps"],
		"diagnostics":      decoded["diagnostics"],
	} {
		items, ok := value.([]any)
		if !ok || len(items) != 0 {
			t.Errorf("%s = %#v (%T), want []", name, value, value)
		}
	}
}

func TestRequirementDiagnosticCodes(t *testing.T) {
	want := []string{
		"SPL_REQUIREMENT_INDETERMINATE",
		"SPL_REQUIREMENT_DYNAMIC",
		"SPL_REQUIREMENT_COVERAGE_INCOMPLETE",
	}
	got := []string{
		CodeRequirementIndeterminate,
		CodeRequirementDynamic,
		CodeRequirementCoverageIncomplete,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("requirement diagnostic codes = %q, want %q", got, want)
	}
}

func BenchmarkProjectRequirementsWildcardGapScaling(b *testing.B) {
	for _, size := range []int{128, 1024, 8192} {
		b.Run(fmt.Sprintf("references_%d", size), func(b *testing.B) {
			trace := newRequirementTrace()
			for i := 0; i < size; i++ {
				reference := testRequirementReference(
					fmt.Sprintf("ref-%d", i),
					"field",
					fmt.Sprintf("field_%d*", i),
					"read",
					"indeterminate",
					"wildcard",
					i,
				)
				trace.recordReference(reference, false, true, trace.nextEvent())
				trace.recordDiagnostic(Diagnostic{
					Code:     CodeUnresolvedWildcard,
					Severity: "warning",
					Message:  "wildcard field membership is unresolved",
					Location: reference.Location,
					StageID:  reference.StageID,
					ScopeID:  reference.ScopeID,
				}, true, []string{reference.ID}, trace.nextEvent())
			}

			document := QueryDocument{Text: "benchmark", Language: "spl", Profile: "splunkd", Version: "current"}
			b.ReportAllocs()
			b.ReportMetric(float64(size), "references/op")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := projectRequirements(document, trace); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
