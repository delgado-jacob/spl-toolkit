package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestInputDiscoveryIdentityAndOccurrences(t *testing.T) {
	for _, query := range []string{"from $events | join left=e right=r where e.id=r.id [from $events]", "FROM 'évents' AS e | fields id"} {
		doc := QueryDocument{Text: query, Language: "spl2"}
		a, err := Analyze(doc)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Analyze(doc)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a.Inputs, b.Inputs) || !reflect.DeepEqual(a.Inputs, a.Requirements.Inputs) {
			t.Fatal("input evidence is unstable or disagrees with requirements")
		}
		if len(a.Inputs) != 1 {
			t.Fatalf("inputs: %#v", a.Inputs)
		}
		for _, o := range a.Inputs[0].Occurrences {
			if query[o.Location.Start.Offset:o.Location.End.Offset] != map[bool]string{true: "'évents'", false: a.Inputs[0].Name}[strings.Contains(query, "évents")] {
				t.Fatal("source slice lost")
			}
		}
	}
	a, _ := Analyze(QueryDocument{Text: "FROM main AS a", Language: "spl2"})
	b, _ := Analyze(QueryDocument{Text: "FROM main AS b", Language: "spl2"})
	if a.Inputs[0].ID != b.Inputs[0].ID || a.Inputs[0].Occurrences[0].Alias != "a" {
		t.Fatal("alias changed logical identity or was lost")
	}
	if opaqueInputID("explicit_dataset", InputIdentity{Form: "identifier", Value: "main"}) == opaqueInputID("explicit_dataset", InputIdentity{Form: "dotted", Value: "main"}) {
		t.Fatal("forms collapsed")
	}
}

func TestInputViewSummaryReuse(t *testing.T) {
	query := "$base = FROM $events | fields id;\n$consumer = FROM $base | fields id;\n$other = FROM $base | fields id;"
	result, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Inputs) != 1 || len(result.Inputs[0].Occurrences) != 2 {
		t.Fatalf("view inputs: %#v", result.Inputs)
	}
	for _, occurrence := range result.Inputs[0].Occurrences {
		if len(occurrence.UseSiteLocations) != 1 {
			t.Fatal("missing view use context")
		}
	}
}

func TestInputAbortedDiscoveryIsPartial(t *testing.T) {
	for _, doc := range []QueryDocument{{Text: "from $events | eval (", Language: "spl2"}, {Text: strings.Repeat("x ", lexerWorkLimit+1)}} {
		r, err := Analyze(doc)
		if err != nil {
			t.Fatal(err)
		}
		if r.InputCoverage.State != "partial" || r.Correlation.Coverage.State == "not_applicable" {
			t.Fatalf("aborted discovery claimed exhaustive absence: %#v", r.InputCoverage)
		}
	}
}

func TestInputCapabilityRevisionNormalized(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: "from $events", Language: "spl2"})
	revision, err := CapabilityRevisionFor(CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if revision != r.Requirements.CapabilityRevision {
		t.Fatal("revision normalization differs")
	}
}

