package analysis

import (
	"encoding/json"
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
