package rewrite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// Returning early on syntax damage or reordering by dialect loses valid results.
func TestRewriteBatchOrderAndAtomicErrors(t *testing.T) {
	request := BatchRequest{SchemaVersion: 1, Mode: Apply, Rules: rewriteRequest("").Rules,
		Documents: []analysis.QueryDocument{
			{Text: "table src", SourceID: "first"},
			{Text: "FROM main SELECT src", Language: "spl2", SourceID: "second"},
			{Text: "eval =", SourceID: "third"},
			{Text: "table src | mystery foo", SourceID: "fourth"},
		}}
	got, err := RewriteBatch(request)
	if err != nil || got == nil {
		t.Fatalf("batch: %v, %v", got, err)
	}
	if got.SchemaVersion != 1 || got.Status != analysis.Invalid || len(got.Reports) != 4 {
		t.Fatalf("batch shape/status: %+v", got)
	}
	for i, document := range request.Documents {
		want := requireRewrite(t, Request{SchemaVersion: 1, Mode: Apply, Document: document, Rules: request.Rules})
		if !reflect.DeepEqual(got.Reports[i], want) || got.Reports[i].Document.SourceID != document.SourceID {
			t.Errorf("report %d lost canonical ordering or single-request policy", i)
		}
	}
	if got.Reports[0].Text != "table user" || !got.Reports[0].Committed || got.Reports[1].Text != "FROM main SELECT user" || !got.Reports[1].Committed || got.Reports[2].Status != analysis.Invalid || got.Reports[2].Committed || got.Reports[3].Status != analysis.Incomplete {
		for i, report := range got.Reports {
			t.Logf("report %d: status=%s committed=%v text=%q coverage=%+v", i, report.Status, report.Committed, report.Text, report.Coverage)
		}
		t.Fatal("mixed outcomes differ")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*BatchRequest)
	}{
		{"empty", func(r *BatchRequest) { r.Documents = nil }},
		{"bad final selector", func(r *BatchRequest) { r.Documents[3].Language = "sql" }},
		{"bad final unicode", func(r *BatchRequest) { r.Documents[3].Text = string([]byte{0xff}) }},
		{"mode", func(r *BatchRequest) { r.Mode = "commit" }},
		{"rules", func(r *BatchRequest) { r.Rules = append(r.Rules, r.Rules[0]) }},
		{"target shape", func(r *BatchRequest) { r.ValidationTarget = &ValidationTarget{Kind: "json_schema"} }},
		{"target semantics", func(r *BatchRequest) { r.ValidationTarget = schemaRewriteTarget(`{"type":42}`) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := request
			bad.Documents = append([]analysis.QueryDocument{}, request.Documents...)
			tc.mutate(&bad)
			formed := 0
			report, err := rewriteBatch(bad, func(document analysis.QueryDocument, rules []Rule, probes []analysis.RewriteFactProbe) (*pendingRewrite, error) {
				formed++
				return formCandidate(document, rules, probes)
			}, validation.ValidateBatch, validation.ValidateSchemaBatch)
			if report != nil || !IsInputError(err) {
				t.Fatalf("partial report or lost input classification: %+v, %v", report, err)
			}
			// Shape errors must be found before preparing even the first candidate.
			// Semantic compilation belongs to the real destination batch validator.
			if tc.name != "target semantics" && formed != 0 {
				t.Fatalf("shape error escaped the preparation phase: formed=%d", formed)
			}
		})
	}
}

