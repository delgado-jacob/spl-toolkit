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
			wantReport := explicitAliasReport(mode)
			if !reflect.DeepEqual(got, wantReport) {
				// Compare every public field, including canonical provenance and empty
				// collections, against the manually derived report below.
				a, b := reflect.ValueOf(*got), reflect.ValueOf(*wantReport)
				for i := 0; i < a.NumField(); i++ {
					if !reflect.DeepEqual(a.Field(i).Interface(), b.Field(i).Interface()) {
						t.Errorf("%s: got %+v; want %+v", a.Type().Field(i).Name, a.Field(i).Interface(), b.Field(i).Interface())
					}
				}
			}
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
				if change.Outcome != "applied" {
					t.Errorf("public outcome=%s want applied for candidate inclusion", change.Outcome)
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

// This expectation is hand-derived from the query's four lexical references:
// the search establishes src; sum reads that same source; total is an explicit
// aggregate output; table reads total through the aggregate's full origin chain.
// It never invokes analysis, selection, rendering, verification, or Rewrite.
func explicitAliasReport(mode Mode) *Result {
	loc := func(start, end int) analysis.Location {
		return analysis.Location{Start: analysis.Position{Offset: start, Line: 1, Column: start + 1}, End: analysis.Position{Offset: end, Line: 1, Column: end + 1}}
	}
	analyses := make([]*analysis.Result, 2)
	for i, expected := range []struct {
		text, name, digest string
		ends               [3]int
		starts             [3]int
		refs               [4][2]int
	}{
		{"search src=x | stats sum(src) AS total | table total", "src", "sha256:bf7f94f7e98b4ee7bdd32070fa2f2738ebb75df11ceb5bd9ade0a30d64fae951", [3]int{12, 38, 52}, [3]int{0, 15, 41}, [4][2]int{{7, 10}, {25, 28}, {33, 38}, {47, 52}}},
		{"search user=x | stats sum(user) AS total | table total", "user", "sha256:40e312d04c889b1fb1e8ef4f8996b5efe32606abd72ced915c8c0c3c15ee678b", [3]int{13, 40, 54}, [3]int{0, 16, 43}, [4][2]int{{7, 11}, {26, 30}, {35, 40}, {49, 54}}},
	} {
		empty := analysis.FieldState{Fields: []analysis.FieldBinding{}, Removed: []string{}, Open: true}
		source := analysis.FieldState{Fields: []analysis.FieldBinding{{Name: expected.name, OriginReferenceIDs: []string{"ref-0"}}}, Removed: []string{}, Open: true}
		total := analysis.FieldState{Fields: []analysis.FieldBinding{{Name: "total", OriginReferenceIDs: []string{"ref-2", "ref-1", "ref-0"}}}, Removed: []string{}}
		analyses[i] = &analysis.Result{
			SchemaVersion: 1, Document: analysis.QueryDocument{Text: expected.text, Language: "spl", Profile: "splunkd", Version: "current", SourceID: "source.spl"}, Status: analysis.Valid,
			Coverage: analysis.Coverage{SyntaxComplete: true, SemanticComplete: true, Reasons: []string{}},
			Stages: []analysis.Stage{
				{ID: "stage-0", Command: "search", Position: 0, ScopeID: "scope-0", Location: loc(expected.starts[0], expected.ends[0]), SemanticComplete: true},
				{ID: "stage-1", Command: "stats", Position: 1, ScopeID: "scope-0", Location: loc(expected.starts[1], expected.ends[1]), SemanticComplete: true},
				{ID: "stage-2", Command: "table", Position: 2, ScopeID: "scope-0", Location: loc(expected.starts[2], expected.ends[2]), SemanticComplete: true},
			},
			Scopes: []analysis.Scope{{ID: "scope-0", Kind: "root", Location: loc(0, expected.ends[2])}},
			References: []analysis.Reference{
				{ID: "ref-0", OriginalName: expected.name, NormalizedName: expected.name, Kind: "field", Role: "filter", StageID: "stage-0", ScopeID: "scope-0", Location: loc(expected.refs[0][0], expected.refs[0][1]), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
				{ID: "ref-1", OriginalName: expected.name, NormalizedName: expected.name, Kind: "field", Role: "read", StageID: "stage-1", ScopeID: "scope-0", Location: loc(expected.refs[1][0], expected.refs[1][1]), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{"ref-0"}},
				{ID: "ref-2", OriginalName: "total", NormalizedName: "total", Kind: "field", Role: "output", StageID: "stage-1", ScopeID: "scope-0", Location: loc(expected.refs[2][0], expected.refs[2][1]), Resolution: "exact", Binding: "not_applicable", OriginReferenceIDs: []string{"ref-1", "ref-0"}},
				{ID: "ref-3", OriginalName: "total", NormalizedName: "total", Kind: "field", Role: "read", StageID: "stage-2", ScopeID: "scope-0", Location: loc(expected.refs[3][0], expected.refs[3][1]), Resolution: "exact", Binding: "derived", OriginReferenceIDs: []string{"ref-2", "ref-1", "ref-0"}},
			},
			Lineage: []analysis.Lineage{
				{StageID: "stage-0", ScopeID: "scope-0", Before: empty, After: source, Transitions: []analysis.Transition{}},
				{StageID: "stage-1", ScopeID: "scope-0", Before: source, After: total, Transitions: []analysis.Transition{{Operation: "aggregate", Output: "total", InputReferenceIDs: []string{"ref-1"}, OutputReferenceID: "ref-2"}}},
				{StageID: "stage-2", ScopeID: "scope-0", Before: total, After: total, Transitions: []analysis.Transition{{Operation: "project", Output: "total", InputReferenceIDs: []string{"ref-3"}}}},
			},
			Dependencies: analysis.Dependencies{Indexes: []string{}, Sources: []string{}, SourceTypes: []string{}, Datasets: []string{}, Lookups: []string{}, DataModels: []string{}, Macros: []string{}}, Diagnostics: []analysis.Diagnostic{},
			Requirements: analysis.RequirementSet{
				SchemaVersion: 1,
				Query: analysis.RequirementQueryIdentity{
					SourceID: "source.spl", Language: "spl", Profile: "splunkd", Version: "current", QueryDigest: expected.digest,
				},
				CapabilityRevision: "sha256:dfb8cedde04204e0a876412fbe217e49405689b54ae7d7d8fc37bfcb7fb2335f",
				QueryStatus:        analysis.Valid,
				Coverage:           analysis.RequirementCoverage{Complete: true, Reasons: []string{}},
				Items: []analysis.RequirementItem{
					{ID: "req-1", Kind: "field", Identity: expected.name, Role: "filter", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []analysis.RequirementOccurrence{{ReferenceID: "ref-0", OriginalName: expected.name, Binding: "source", StageID: "stage-0", ScopeID: "scope-0", Location: loc(expected.refs[0][0], expected.refs[0][1])}}},
					{ID: "req-2", Kind: "field", Identity: expected.name, Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []analysis.RequirementOccurrence{{ReferenceID: "ref-1", OriginalName: expected.name, Binding: "source", StageID: "stage-1", ScopeID: "scope-0", Location: loc(expected.refs[1][0], expected.refs[1][1])}}},
				},
				Gaps:        []analysis.RequirementGap{},
				Diagnostics: []analysis.Diagnostic{},
			},
		}
	}
	a0, b0, a1, b1, alias, consumer := loc(7, 10), loc(7, 11), loc(25, 28), loc(26, 30), loc(33, 38), loc(47, 52)
	text, committed := analyses[0].Document.Text, false
	if mode == Apply {
		text, committed = analyses[1].Document.Text, true
	}
	return &Result{SchemaVersion: 1, Document: analyses[0].Document, Mode: mode, Status: analysis.Valid,
		Coverage:     Coverage{SyntaxComplete: true, SemanticComplete: true, RewriteComplete: true, Reasons: []string{}},
		OriginalText: analyses[0].Document.Text, CandidateText: analyses[1].Document.Text, Text: text, Committed: committed,
		Changes: []Change{
			{Outcome: "applied", Reason: "matched", GroupID: "group-0", RuleIDs: []string{"src-user"}, OriginalReferenceIDs: []string{"ref-0"}, CandidateReferenceIDs: []string{"ref-0"}, OriginalLocation: &a0, CandidateLocation: &b0, OldText: "src", NewText: "user", CandidateApplied: true, Committed: committed},
			{Outcome: "applied", Reason: "matched", GroupID: "group-0", RuleIDs: []string{"src-user"}, OriginalReferenceIDs: []string{"ref-1"}, CandidateReferenceIDs: []string{"ref-1"}, OriginalLocation: &a1, CandidateLocation: &b1, OldText: "src", NewText: "user", CandidateApplied: true, Committed: committed},
		},
		RuleEvaluations: []RuleEvaluation{
			{RuleID: "src-user", Outcome: "proposed", Reason: "matched", ReferenceIDs: []string{"ref-0"}, Location: &a0},
			{RuleID: "src-user", Outcome: "proposed", Reason: "matched", ReferenceIDs: []string{"ref-1"}, Location: &a1},
			{RuleID: "keep-alias", Outcome: "skipped", Reason: "unsupported_reference", ReferenceIDs: []string{"ref-2"}, Location: &alias},
			{RuleID: "keep-alias", Outcome: "skipped", Reason: "unsupported_reference", ReferenceIDs: []string{"ref-3"}, Location: &consumer},
		}, OriginalAnalysis: analyses[0], CandidateAnalysis: analyses[1],
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

// The corpus detects changed public publication policy, linked edit loss, byte
// damage outside edit intervals, or IDs that no longer resolve in the report.
func TestRewriteCorpus(t *testing.T) {
	data, err := os.ReadFile("../../testdata/rewrite/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	type expectation struct {
		Status           analysis.Status `json:"status"`
		CandidateText    string          `json:"candidate_text"`
		Text             string          `json:"text"`
		Committed        bool            `json:"committed"`
		SyntaxComplete   bool            `json:"syntax_complete"`
		SemanticComplete bool            `json:"semantic_complete"`
		RewriteComplete  bool            `json:"rewrite_complete"`
		ChangeOutcomes   []string        `json:"change_outcomes"`
	}
	var cases []struct {
		Name    string          `json:"name"`
		Request json.RawMessage `json:"request"`
		Want    expectation     `json:"want"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("missing durable cases")
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			request, err := DecodeRequest(tc.Request)
			if err != nil {
				t.Fatal(err)
			}
			got := requireRewrite(t, request)
			actual := expectation{Status: got.Status, CandidateText: got.CandidateText, Text: got.Text, Committed: got.Committed,
				SyntaxComplete: got.Coverage.SyntaxComplete, SemanticComplete: got.Coverage.SemanticComplete, RewriteComplete: got.Coverage.RewriteComplete, ChangeOutcomes: []string{}}
			originalEnd, candidateEnd := 0, 0
			for _, change := range got.Changes {
				actual.ChangeOutcomes = append(actual.ChangeOutcomes, change.Outcome)
				if change.Committed != (got.Committed && change.CandidateApplied) {
					t.Errorf("change commitment contradicts publication: %+v", change)
				}
				if !change.CandidateApplied {
					continue
				}
				if change.OriginalLocation == nil || change.CandidateLocation == nil {
					t.Fatal("included edit lost locations")
				}
				a, b := change.OriginalLocation, change.CandidateLocation
				if a.Start.Offset < originalEnd || b.Start.Offset < candidateEnd || a.End.Offset > len(got.OriginalText) || b.End.Offset > len(got.CandidateText) {
					t.Fatalf("invalid or unordered audit interval: %+v", change)
				}
				if got.OriginalText[originalEnd:a.Start.Offset] != got.CandidateText[candidateEnd:b.Start.Offset] || got.OriginalText[a.Start.Offset:a.End.Offset] != change.OldText || got.CandidateText[b.Start.Offset:b.End.Offset] != change.NewText {
					t.Errorf("audit cannot reconstruct byte-preserving edits: %+v", change)
				}
				originalEnd, candidateEnd = a.End.Offset, b.End.Offset
				for i, ids := range [][]string{change.OriginalReferenceIDs, change.CandidateReferenceIDs} {
					if len(ids) == 0 {
						t.Fatal("included edit lost canonical identities")
					}
					refs := got.OriginalAnalysis.References
					if i == 1 {
						refs = got.CandidateAnalysis.References
					}
					for _, id := range ids {
						found := false
						for _, ref := range refs {
							found = found || ref.ID == id
						}
						if !found {
							t.Errorf("unresolvable reference ID %q", id)
						}
					}
				}
			}
			if got.OriginalText[originalEnd:] != got.CandidateText[candidateEnd:] {
				t.Error("text changed after last included edit")
			}
			if !reflect.DeepEqual(actual, tc.Want) {
				t.Fatalf("got %+v; want %+v", actual, tc.Want)
			}
			if got.OriginalText != request.Document.Text || got.Document != request.Document || got.CandidateAnalysis.Document.SourceID != request.Document.SourceID {
				t.Error("document source identity changed")
			}
		})
	}
}