func TestInputTypedFormsCanonicalDescriptorsAndLocals(t *testing.T) {
	cases := []struct {
		query, kind, form, name string
		count                   int
	}{
		{"FROM $events", "named_placeholder", "parameter", "$events", 1},
		{"FROM '$events'", "explicit_dataset", "identifier", "$events", 1},
		{"FROM 'main.events'", "explicit_dataset", "identifier", "main.events", 1},
		{"FROM main.events", "explicit_dataset", "dotted", "main.events", 1},
		{`FROM {kind: "index", properties: {name: "main"}}`, "explicit_dataset", "descriptor", `{"kind":"index","properties":{"name":"main"}}`, 1},
		{`FROM {kind: $kind, properties: {name: "main"}}`, "unresolved_source", "descriptor", `{kind: $kind, properties: {name: "main"}}`, 1},
		{`FROM [{id: 1}] | fields id`, "", "", "", 0},
		{`function identity($local) { return $local; }
$base = FROM $events | eval copied=identity(id), mapped=map(items, ($item) -> $item);`, "named_placeholder", "parameter", "$events", 1},
	}
	ids := map[string]string{}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			r, err := Analyze(QueryDocument{Text: tc.query, Language: "spl2"})
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Inputs) != tc.count {
				t.Fatalf("got inputs %#v", r.Inputs)
			}
			if tc.count == 0 {
				if r.InputCoverage.State != "not_applicable" {
					t.Fatalf("literal generated rows: %#v", r.InputCoverage)
				}
				return
			}
			in := r.Inputs[0]
			if in.Kind != tc.kind || in.Identity.Form != tc.form || in.Name != tc.name {
				t.Fatalf("unexpected identity %#v", in)
			}
			if previous := ids[tc.name]; previous == in.ID {
				t.Fatal("equal display names collapsed typed source forms")
			}
			ids[tc.name] = in.ID
		})
	}
	a, err := Analyze(QueryDocument{Text: `FROM {kind: "index", properties: {name: "main", nested: {a: 1, b: 2}}}`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Analyze(QueryDocument{Text: `FROM {kind: "index", properties: {nested: {b: 2, a: 1}, name: "main"}}`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Inputs) != 1 || len(b.Inputs) != 1 || a.Inputs[0].ID != b.Inputs[0].ID {
		t.Fatal("descriptor canonicalization differs")
	}
}

func TestInputDiscoveryCoverageIsDimensionSpecific(t *testing.T) {
	for _, query := range []string{"FROM main | where unknown_fn(id)", "FROM $events AS e JOIN $users AS u ON e.id=u.id SELECT e.id"} {
		r, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
		if err != nil {
			t.Fatal(err)
		}
		if r.InputCoverage.State != "complete" {
			t.Fatalf("field scheduling gap demoted discovery: %#v", r.InputCoverage)
		}
	}
	for _, query := range []string{"FROM []", "FROM [unknown_fn(id)]", "| mystery"} {
		r, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
		if err != nil {
			t.Fatal(err)
		}
		if r.InputCoverage.State != "partial" || len(r.InputCoverage.Reasons) == 0 {
			t.Fatalf("unproved discovery claimed absence: %#v", r.InputCoverage)
		}
	}
}

func TestInputSourceIdentityAndTraceDetached(t *testing.T) {
	a, _ := Analyze(QueryDocument{Text: "FROM $events", Language: "spl2", SourceID: "a"})
	b, _ := Analyze(QueryDocument{Text: "FROM $events", Language: "spl2", SourceID: "b"})
	if a.Inputs[0].ID != b.Inputs[0].ID || a.Inputs[0].Occurrences[0].ID == b.Inputs[0].Occurrences[0].ID {
		t.Fatal("logical and situated source identities conflated")
	}
	original := a.Requirements.Inputs[0].Occurrences[0].UseSiteLocations
	a.Inputs[0].Occurrences[0].UseSiteLocations = append(a.Inputs[0].Occurrences[0].UseSiteLocations, Location{})
	if !reflect.DeepEqual(original, a.Requirements.Inputs[0].Occurrences[0].UseSiteLocations) {
		t.Fatal("result mutation leaked")
	}
	trace := newRequirementTrace()
	trace.inputs = []inputFact{{sourceID: "a", kind: "named_placeholder", identity: InputIdentity{Form: "parameter", Value: "$events"}, occurrence: InputOccurrence{UseSiteLocations: []Location{{}}, UseSiteReferenceIDs: []string{"pending-0"}}}}
	copy := trace.clone()
	copy.inputs[0].occurrence.UseSiteReferenceIDs[0] = "changed"
	copy.inputs[0].occurrence.UseSiteLocations[0].Start.Offset = 42
	if trace.inputs[0].occurrence.UseSiteReferenceIDs[0] != "pending-0" || trace.inputs[0].occurrence.UseSiteLocations[0].Start.Offset != 0 {
		t.Fatal("trace clone aliases contextual input slices")
	}
}

func TestInputRequirementsCorpusDiscoveryContract(t *testing.T) {
	// The existing requirement corpus protects unchanged obligation semantics.
	// These independently authored source expectations protect its added evidence.
	partial := map[string]bool{"indeterminate_spl": true, "macro_spl": true, "syntax_incomplete_spl": true, "spl2_recovered_diagnostics": true, "branch_macro_boundaries_spl": true}
	occurrences := map[string]int{"branch_scope_spl": 2, "knowledge_composed_spl": 5, "branch_macro_boundaries_spl": 2}
	for _, c := range loadRequirementsCorpus(t) {
		t.Run(c.ID, func(t *testing.T) {
			r, err := Requirements(c.Document)
			if err != nil {
				t.Fatal(err)
			}
			kind, form, name, state, count, n := "implicit_stream", "implicit", "", "complete", 1, 1
			if c.Document.Language == "spl2" {
				kind, form, name = "explicit_dataset", "identifier", "main"
			}
			if c.ID == "spl2_literal_invalid" {
				count, n, state = 0, 0, "not_applicable"
			}
			if partial[c.ID] {
				state = "partial"
			}
			if expected, ok := occurrences[c.ID]; ok {
				n = expected
			}
			if len(r.Inputs) != count || r.InputCoverage.State != state {
				t.Fatalf("discovery count/state = %d/%s, want %d/%s", len(r.Inputs), r.InputCoverage.State, count, state)
			}
			if count == 0 {
				return
			}
			in := r.Inputs[0]
			if in.Kind != kind || in.Identity.Form != form || in.Name != name || len(in.Occurrences) != n {
				t.Fatalf("source identity/occurrences = %#v, want %s/%s/%s/%d", in, kind, form, name, n)
			}
		})
	}
}

func TestInputCompatibilityCorpusSources(t *testing.T) {
	var cases []struct {
		Name     string        `json:"name"`
		Document QueryDocument `json:"document"`
		Expected struct {
			InputCount      *int `json:"input_count"`
			OccurrenceCount int  `json:"occurrence_count"`
			Inputs          []struct {
				Kind, Name, Form string
				SourceSlices     []string `json:"source_slices"`
			} `json:"inputs"`
		} `json:"expected"`
	}
	data, err := os.ReadFile("../../testdata/compatibility/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if c.Expected.InputCount == nil {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			result, err := Analyze(c.Document)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Inputs) != *c.Expected.InputCount {
				t.Fatalf("input count=%d, want %d: %#v", len(result.Inputs), *c.Expected.InputCount, result.Inputs)
			}
			total := 0
			for i, in := range result.Inputs {
				expected := c.Expected.Inputs[i]
				total += len(in.Occurrences)
				if in.Kind != expected.Kind || in.Name != expected.Name || in.Identity.Form != expected.Form {
					t.Fatalf("input %d identity mismatch: %#v", i, in)
				}
				slices := []string{}
				for _, o := range in.Occurrences {
					slices = append(slices, c.Document.Text[o.Location.Start.Offset:o.Location.End.Offset])
				}
				if !reflect.DeepEqual(slices, expected.SourceSlices) {
					t.Fatalf("source slices %v, want %v", slices, expected.SourceSlices)
				}
			}
			if total != c.Expected.OccurrenceCount {
				t.Fatalf("occurrences=%d, want %d", total, c.Expected.OccurrenceCount)
			}
		})
	}
}

