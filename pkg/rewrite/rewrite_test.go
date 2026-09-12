package rewrite

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func rewriteRequest(text string) Request {
	return Request{SchemaVersion: 1, Mode: Apply, Document: analysis.QueryDocument{Text: text, SourceID: "source.spl"},
		Rules: []Rule{{ID: "src-user", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")}}}
}

func requireRewrite(t *testing.T, request Request) *Result {
	t.Helper()
	got, err := Rewrite(request)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("missing rewrite report")
	}
	return got
}

// Publishing preview text or renaming the explicit derived alias breaks this test.
func TestRewritePreviewApply(t *testing.T) {
	const original = "search src=x | stats sum(src) AS total | table total"
	const candidate = "search user=x | stats sum(user) AS total | table total"
	for _, mode := range []Mode{Preview, Apply} {
		t.Run(string(mode), func(t *testing.T) {
			request := rewriteRequest(original)
			request.Mode = mode
			request.Rules = append(request.Rules, Rule{ID: "keep-alias", Kind: "field", Source: conditionIdentity("total"), Target: conditionIdentity("other")})
			got := requireRewrite(t, request)
			wantText := original
			if mode == Apply {
				wantText = candidate
			}
			if got.SchemaVersion != 1 || got.Mode != mode || got.Status != analysis.Valid || got.Text != wantText || got.OriginalText != original || got.CandidateText != candidate || got.Committed != (mode == Apply) {
				t.Fatalf("publication policy: %+v", got)
			}
			if got.Document.Text != original || got.Document.Language != "spl" || got.Document.Profile != "splunkd" || got.Document.Version != "current" || got.Document.SourceID != "source.spl" {
				t.Fatalf("document selectors or identity lost: %+v", got.Document)
			}
			if !got.Coverage.SyntaxComplete || !got.Coverage.SemanticComplete || !got.Coverage.RewriteComplete || len(got.Coverage.Reasons) != 0 || got.CandidateValidation != nil {
				t.Fatalf("coverage: %+v", got.Coverage)
			}
			if got.OriginalAnalysis == nil || got.CandidateAnalysis == nil || got.OriginalAnalysis.Document != got.Document || got.CandidateAnalysis.Document.Text != candidate || got.CandidateAnalysis.Document.SourceID != got.Document.SourceID {
				t.Fatal("canonical analyses missing or wrong")
			}
			if len(got.Changes) != 2 || len(got.RuleEvaluations) != 4 {
				t.Fatalf("changes/evaluations: %+v / %+v", got.Changes, got.RuleEvaluations)
			}
			for i, change := range got.Changes {
				if change.OldText != "src" || change.NewText != "user" || change.Reason != "matched" || !change.CandidateApplied || change.Committed != got.Committed || !reflect.DeepEqual(change.RuleIDs, []string{"src-user"}) {
					t.Errorf("change %d: %+v", i, change)
				}
				wantOutcome := "proposed"
				if mode == Apply {
					wantOutcome = "applied"
				}
				if change.Outcome != wantOutcome {
					t.Errorf("outcome=%s want %s", change.Outcome, wantOutcome)
				}
				if change.OriginalLocation == nil || change.CandidateLocation == nil || len(change.OriginalReferenceIDs) != 1 || len(change.CandidateReferenceIDs) != 1 {
					t.Fatalf("missing edit correspondence: %+v", change)
				}
				originalOffsets, candidateOffsets := []int{7, 25}, []int{7, 26}
				if change.OriginalLocation.Start.Offset != originalOffsets[i] || change.OriginalLocation.End.Offset != originalOffsets[i]+3 || change.CandidateLocation.Start.Offset != candidateOffsets[i] || change.CandidateLocation.End.Offset != candidateOffsets[i]+4 {
					t.Errorf("incorrect byte translation: %+v", change)
				}
			}
			for _, ref := range got.CandidateAnalysis.References {
				if ref.NormalizedName == "other" || ref.NormalizedName == "src" {
					t.Errorf("alias or source contract broken: %+v", ref)
				}
				if ref.NormalizedName == "total" && ref.Role == "read" && (ref.Binding != "derived" || len(ref.OriginReferenceIDs) == 0) {
					t.Errorf("alias consumer lost binding: %+v", ref)
				}
			}
		})
	}
}

func TestRewriteUnchanged(t *testing.T) {
	for _, tc := range []struct{ name, source, target, reason string }{
		{"empty rules", "", "", ""}, {"no match", "absent", "user", ReasonNoMatch}, {"same identity", "src", "src", ReasonNoChange},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := rewriteRequest("search src=x\r\n| table src")
			request.Rules = nil
			if tc.source != "" {
				request.Rules = []Rule{{ID: "rule", Kind: "field", Source: conditionIdentity(tc.source), Target: conditionIdentity(tc.target)}}
			}
			got := requireRewrite(t, request)
			if got.Text != request.Document.Text || got.CandidateText != request.Document.Text || got.Committed || len(got.Changes) != 0 || got.Status != analysis.Valid || !got.Coverage.RewriteComplete {
				t.Fatalf("no-op fabricated an apply: %+v", got)
			}
			for _, evaluation := range got.RuleEvaluations {
				if evaluation.Outcome != "skipped" || evaluation.Reason != tc.reason {
					t.Errorf("ordinary skip: %+v", evaluation)
				}
			}
		})
	}
}

