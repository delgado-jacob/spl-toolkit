package corpus

import (
	"encoding/json"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func corpusDocument(id, text string) Entry {
	return Entry{ID: id, Origin: Origin{Kind: "inline"}, Document: &analysis.QueryDocument{Text: text}}
}

func TestSharedTargetRejectedBeforeAcquisition(t *testing.T) {
	target := ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"host", "host"}}}
	prepared, err := Prepare(ScanOptions{ValidationTarget: &target})
	if !IsInputError(err) || prepared != nil {
		t.Fatalf("invalid target reached acquisition: %v, %v", prepared, err)
	}
	badSchema := ValidationTarget{Kind: "json_schema", SchemaTarget: &validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":7}`)}}
	prepared, err = Prepare(ScanOptions{ValidationTarget: &badSchema})
	if !IsInputError(err) || prepared != nil {
		t.Fatalf("invalid compiled schema target was not a corpus input error: %v, %v", prepared, err)
	}
	// The loader is intentionally not part of corpus; callers must Prepare
	// before invoking any corpusio acquisition operation.
}

func TestAllAcquisitionsFail(t *testing.T) {
	p, err := Prepare(ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	input := Input{Selection: Selection{Mode: "manifest", Complete: true}, Entries: []Entry{{ID: "a", Origin: Origin{Kind: "file", RelativePath: "a.spl"}, Failure: &AcquisitionError{Code: "not_found", Phase: "open", ID: "a"}}, {ID: "b", Origin: Origin{Kind: "file", RelativePath: "b.spl"}, Failure: &AcquisitionError{Code: "permission_denied", Phase: "read", ID: "b"}}}}
	r, err := p.Scan(input)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != analysis.Incomplete || r.ExecutionComplete || r.Counts != (Counts{Selected: 2, AcquisitionFailed: 2}) || len(r.Entries) != 2 || r.Entries[0].Failure == nil || r.Entries[1].Failure == nil || r.Coverage.Syntax.Denominator != 0 || !r.Coverage.Schema.NotRequested {
		t.Fatalf("wrong all-failed report: %+v", r)
	}
	input.Selection.Complete = false
	input.Selection.TraversalFailures = []AcquisitionError{{Code: "traversal_failed", Path: "sub"}}
	input.Entries = nil
	r, err = p.Scan(input)
	if err != nil || r == nil || r.Status != analysis.Incomplete || r.ExecutionComplete || r.Counts.TraversalFailed != 1 {
		t.Fatalf("zero-entry traversal loss: %+v, %v", r, err)
	}
}

func TestCorpusTypedInputRejectsInvalidIdentityAndOptions(t *testing.T) {
	p, err := Prepare(ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []Entry{
		{ID: string([]byte{0xff}), Origin: Origin{Kind: "inline"}, Document: &analysis.QueryDocument{Text: "search host=web"}},
		{ID: "bad-options", Origin: Origin{Kind: "inline"}, Document: &analysis.QueryDocument{Text: "search host=web", Language: "other"}},
	} {
		r, err := p.Scan(Input{Selection: Selection{Mode: "inline", Complete: true}, Entries: []Entry{entry}})
		if r != nil || !IsInputError(err) {
			t.Fatalf("invalid typed input published a report: %+v, %v", r, err)
		}
	}
}

func TestCorpusMixedStatusCoverage(t *testing.T) {
	catalog := validation.FieldCatalog{Fields: []string{"host"}}
	p, err := Prepare(ScanOptions{ValidationTarget: &ValidationTarget{Kind: "field_list", Catalog: &catalog}})
	if err != nil {
		t.Fatal(err)
	}
	catalog.Fields[0] = "caller-mutated"
	in := Input{Selection: Selection{Mode: "manifest", Complete: true}, Entries: []Entry{corpusDocument("a", "search host=web"), corpusDocument("b", "search missing=x"), corpusDocument("c", "| mystery | table host"), {ID: "d", Failure: &AcquisitionError{Code: "not_found", Phase: "open"}}}}
	r, err := p.Scan(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.Mode != "field_list" || r.Status != analysis.Invalid || r.ExecutionComplete || r.Counts != (Counts{Selected: 4, Analyzed: 3, AcquisitionFailed: 1}) {
		t.Fatalf("status/counts: %+v", r)
	}
	if r.StatusCounts != (StatusCounts{Valid: 1, Invalid: 1, Incomplete: 1}) || len(r.CoverageReasons) == 0 || len(r.CommandCoverage) == 0 || len(r.ObservedReferenceForms) == 0 {
		t.Fatalf("analyzed status or canonical capability evidence missing: %+v", r)
	}
	if !sort.StringsAreSorted(r.CoverageReasons) {
		t.Fatalf("coverage reasons not deterministic: %v", r.CoverageReasons)
	}
	if n := sort.SearchStrings(r.CoverageReasons, analysis.CodeUnsupportedCommand); n == len(r.CoverageReasons) || r.CoverageReasons[n] != analysis.CodeUnsupportedCommand {
		t.Fatalf("unsupported command reason lost: %v", r.CoverageReasons)
	}
	var observedSearch, observedMystery bool
	for _, item := range r.CommandCoverage {
		if item.Command == "search" && item.Count == 2 && item.Declared && item.SemanticSupported {
			observedSearch = true
		}
		if item.Command == "mystery" && item.Count == 1 && !item.Declared && !item.SemanticSupported {
			observedMystery = true
		}
	}
	if !observedSearch || !observedMystery {
		t.Fatalf("command capability projection: %+v", r.CommandCoverage)
	}
	if r.Coverage.Syntax.Denominator != 3 || r.Coverage.Semantic.Denominator != 3 || r.Coverage.Schema.Denominator != 3 || r.Coverage.Semantic.Incomplete == 0 || r.Coverage.Schema.Incomplete == 0 {
		t.Fatalf("coverage: %+v", r.Coverage)
	}
	if r.Entries[1].Evaluation.FieldValidation.Status != analysis.Invalid || len(r.Entries[1].Evaluation.FieldValidation.Outcomes) == 0 || r.Entries[2].Evaluation.FieldValidation.Status != analysis.Incomplete || r.Entries[3].Failure == nil {
		t.Fatalf("findings lost: %+v", r.Entries)
	}
}

func TestCorpusEmptyCollectionsSerializeAsArrays(t *testing.T) {
	p, err := Prepare(ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Scan(Input{Selection: Selection{Mode: "directory", Complete: false, TraversalFailures: []AcquisitionError{{Code: "traversal_failed"}}}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"entries", "dependencies", "coverage_reasons", "command_coverage", "observed_reference_forms"} {
		if string(object[key]) != "[]" {
			t.Fatalf("%s is not an array: %s", key, object[key])
		}
	}
}

func TestCorpusSourceParityAndConcurrentUse(t *testing.T) {
	p, err := Prepare(ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	in := Input{Selection: Selection{Mode: "manifest", Complete: true}, Entries: []Entry{corpusDocument("a", "search host=web"), corpusDocument("b", "| inputlookup users")}}
	want, err := analysis.Analyze(*in.Entries[0].Document)
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Scan(in)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(r.Entries[0].Evaluation.Analysis)
	b, _ := json.Marshal(want)
	var av, bv any
	_ = json.Unmarshal(a, &av)
	_ = json.Unmarshal(b, &bv)
	if !reflect.DeepEqual(av, bv) {
		t.Fatalf("canonical report parity lost: %s\n%s", a, b)
	}
	if r.Coverage.Schema.Denominator != 0 || !r.Coverage.Schema.NotRequested || len(r.Dependencies) != 1 || r.Dependencies[0].Kind != "lookup" || r.Dependencies[0].Name != "users" || len(r.Dependencies[0].ReferenceIDs) == 0 {
		t.Fatalf("summary: %+v", r)
	}
	var wg sync.WaitGroup
	failures := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			next, e := p.Scan(in)
			if e != nil || !reflect.DeepEqual(next, r) {
				failures <- "repeat scan differed"
			}
		}()
	}
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Fatal(failure)
	}
}

func TestCorpusValidationSourceParity(t *testing.T) {
	document := analysis.QueryDocument{Text: "search host=web | table host*"}
	field := validation.FieldCatalog{Fields: []string{"host"}}
	schema := validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"host":{"type":"string"}},"additionalProperties":true}`)}
	cases := []struct {
		name      string
		target    ValidationTarget
		canonical func() (any, error)
	}{
		{"field", ValidationTarget{Kind: "field_list", Catalog: &field}, func() (any, error) { return validation.Validate(document, field) }},
		{"schema", ValidationTarget{Kind: "json_schema", SchemaTarget: &schema}, func() (any, error) { return validation.ValidateSchema(document, schema) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Prepare(ScanOptions{ValidationTarget: &tc.target})
			if err != nil {
				t.Fatal(err)
			}
			report, err := p.Scan(Input{Selection: Selection{Mode: "inline", Complete: true}, Entries: []Entry{{ID: "q", Origin: Origin{Kind: "inline"}, Document: &document}}})
			if err != nil {
				t.Fatal(err)
			}
			want, err := tc.canonical()
			if err != nil {
				t.Fatal(err)
			}
			var got any
			if tc.name == "field" {
				got = report.Entries[0].Evaluation.FieldValidation
			} else {
				got = report.Entries[0].Evaluation.SchemaValidation
			}
			gotRaw, _ := json.Marshal(got)
			wantRaw, _ := json.Marshal(want)
			var gotValue, wantValue any
			if err := json.Unmarshal(gotRaw, &gotValue); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(wantRaw, &wantValue); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotValue, wantValue) {
				t.Fatalf("canonical report diverged:\n%s\n%s", gotRaw, wantRaw)
			}
			if report.Coverage.Schema.Denominator != 1 || report.StatusCounts.Valid+report.StatusCounts.Invalid+report.StatusCounts.Incomplete != 1 {
				t.Fatalf("validation denominator: %+v", report)
			}
			if tc.name == "schema" && report.Coverage.Schema.Incomplete != 1 {
				t.Fatalf("open schema wildcard was reported complete: %+v", report.Coverage)
			}
		})
	}
}