// Hand-built full-state oracles retain their independently authored field
// transfers and add the exact external source and pending ownership evidence.
func addExactInputTestEvidence(want *Result, sourceReference Reference, fieldReferences ...Reference) {
	identity := InputIdentity{Form: "identifier", Value: sourceReference.NormalizedName}
	fact := inputFact{sourceID: want.Document.SourceID, kind: "explicit_dataset", name: sourceReference.NormalizedName, identity: identity, occurrence: InputOccurrence{ReferenceID: sourceReference.ID, OriginalReferenceID: sourceReference.ID, StageID: sourceReference.StageID, ScopeID: sourceReference.ScopeID, Location: sourceReference.Location, UseSiteLocations: []Location{}, UseSiteReferenceIDs: []string{}}}
	fact.occurrence.ID = inputOccurrenceID(fact)
	want.Inputs = []QueryInput{{ID: opaqueInputID(fact.kind, identity), Kind: fact.kind, Name: fact.name, Identity: identity, Evidence: InputCoverage{State: "complete", Reasons: []InputReason{}}, Occurrences: []InputOccurrence{fact.occurrence}}}
	want.InputCoverage = InputCoverage{State: "complete", Reasons: []InputReason{}}
	want.FieldAttributionCoverage = InputCoverage{State: "partial", Reasons: []InputReason{}}
	for _, ref := range fieldReferences {
		want.FieldAttributionCoverage.Reasons = append(want.FieldAttributionCoverage.Reasons, InputReason{Code: "field_attribution_incomplete", Message: "source field ownership has not been proved", Location: ref.Location, StageID: ref.StageID, ScopeID: ref.ScopeID, ReferenceIDs: []string{ref.ID}})
	}
	want.Correlation = CorrelationGraph{Outcome: "not_applicable", Coverage: InputCoverage{State: "not_applicable", Reasons: []InputReason{}}, Nodes: []CorrelationNode{{InputID: want.Inputs[0].ID, OccurrenceID: fact.occurrence.ID}}, Edges: []CorrelationEdge{}, Components: [][]string{{fact.occurrence.ID}}}
	want.Requirements.Inputs = cloneInputs(want.Inputs)
	want.Requirements.InputCoverage = cloneInputCoverage(want.InputCoverage)
	want.Requirements.FieldAttributionCoverage = cloneInputCoverage(want.FieldAttributionCoverage)
	want.Requirements.Correlation = cloneCorrelation(want.Correlation)
}