func schemaRewriteTarget(schema string) *ValidationTarget {
	return &ValidationTarget{Kind: "json_schema", SchemaTarget: &validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(schema)}}
}

func ocsfRewriteTarget(t *testing.T) *ValidationTarget {
	t.Helper()
	f, err := os.Open("../../testdata/schemas/ocsf/1.6.0/base.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137" {
		t.Fatalf("accepted OCSF fixture hash changed: %s", got)
	}
	return &ValidationTarget{Kind: "ocsf", SchemaTarget: &validation.SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &validation.OCSFSelection{Version: "1.6.0", Class: "authentication"}}}
}

// Validating the original source or accepting incomplete destination evidence
// would respectively reject the migration or incorrectly publish it.
func TestRewriteDestinationValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target *ValidationTarget
		status analysis.Status
	}{
		{"field valid", &ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"user"}, Identity: "destination", Version: "2"}}, analysis.Valid},
		{"field missing", &ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"other"}}}, analysis.Invalid},
		{"schema valid", schemaRewriteTarget(`{"type":"object","properties":{"user":true},"additionalProperties":false}`), analysis.Valid},
		{"schema missing", schemaRewriteTarget(`{"type":"object","additionalProperties":false}`), analysis.Invalid},
		{"schema conditional", schemaRewriteTarget(`{"anyOf":[{"type":"object","properties":{"user":true},"additionalProperties":false},{"type":"object","additionalProperties":false}]}`), analysis.Incomplete},
		{"schema unresolved", schemaRewriteTarget(`{"$ref":"https://offline.invalid/missing"}`), analysis.Incomplete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := rewriteRequest("table src")
			request.ValidationTarget = tc.target
			got := requireRewrite(t, request)
			if got.Status != tc.status || got.Committed != (tc.status == analysis.Valid) || got.CandidateText != "table user" || got.CandidateValidation == nil || got.Coverage.Validation == nil {
				t.Fatalf("destination gate: %+v", got)
			}
			candidate := got.Document
			candidate.Text = "table user"
			if tc.target.Kind == "field_list" {
				want, err := validation.Validate(candidate, *tc.target.Catalog)
				if err != nil || !reflect.DeepEqual(got.CandidateValidation.FieldList, want) || !reflect.DeepEqual(*got.Coverage.Validation, want.Coverage) {
					t.Fatalf("canonical field report changed: %v", err)
				}
			} else {
				want, err := validation.ValidateSchema(candidate, *tc.target.SchemaTarget)
				if err != nil || !reflect.DeepEqual(got.CandidateValidation.Schema, want) || !reflect.DeepEqual(*got.Coverage.Validation, want.Coverage) {
					t.Fatalf("canonical schema report changed: %v", err)
				}
			}
			if tc.status != analysis.Valid {
				assertRewriteRollback(t, got)
			}
		})
	}
	t.Run("real local OCSF", func(t *testing.T) {
		request := rewriteRequest("table src")
		request.Rules[0].Target = conditionIdentity("time")
		request.ValidationTarget = ocsfRewriteTarget(t)
		got := requireRewrite(t, request)
		if !got.Committed || got.Status != analysis.Valid || got.Text != "table time" || got.CandidateValidation.Schema.Target.Version != "1.6.0" || got.CandidateValidation.Schema.Outcomes[0].Outcome != "required" {
			t.Fatalf("real OCSF destination gate: %+v", got)
		}
	})
	for _, query := range []string{"table src", "eval ="} {
		t.Run("no-op target preparation "+query, func(t *testing.T) {
			request := rewriteRequest(query)
			request.Rules = nil
			request.ValidationTarget = schemaRewriteTarget(`{"type":42}`)
			got, err := Rewrite(request)
			if got != nil || !IsInputError(err) || !validation.IsInputError(err) {
				t.Fatalf("semantic target error not retained: report=%+v err=%v", got, err)
			}
			request.ValidationTarget = &ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"user"}}}
			got = requireRewrite(t, request)
			if got.Committed || got.Status != analysis.Invalid || got.CandidateValidation == nil || len(got.Changes) != 0 {
				t.Fatalf("no-op skipped destination validation: %+v", got)
			}
		})
	}
}

func assertRewriteRollback(t *testing.T, got *Result) {
	t.Helper()
	if got.Committed || got.Text != got.OriginalText || got.CandidateText == got.OriginalText || got.CandidateAnalysis.Document.Text != got.CandidateText {
		t.Fatalf("failed gate published text or discarded candidate: %+v", got)
	}
	for _, change := range got.Changes {
		if change.Committed {
			t.Errorf("failed gate committed change: %+v", change)
		}
		if change.CandidateApplied && (change.Outcome != "skipped" || change.Reason != ReasonPostVerificationFailed || change.CandidateLocation == nil) {
			t.Errorf("rollback audit lost inclusion/coordinates: %+v", change)
		}
	}
}