// A valid destination report cannot erase unresolved mapping alternatives, and
// one query's whole-request failure must not prevent another query committing.
func TestRewriteBatchSafeBesideAmbiguous(t *testing.T) {
	request := BatchRequest{SchemaVersion: 1, Mode: Apply,
		Documents: []analysis.QueryDocument{
			{Text: "search src=x other=y | table src other", SourceID: "safe-and-ambiguous"},
			{Text: "FROM main SELECT src", Language: "spl2", SourceID: "safe"},
			{Text: "table missing", SourceID: "destination-missing"},
		},
		Rules: []Rule{
			{ID: "safe", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")},
			{ID: "a", Kind: "field", Source: conditionIdentity("other"), Target: conditionIdentity("a")},
			{ID: "b", Kind: "field", Source: conditionIdentity("other"), Target: conditionIdentity("b")},
		},
		ValidationTarget: &ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"user", "other"}}},
	}
	got, err := RewriteBatch(request)
	if err != nil || got == nil || got.Status != analysis.Invalid || len(got.Reports) != 3 {
		t.Fatalf("batch: %+v, %v", got, err)
	}
	for i, want := range []struct {
		text       string
		status     analysis.Status
		validation analysis.Status
		committed  bool
	}{
		{"search user=x other=y | table user other", analysis.Incomplete, analysis.Valid, true},
		{"FROM main SELECT user", analysis.Valid, analysis.Valid, true},
		{"table missing", analysis.Invalid, analysis.Invalid, false},
	} {
		report := got.Reports[i]
		if report.Text != want.text || report.Status != want.status || report.Committed != want.committed || report.CandidateValidation.FieldList.Status != want.validation || report.Document.SourceID != request.Documents[i].SourceID {
			t.Errorf("report %d: status=%s validation=%s committed=%v text=%q", i, report.Status, report.CandidateValidation.FieldList.Status, report.Committed, report.Text)
		}
	}
	if got.Reports[0].Coverage.RewriteComplete || len(got.Reports[0].Changes) != 6 || len(got.Reports[2].Changes) != 0 {
		t.Fatal("mapping uncertainty or no-op audit was erased")
	}
}

// Status precedence and no-op target checking must not depend on document order
// or whether any rule produced an edit.
func TestRewriteBatchStatusAndNoopTargets(t *testing.T) {
	for _, tc := range []struct {
		queries []string
		status  analysis.Status
	}{
		{[]string{"table src", "table other"}, analysis.Valid},
		{[]string{"table src", "table other | mystery"}, analysis.Incomplete},
		{[]string{"table other | mystery", "table src"}, analysis.Incomplete},
		{[]string{"eval =", "table other | mystery", "table src"}, analysis.Invalid},
		{[]string{"table src", "table other | mystery", "eval ="}, analysis.Invalid},
	} {
		request := BatchRequest{SchemaVersion: 1}
		for _, query := range tc.queries {
			request.Documents = append(request.Documents, analysis.QueryDocument{Text: query})
		}
		got, err := RewriteBatch(request)
		if err != nil || got == nil || got.Status != tc.status {
			t.Fatalf("aggregate %v: %+v, %v", tc.queries, got, err)
		}
	}
	for _, target := range []*ValidationTarget{
		{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"user"}}},
		schemaRewriteTarget(`{"type":"object","properties":{"user":true},"additionalProperties":false}`),
		schemaRewriteTarget(`{"$ref":"https://offline.invalid/missing"}`),
		schemaRewriteTarget(`{"type":42}`),
	} {
		request := BatchRequest{SchemaVersion: 1, Mode: Apply, Documents: []analysis.QueryDocument{{Text: "table src"}}, ValidationTarget: target}
		got, err := RewriteBatch(request)
		if target.Kind == "json_schema" && string(target.SchemaTarget.Schema) == `{"type":42}` {
			if got != nil || !IsInputError(err) || !validation.IsInputError(err) {
				t.Fatalf("no-op semantic target error escaped: %+v, %v", got, err)
			}
			continue
		}
		if err != nil || got == nil || got.Status == analysis.Valid || got.Reports[0].Committed || got.Reports[0].Text != "table src" || got.Reports[0].CandidateValidation == nil || len(got.Reports[0].Changes) != 0 {
			t.Fatalf("no-op skipped requested validation: %+v, %v", got, err)
		}
	}
}