func TestInputClassicSourceResetSummary(t *testing.T) {
	document := QueryDocument{Text: "search host=x | inputlookup assets | table id", Language: "spl", Profile: "splunkd", Version: "current"}
	result, trace, err := analyzeRewriteWithTrace(document, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Inputs) != 1 || len(result.Inputs[0].Occurrences) != 2 || len(trace.inputs) != 2 {
		t.Fatalf("source reset occurrence facts missing: %#v", result.Inputs)
	}
	// Each source reset installs its own detached summary for subsequent fields.
	stage := &semanticStage{result: result, stage: 1, env: newEnvironmentWithRequirementTrace(newRequirementTrace())}
	stage.applySource()
	if len(stage.env.inputs) != 1 || stage.env.inputs[0].occurrence.StageID != "stage-1" {
		t.Fatal("source reset lost current summary")
	}
	stage.applySource()
	if len(stage.env.requirements.trace.inputs) != 1 {
		t.Fatal("same stage source-cache reset duplicated occurrence")
	}
}

func TestInputViewAndDynamicSourceAliases(t *testing.T) {
	for _, query := range []string{`$base = FROM $events AS original; $consumer = FROM $base AS situated;`, `FROM {kind: $kind, properties: {name: "main"}} AS situated`} {
		r, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Inputs) != 1 || len(r.Inputs[0].Occurrences) != 1 || r.Inputs[0].Occurrences[0].Alias != "situated" {
			t.Fatalf("situated alias lost: %#v", r.Inputs)
		}
	}
}

func TestInputJoinedSourcesAfterLexicalBaseResolution(t *testing.T) {
	direct, err := Analyze(QueryDocument{Text: `FROM $events AS b JOIN $users AS u ON b.id=u.id | fields b.id`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(direct.Inputs) != 2 {
		t.Fatalf("direct source inputs: %#v", direct.Inputs)
	}
	for _, tc := range []struct {
		name, query, coverage string
		matchDirect           bool
	}{
		{"local view", `$base = FROM $events; $consumer = FROM $base AS b JOIN $users AS u ON b.id=u.id | fields b.id;`, "complete", true},
		{"imported member", `import remote as source from vendor/security; $consumer = FROM source AS b JOIN $users AS u ON b.id=u.id | fields b.id;`, "partial", false},
		{"imported container", `import * as external from vendor/security; $consumer = FROM external.events AS b JOIN $users AS u ON b.id=u.id | fields b.id;`, "partial", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: tc.query, Language: "spl2"})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Inputs) != 2 {
				t.Fatalf("joined source disappeared after lexical binding: %#v", result.Inputs)
			}
			if !reflect.DeepEqual(result.Inputs, result.Requirements.Inputs) || !reflect.DeepEqual(result.InputCoverage, result.Requirements.InputCoverage) || !reflect.DeepEqual(result.FieldAttributionCoverage, result.Requirements.FieldAttributionCoverage) || !reflect.DeepEqual(result.Correlation, result.Requirements.Correlation) {
				t.Fatal("result/requirement query evidence differs")
			}
			if result.InputCoverage.State != tc.coverage {
				t.Fatalf("discovery coverage %#v, want %s", result.InputCoverage, tc.coverage)
			}
			for i, input := range result.Inputs {
				if len(input.Occurrences) != 1 {
					t.Fatalf("unexpected source multiplicity: %#v", input)
				}
				if tc.matchDirect && (input.ID != direct.Inputs[i].ID || input.Identity != direct.Inputs[i].Identity || input.Kind != direct.Inputs[i].Kind) {
					t.Fatalf("local source identity differs from direct source: %#v", input)
				}
			}
			users := result.Inputs[1]
			occurrence := users.Occurrences[0]
			if users.ID != direct.Inputs[1].ID || users.Name != "$users" || occurrence.Alias != "u" || tc.query[occurrence.Location.Start.Offset:occurrence.Location.End.Offset] != "$users" || occurrence.ReferenceID == "" || occurrence.OriginalReferenceID != occurrence.ReferenceID || len(occurrence.UseSiteLocations) != 0 {
				t.Fatalf("joined source occurrence mismatch: %#v", users)
			}
			if tc.matchDirect && len(result.Inputs[0].Occurrences[0].UseSiteLocations) != 1 {
				t.Fatal("local view source context lost")
			}
			unsupported := false
			for _, diagnostic := range result.Diagnostics {
				unsupported = unsupported || diagnostic.Code == CodeUnsupportedSemantics && diagnostic.StageID == occurrence.StageID
			}
			if !unsupported || len(result.Correlation.Edges) != 0 || result.Correlation.Outcome != "indeterminate" {
				t.Fatalf("unmodeled join gained proof: diagnostics=%#v graph=%#v", result.Diagnostics, result.Correlation)
			}
		})
	}
}