// Retaining caller slices, condition bytes, or schema resources after request
// preparation changes the later document's result. Shared report storage breaks
// repeated calls, while shared analysis state breaks the concurrent equality.
func TestRewriteConcurrentOwnership(t *testing.T) {
	for _, kind := range []string{"field_list", "json_schema", "ocsf"} {
		t.Run("prepared copies "+kind, func(t *testing.T) {
			rule := rewriteRequest("").Rules[0]
			factIdentity := conditionIdentity("EventCode")
			rule.When = &Condition{All: []Condition{{Fact: "literal", Kind: "field", Identity: &factIdentity, Operator: "equals", Value: json.RawMessage(`1`)}}}
			target := &ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"EventCode", "user"}, OptionalFields: []string{"optional"}, Identity: "owned", Version: "2"}}
			if kind == "json_schema" {
				target = schemaRewriteTarget(`{"$ref":"https://local.test/root"}`)
				target.SchemaTarget.Resources = map[string]json.RawMessage{"https://local.test/root": json.RawMessage(`{"type":"object","properties":{"EventCode":true,"user":true},"additionalProperties":false}`)}
			} else if kind == "ocsf" {
				target = ocsfRewriteTarget(t)
				rule.Target = conditionIdentity("time")
			}
			request := BatchRequest{SchemaVersion: 1, Mode: Apply, Documents: []analysis.QueryDocument{{Text: "search EventCode=1 src=x"}, {Text: "search EventCode=1 src=x"}}, Rules: []Rule{rule}, ValidationTarget: target}
			want, err := RewriteBatch(request)
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			got, err := rewriteBatch(request, func(document analysis.QueryDocument, rules []Rule, probes []analysis.RewriteFactProbe) (*pendingRewrite, error) {
				calls++
				if calls == 1 {
					request.Documents[1].Text = "eval ="
					*request.Rules[0].Source.Name = "mutated"
					*request.Rules[0].Target.Name = "mutated"
					*request.Rules[0].When.All[0].Identity.Name = "mutated"
					request.Rules[0].When.All[0].Value[0] = '2'
					request.Rules[0].When.All = nil
					if kind == "field_list" {
						target.Catalog.Fields[0] = "mutated"
						target.Catalog.OptionalFields[0] = "mutated"
						target.Catalog.Identity = "mutated"
					} else if kind == "json_schema" {
						target.SchemaTarget.Schema[0] = '['
						target.SchemaTarget.Resources["https://local.test/root"][0] = '['
						delete(target.SchemaTarget.Resources, "https://local.test/root")
					} else {
						target.SchemaTarget.Catalog[0] = '['
						target.SchemaTarget.Selection.Version = "mutated"
						target.SchemaTarget.Selection.Class = "mutated"
					}
				}
				return formCandidate(document, rules, probes)
			}, validation.ValidateBatch, validation.ValidateSchemaBatch)
			if err != nil || calls != 2 || !reflect.DeepEqual(got, want) {
				t.Fatalf("caller mutation leaked after preparation: err=%v calls=%d", err, calls)
			}
		})
	}
	t.Run("repeated and concurrent", func(t *testing.T) {
		request := rewriteRequest("search src=x | stats sum(src) AS total | table total")
		request.ValidationTarget = schemaRewriteTarget(`{"type":"object","properties":{"user":true},"additionalProperties":false}`)
		batchRequest := BatchRequest{SchemaVersion: 1, Mode: Apply, Rules: request.Rules, ValidationTarget: request.ValidationTarget,
			Documents: []analysis.QueryDocument{request.Document, {Text: "FROM main SELECT src", Language: "spl2", SourceID: "second"}}}
		single := requireRewrite(t, request)
		batch, err := RewriteBatch(batchRequest)
		if err != nil {
			t.Fatal(err)
		}
		wantSingle, _ := json.Marshal(single)
		wantBatch, _ := json.Marshal(batch)
		// Results can be edited by a caller without contaminating a later call.
		single.Changes[0].RuleIDs[0] = "mutated"
		single.Changes[0].OriginalLocation.Start.Offset = 999
		single.OriginalAnalysis.References[0].NormalizedName = "mutated"
		single.CandidateValidation.Schema.Target.Limitations = []string{"mutated"}
		batch.Reports[0].CandidateAnalysis.Document.Text = "mutated"
		check := func() error {
			one, err := Rewrite(request)
			if err != nil {
				return err
			}
			many, err := RewriteBatch(batchRequest)
			if err != nil {
				return err
			}
			a, err := json.Marshal(one)
			if err != nil {
				return err
			}
			b, err := json.Marshal(many)
			if err != nil {
				return err
			}
			if !bytes.Equal(a, wantSingle) || !bytes.Equal(b, wantBatch) {
				return fmt.Errorf("non-equivalent JSON reports")
			}
			return nil
		}
		for i := 0; i < 3; i++ {
			if err := check(); err != nil {
				t.Fatal(err)
			}
		}
		var wg sync.WaitGroup
		failures := make(chan error, 8)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); failures <- check() }()
		}
		wg.Wait()
		close(failures)
		for err := range failures {
			if err != nil {
				t.Error(err)
			}
		}
	})
}