func TestInputIndependentUnionHasNoImplicitParent(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `union $events, $users`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Inputs) != 2 || result.Inputs[0].Name != "$events" || result.Inputs[1].Name != "$users" || result.InputCoverage.State != "complete" {
		t.Fatalf("independent union fabricated a parent input: %#v", result.Inputs)
	}
	for _, input := range result.Inputs {
		if input.Kind != "named_placeholder" || len(input.Occurrences) != 1 {
			t.Fatalf("independent union input: %#v", input)
		}
	}
}

func TestInputViewSummarySituatesEveryUnderlyingSource(t *testing.T) {
	for _, base := range []string{
		`FROM $events | join left=e right=u where e.id=u.id [from $users]`,
		`FROM $events | from $users`,
	} {
		t.Run(base, func(t *testing.T) {
			query := `$unrelated = FROM $noise; $base = ` + base + `; $consumer = FROM $base; $other = FROM $base;`
			result, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Inputs) != 3 {
				t.Fatalf("view source discovery: %#v", result.Inputs)
			}
			for _, input := range result.Inputs {
				if input.Name == "$noise" {
					if len(input.Occurrences) != 1 || len(input.Occurrences[0].UseSiteLocations) != 0 {
						t.Fatalf("unrelated declaration leaked into view source summary: %#v", input)
					}
					continue
				}
				if len(input.Occurrences) != 2 {
					t.Fatalf("not every source received terminal view use contexts: %#v", input)
				}
				for _, occurrence := range input.Occurrences {
					if len(occurrence.UseSiteLocations) != 1 || len(occurrence.UseSiteReferenceIDs) != 1 || occurrence.ReferenceID != occurrence.UseSiteReferenceIDs[0] || occurrence.OriginalReferenceID == "" || query[occurrence.Location.Start.Offset:occurrence.Location.End.Offset] != input.Name {
						t.Fatalf("underlying source location/links missing: %#v", occurrence)
					}
					use := occurrence.UseSiteLocations[0]
					if query[use.Start.Offset:use.End.Offset] != "$base" {
						t.Fatalf("wrong terminal use location: %#v", occurrence)
					}
				}
			}
			if !reflect.DeepEqual(result.Inputs, result.Requirements.Inputs) {
				t.Fatal("result and requirements source summaries differ")
			}
		})
	}
}

func TestInputSourceEvidenceResourceLimit(t *testing.T) {
	for _, depth := range []int{12, 15} {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) { testInputSourceEvidenceResourceLimit(t, depth) })
	}
}

func testInputSourceEvidenceResourceLimit(t *testing.T, maxDepth int) {
	t.Helper()
	query := "$v0 = FROM $events;"
	for depth := 1; depth <= maxDepth; depth++ {
		query += fmt.Sprintf("$v%d = union $v%d, $v%d;", depth, depth-1, depth-1)
	}
	assertInputSourceEvidenceResourceLimit(t, query)
}