// Revalidating each query or attaching reports by status breaks call count,
// destination identity, and report pointer preservation at this real boundary.
func TestRewriteBatchDestinationValidation(t *testing.T) {
	for _, target := range []*ValidationTarget{
		{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"user"}, Identity: "destination", Version: "2"}},
		schemaRewriteTarget(`{"type":"object","properties":{"user":true},"additionalProperties":false}`),
		ocsfRewriteTarget(t),
	} {
		t.Run(target.Kind, func(t *testing.T) {
			rules := rewriteRequest("").Rules
			name := "user"
			if target.Kind == "ocsf" {
				name = "time"
				rules[0].Target = conditionIdentity(name)
			}
			request := BatchRequest{SchemaVersion: 1, Mode: Apply, Rules: rules, ValidationTarget: target,
				Documents: []analysis.QueryDocument{{Text: "table src", SourceID: "a"}, {Text: "FROM main SELECT src", Language: "spl2", SourceID: "b"}, {Text: "eval =", SourceID: "bad"}}}
			calls := 0
			var fields *validation.BatchReport
			var schemas *validation.SchemaBatchReport
			check := func(documents []analysis.QueryDocument) {
				t.Helper()
				calls++
				if len(documents) != 3 || documents[0].Text != "table "+name || documents[1].Text != "FROM main SELECT "+name || documents[2].Text != "eval =" || documents[0].SourceID != "a" || documents[1].SourceID != "b" || documents[2].SourceID != "bad" {
					t.Fatalf("validation saw original, synthetic, or reordered documents: %+v", documents)
				}
			}
			got, err := rewriteBatch(request, formCandidate,
				func(documents []analysis.QueryDocument, catalog validation.FieldCatalog) (*validation.BatchReport, error) {
					check(documents)
					var err error
					fields, err = validation.ValidateBatch(documents, catalog)
					return fields, err
				},
				func(documents []analysis.QueryDocument, schema validation.SchemaTarget) (*validation.SchemaBatchReport, error) {
					check(documents)
					var err error
					schemas, err = validation.ValidateSchemaBatch(documents, schema)
					return schemas, err
				})
			if err != nil || got == nil || calls != 1 || got.Status != analysis.Invalid || len(got.Reports) != 3 {
				t.Fatalf("batch validation: got=%+v err=%v calls=%d", got, err, calls)
			}
			for i, report := range got.Reports {
				if report.CandidateValidation == nil || report.CandidateValidation.Kind != target.Kind || report.Committed != (i < 2) {
					t.Fatalf("report %d destination policy: %+v", i, report)
				}
				if target.Kind == "field_list" {
					if report.CandidateValidation.FieldList != fields.Reports[i] {
						t.Errorf("field report %d was rebuilt", i)
					}
				} else if report.CandidateValidation.Schema != schemas.Reports[i] {
					t.Errorf("schema report %d was rebuilt", i)
				}
			}
		})
	}
}

// Internal failures must not leak already formed or validated candidate reports.
func TestRewriteBatchInternalErrors(t *testing.T) {
	sentinel := errors.New("internal operation failed")
	for _, stage := range []string{"candidate", "field_list", "json_schema"} {
		t.Run(stage, func(t *testing.T) {
			request := BatchRequest{SchemaVersion: 1, Mode: Apply, Rules: rewriteRequest("").Rules,
				Documents: []analysis.QueryDocument{{Text: "table src"}, {Text: "table src"}}}
			if stage == "field_list" {
				request.ValidationTarget = &ValidationTarget{Kind: stage, Catalog: &validation.FieldCatalog{Fields: []string{"user"}}}
			} else if stage == "json_schema" {
				request.ValidationTarget = schemaRewriteTarget(`true`)
			}
			formed := 0
			got, err := rewriteBatch(request,
				func(document analysis.QueryDocument, rules []Rule, probes []analysis.RewriteFactProbe) (*pendingRewrite, error) {
					formed++
					if stage == "candidate" && formed == 2 {
						return nil, sentinel
					}
					return formCandidate(document, rules, probes)
				},
				func([]analysis.QueryDocument, validation.FieldCatalog) (*validation.BatchReport, error) {
					return nil, sentinel
				},
				func([]analysis.QueryDocument, validation.SchemaTarget) (*validation.SchemaBatchReport, error) {
					return nil, sentinel
				})
			if got != nil || !errors.Is(err, sentinel) || IsInputError(err) || formed != 2 {
				t.Fatalf("partial result or changed error: got=%+v err=%v formed=%d", got, err, formed)
			}
		})
	}
}