func TestInputSourceEvidenceLinearViewResourceLimit(t *testing.T) {
	for _, depth := range []int{384, 512} {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			query := "$v0 = FROM $events;\n"
			for link := 1; link <= depth; link++ {
				query += fmt.Sprintf("$v%d=FROM$v%d;", link, link-1)
			}
			assertInputSourceEvidenceResourceLimit(t, query)
		})
	}
}

func assertInputSourceEvidenceResourceLimit(t *testing.T, query string) {
	t.Helper()
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Coverage.SyntaxComplete || result.Status != Incomplete || result.InputCoverage.State != "partial" || result.FieldAttributionCoverage.State != "partial" || result.Correlation.Outcome != "indeterminate" || result.Correlation.Coverage.State != "partial" {
		t.Fatalf("expanded evidence claimed complete: status=%s input=%s attribution=%s correlation=%s/%s", result.Status, result.InputCoverage.State, result.FieldAttributionCoverage.State, result.Correlation.Outcome, result.Correlation.Coverage.State)
	}
	if trace.sourceEvidenceBudget == nil || trace.sourceEvidenceBudget.failure == nil || trace.sourceEvidenceBudget.units > sourceEvidenceWorkLimit {
		t.Fatal("source evidence was materialized without a bounded reservation")
	}
	if result.Requirements.Coverage.Complete || result.Requirements.QueryStatus != Incomplete {
		t.Fatal("requirements claimed exhaustive evidence")
	}
	for _, coverage := range []InputCoverage{result.InputCoverage, result.FieldAttributionCoverage, result.Correlation.Coverage} {
		if len(coverage.Reasons) == 0 {
			t.Fatal("partial evidence has no located reason")
		}
	}
	retainedUnits := 0
	for _, fact := range trace.inputs {
		if len(fact.occurrence.UseSiteLocations) != len(fact.occurrence.UseSiteReferenceIDs) {
			t.Fatal("source context links lost pairing")
		}
		retainedUnits += 1 + len(fact.occurrence.UseSiteLocations)
	}
	if retainedUnits > sourceEvidenceWorkLimit {
		t.Fatalf("retained source context is unbounded: %d units", retainedUnits)
	}
	t.Logf("reserved units=%d, retained fact/context units=%d, trace facts=%d, source inputs=%d", trace.sourceEvidenceBudget.units, retainedUnits, len(trace.inputs), len(result.Inputs))
	if len(result.Inputs) != 1 || len(result.Inputs[0].Occurrences) == 0 || len(trace.inputs) > sourceEvidenceWorkLimit {
		t.Fatalf("known facts missing or unbounded: inputs=%d facts=%d", len(result.Inputs), len(trace.inputs))
	}
	found := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code != CodeAnalysisResourceLimit {
			continue
		}
		found = true
		if !strings.Contains(diagnostic.Message, "source-evidence") || strings.Contains(diagnostic.Message, "lexer") || !strings.HasPrefix(query[diagnostic.Location.Start.Offset:diagnostic.Location.End.Offset], "$v") || diagnostic.StageID == "" || diagnostic.ScopeID == "" {
			t.Fatalf("inaccurate or unlocated semantic limit: %#v", diagnostic)
		}
	}
	if !found {
		t.Fatal("missing source-evidence resource diagnostic")
	}
	repeated, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Inputs, repeated.Inputs) || !reflect.DeepEqual(result.InputCoverage, repeated.InputCoverage) {
		t.Fatal("budgeted discovery is not deterministic")
	}
	found = false
	for _, gap := range result.Requirements.Gaps {
		if gap.Code == CodeAnalysisResourceLimit {
			found = true
			if !strings.Contains(gap.Message, "source-evidence") || strings.Contains(gap.Message, "lexer") {
				t.Fatalf("inaccurate resource gap: %#v", gap)
			}
		}
	}
	if !found {
		t.Fatal("missing requirement resource gap")
	}
	if !reflect.DeepEqual(result.Inputs, result.Requirements.Inputs) || !reflect.DeepEqual(result.InputCoverage, result.Requirements.InputCoverage) || !reflect.DeepEqual(result.FieldAttributionCoverage, result.Requirements.FieldAttributionCoverage) || !reflect.DeepEqual(result.Correlation, result.Requirements.Correlation) {
		t.Fatal("result and requirements disagree")
	}
}
