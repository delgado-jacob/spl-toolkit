package impact

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpusio"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func TestImpactStrictInlineRequests(t *testing.T) {
	goodSchema := []byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"table host"}}],"before_target":{"kind":"field_list","catalog":{"fields":["host"]}},"after_target":{"kind":"field_list","catalog":{"fields":[]}}}`)
	r, err := DecodeSchemaRequest(goodSchema)
	if err != nil || len(r.Input.Entries) != 1 {
		t.Fatalf("valid inline request: %+v %v", r, err)
	}
	if _, err := CompareSchemas(r); err != nil {
		t.Fatal(err)
	}
	for _, raw := range [][]byte{
		[]byte(`{"schema_version":1,"schema_version":1,"documents":[],"before_target":{},"after_target":{}}`),
		[]byte(`{"schema_version":1,"documents":[],"before_target":{},"after_target":{},"path":"q.spl"}`),
		[]byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"x"},"failure":{"code":"x"}}],"before_target":{},"after_target":{}}`),
		[]byte(`{"schema_version":1,"documents":null,"before_target":{},"after_target":{}}`),
	} {
		if _, err := DecodeSchemaRequest(raw); !IsInputError(err) {
			t.Fatalf("malformed schema request accepted: %s, %v", raw, err)
		}
	}
	goodMapping := []byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"table src"}}],"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[]}}`)
	m, err := DecodeMappingRequest(goodMapping)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompareMappings(m); err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeMappingRequest([]byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"table src"}}],"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[]},"before_target":null}`)); !IsInputError(err) {
		t.Fatalf("null target accepted: %v", err)
	}
}

func TestImpactTargetDigestAndAllFailures(t *testing.T) {
	before := fieldTarget("host")
	after := fieldTarget("other")
	before.Catalog.Identity, after.Catalog.Identity = "same", "same"
	p, err := PrepareSchemas(before, after)
	if err != nil {
		t.Fatal(err)
	}
	input := corpus.Input{Selection: corpus.Selection{Mode: "manifest", Complete: true}, Entries: []corpus.Entry{{ID: "failed", Origin: corpus.Origin{Kind: "file", RelativePath: "failed.spl"}, Failure: &corpus.AcquisitionError{Code: "not_found", Phase: "open"}}}}
	r, err := p.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	if r.BeforeDigest == r.AfterDigest || r.ExecutionComplete || r.Counts.Failed != 1 || r.Counts.Compared != 0 || r.Entries[0].Failure == nil {
		t.Fatalf("target content or failure lost: %+v", r)
	}
	mp, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &before}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &after})
	if err != nil {
		t.Fatal(err)
	}
	mr, err := mp.Compare(input)
	if err != nil || mr.Counts.Failed != 1 || mr.Counts.Compared != 0 || mr.BeforeDigest == mr.AfterDigest {
		t.Fatalf("zero-success mapping: %+v, %v", mr, err)
	}
	bad := fieldTarget("host", "host")
	if p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &bad}); p != nil || !IsInputError(err) {
		t.Fatalf("invalid target escaped preflight: %+v, %v", p, err)
	}
}

func TestImpactTraversalFailureDenominator(t *testing.T) {
	p, err := PrepareSchemas(fieldTarget(), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	input := corpus.Input{Selection: corpus.Selection{Mode: "directory", Complete: false, TraversalFailures: []corpus.AcquisitionError{{Code: "traversal_failed", Phase: "traverse", Path: "sub"}}}}
	r, err := p.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	if r.ExecutionComplete || r.Counts.Selected != 0 || r.Counts.Compared != 0 || r.Counts.Denominator != 0 || r.Counts.TraversalFailed != 1 || len(r.Entries) != 0 {
		t.Fatalf("traversal denominator understated: %+v", r)
	}
}

func TestImpactProgrammaticInputErrorsAreClassified(t *testing.T) {
	input := corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true}, Entries: []corpus.Entry{{ID: "duplicate", Document: &analysis.QueryDocument{Text: "table src"}}, {ID: "duplicate", Document: &analysis.QueryDocument{Text: "table src"}}}}
	sp, err := PrepareSchemas(fieldTarget(), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	if report, err := sp.Compare(input); report != nil || !IsInputError(err) {
		t.Fatalf("schema input error classification: %+v %v", report, err)
	}
	mp, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if report, err := mp.Compare(input); report != nil || !IsInputError(err) {
		t.Fatalf("mapping input error classification: %+v %v", report, err)
	}
}

func TestImpactUnalignedReferencesStayUnmatched(t *testing.T) {
	b := &analysis.Result{References: []analysis.Reference{{ID: "same", Kind: "field", Role: "read", ScopeID: "scope-a", Location: analysis.Location{Start: analysis.Position{Offset: 3}, End: analysis.Position{Offset: 6}}}}}
	a := &analysis.Result{References: []analysis.Reference{{ID: "same", Kind: "field", Role: "read", ScopeID: "scope-b", Location: analysis.Location{Start: analysis.Position{Offset: 4}, End: analysis.Position{Offset: 7}}}}}
	alignment := alignOriginal(b, a, "revision", "revision")
	if len(alignment.Pairs) != 0 || len(alignment.Unmatched) != 2 {
		t.Fatalf("same ID with changed evidence paired: %+v", alignment)
	}
	deltas := outcomeDeltas([]outcomeEvidence{{"same", validation.ReferenceOutcome{ReferenceID: "same", Outcome: "matching"}}}, []outcomeEvidence{{"same", validation.ReferenceOutcome{ReferenceID: "same", Outcome: "missing"}}}, alignment, "original")
	if len(deltas) != 2 || deltas[0].Change != "unmatched" || deltas[1].Change != "unmatched" {
		t.Fatalf("unaligned same-ID outcomes falsely paired: %+v", deltas)
	}
	beforeDiagnostic := analysis.Diagnostic{Code: "SPL_UNKNOWN_FIELD", Category: "unknown_field", Location: b.References[0].Location, ScopeID: "scope-a"}
	afterDiagnostic := analysis.Diagnostic{Code: "SPL_UNKNOWN_FIELD", Category: "unknown_field", Location: a.References[0].Location, ScopeID: "scope-b"}
	deltas = diagnosticDeltas([]analysis.Diagnostic{beforeDiagnostic}, []analysis.Diagnostic{afterDiagnostic}, b, a, alignment, "original")
	if len(deltas) != 2 || deltas[0].Change != "unmatched" || deltas[1].Change != "unmatched" {
		t.Fatalf("unaligned same-ID diagnostics falsely paired: %+v", deltas)
	}
}

func TestImpactEmptyCollectionsAreArrays(t *testing.T) {
	p, err := PrepareSchemas(fieldTarget(), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table host"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	if string(top["entries"]) == "null" {
		t.Fatalf("null entries: %s", raw)
	}
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(top["entries"], &entries); err != nil {
		t.Fatal(err)
	}
	entry := entries[0]
	for _, key := range []string{"deltas", "reasons"} {
		if string(entry[key]) == "null" {
			t.Fatalf("null %s: %s", key, raw)
		}
	}
}

func TestAmbiguousAlignmentWithoutOutcomeDeltaIsIndeterminate(t *testing.T) {
	classification, reasons := classify(nil, true, false, Alignment{Ambiguous: []string{"before:ref-0"}})
	if classification != Indeterminate || len(reasons) == 0 {
		t.Fatalf("ambiguous evidence called unchanged: %s %v", classification, reasons)
	}
}

func TestCandidateAlignmentRequiresCandidateRevision(t *testing.T) {
	ref := analysis.Reference{ID: "ref-0", Kind: "field", Role: "read", ScopeID: "scope-0", StageID: "stage-0", Location: analysis.Location{Start: analysis.Position{Offset: 6}, End: analysis.Position{Offset: 9}}}
	before := &rewrite.Result{OriginalText: "table src", CandidateText: "table src", OriginalAnalysis: &analysis.Result{}, CandidateAnalysis: &analysis.Result{Document: analysis.QueryDocument{Text: "table src", Language: "spl"}, References: []analysis.Reference{ref}}}
	after := &rewrite.Result{OriginalText: "table src", CandidateText: "table src", OriginalAnalysis: &analysis.Result{}, CandidateAnalysis: &analysis.Result{Document: analysis.QueryDocument{Text: "table src", Language: "spl2"}, References: []analysis.Reference{ref}}}
	alignment := alignMapping(before, after, "original-revision")
	for _, pair := range alignment.Pairs {
		if pair.Domain == "candidate" {
			t.Fatalf("candidate revision mismatch paired: %+v", alignment)
		}
	}
	if len(alignment.Unmatched) == 0 {
		t.Fatalf("candidate revision mismatch hidden: %+v", alignment)
	}
}

func TestCandidateProvenanceConflictNeverRecoversByOrder(t *testing.T) {
	report := &rewrite.Result{OriginalText: "table src", CandidateText: "table first", OriginalAnalysis: &analysis.Result{References: []analysis.Reference{{ID: "original"}}}, CandidateAnalysis: &analysis.Result{References: []analysis.Reference{{ID: "first"}, {ID: "second"}}}, Changes: []rewrite.Change{
		{OriginalReferenceIDs: []string{"original"}, CandidateReferenceIDs: []string{"first"}, CandidateApplied: true},
		{OriginalReferenceIDs: []string{"original"}, CandidateReferenceIDs: []string{"second"}, CandidateApplied: true},
		{OriginalReferenceIDs: []string{"original"}, CandidateReferenceIDs: []string{"first"}, CandidateApplied: true},
	}}
	if got := provenCandidateOrigins(report); len(got) != 0 {
		t.Fatalf("conflicting provenance resolved by audit order: %+v", got)
	}
}

func TestCandidateManyToOneProvenanceIsAmbiguous(t *testing.T) {
	original := &analysis.Result{Document: analysis.QueryDocument{Text: "table a b"}, References: []analysis.Reference{{ID: "o1", Kind: "field", Role: "read", Location: analysis.Location{Start: analysis.Position{Offset: 6}, End: analysis.Position{Offset: 7}}}, {ID: "o2", Kind: "field", Role: "read", Location: analysis.Location{Start: analysis.Position{Offset: 8}, End: analysis.Position{Offset: 9}}}}}
	before := &rewrite.Result{OriginalText: "table a b", CandidateText: "table a b", OriginalAnalysis: original, CandidateAnalysis: original}
	after := &rewrite.Result{OriginalText: "table a b", CandidateText: "table c", OriginalAnalysis: original, CandidateAnalysis: &analysis.Result{Document: analysis.QueryDocument{Text: "table c"}, References: []analysis.Reference{{ID: "dest"}}}, Changes: []rewrite.Change{{GroupID: "g1", OriginalReferenceIDs: []string{"o1"}, CandidateReferenceIDs: []string{"dest"}, CandidateApplied: true}, {GroupID: "g2", OriginalReferenceIDs: []string{"o2"}, CandidateReferenceIDs: []string{"dest"}, CandidateApplied: true}}}
	alignment := alignMapping(before, after, "same-revision")
	for _, pair := range alignment.Pairs {
		if pair.Domain == "candidate" {
			t.Fatalf("many-to-one candidate correspondence paired: %+v", alignment)
		}
	}
	if len(alignment.Ambiguous) == 0 {
		t.Fatalf("many-to-one candidate correspondence not marked ambiguous: %+v", alignment)
	}
}

func TestImpactDurableWireCases(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/tooling/impact-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string]struct {
		Request             json.RawMessage `json:"request"`
		Classification      Classification  `json:"classification"`
		BeforeStatus        analysis.Status `json:"before_status"`
		AfterStatus         analysis.Status `json:"after_status"`
		BeforeCandidateText string          `json:"before_candidate_text"`
		AfterCandidateText  string          `json:"after_candidate_text"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 3 {
		t.Fatalf("fixture case count: %d", len(cases))
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var report *Report
			var err error
			if name == "schema_same_status_changed_outcome" {
				request, e := DecodeSchemaRequest(tc.Request)
				if e != nil {
					t.Fatal(e)
				}
				report, err = CompareSchemas(request)
			} else {
				request, e := DecodeMappingRequest(tc.Request)
				if e != nil {
					t.Fatal(e)
				}
				report, err = CompareMappings(request)
			}
			if err != nil {
				t.Fatal(err)
			}
			entry := report.Entries[0]
			if entry.Classification != tc.Classification || (tc.BeforeStatus != "" && entry.BeforeStatus != tc.BeforeStatus) || (tc.AfterStatus != "" && entry.AfterStatus != tc.AfterStatus) {
				t.Fatalf("wire case classification/status: %+v", entry)
			}
			if tc.BeforeCandidateText != "" && (entry.BeforeRewrite.CandidateText != tc.BeforeCandidateText || entry.AfterRewrite.CandidateText != tc.AfterCandidateText) {
				t.Fatalf("wire case candidate text: %+v", entry)
			}
		})
	}
}

func impactInput(text string) corpus.Input {
	return corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true}, Entries: []corpus.Entry{{ID: "q", Origin: corpus.Origin{Kind: "inline"}, Document: &analysis.QueryDocument{Text: text}}}}
}

func fieldTarget(fields ...string) corpus.ValidationTarget {
	return corpus.ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: fields}}
}

func fieldRule(id, source, target string) rewrite.Rule {
	return rewrite.Rule{ID: id, Kind: "field", Source: rewrite.Identity{Name: &source}, Target: rewrite.Identity{Name: &target}}
}

func TestSchemaSameStatusChangedOutcome(t *testing.T) {
	p, err := PrepareSchemas(fieldTarget("host"), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("search host=x missing=y"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Entries) != 1 || r.Entries[0].Before == nil || r.Entries[0].After == nil {
		t.Fatalf("canonical sides absent: %+v", r)
	}
	e := r.Entries[0]
	if e.Before.FieldValidation.Status != analysis.Invalid || e.After.FieldValidation.Status != analysis.Invalid || e.Classification != Affected {
		t.Fatalf("same-status changed outcome hidden: %+v", e)
	}
	var hostChanged bool
	for _, d := range e.Deltas {
		if d.Category == "outcome" && d.Change == "changed" {
			hostChanged = true
		}
	}
	if !hostChanged || r.Counts.Affected != 1 || r.Counts.Denominator != 1 {
		t.Fatalf("outcome delta/counts absent: %+v", r)
	}
}

func TestMappingIdentityBaseline(t *testing.T) {
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{}}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "user")}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.BeforeRewrite == nil || e.AfterRewrite == nil || e.BeforeRewrite.CandidateText != "table src" || e.AfterRewrite.CandidateText != "table user" || e.AfterRewrite.Text != "table src" || e.AfterRewrite.Committed || e.Classification != Affected {
		t.Fatalf("identity baseline or preview policy lost: %+v", e)
	}
}

func TestMappingIdenticalCandidateDoesNotDuplicateAlignment(t *testing.T) {
	side := MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}
	p, err := PrepareMappings(side, side)
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.Classification != Unchanged {
		t.Fatalf("equal complete mapping evidence: %+v", e)
	}
	seen := map[string]bool{}
	for _, pair := range e.Alignment.Pairs {
		if pair.Domain != "original" && pair.Domain != "candidate" {
			t.Fatalf("alignment has no domain: %+v", pair)
		}
		key := pair.Domain + "\x00" + pair.BeforeID + "\x00" + pair.AfterID
		if seen[key] {
			t.Fatalf("duplicate reference alignment: %+v", e.Alignment)
		}
		seen[key] = true
	}
}

func TestMappingPreservesCandidateValidationDeltas(t *testing.T) {
	target := fieldTarget("src")
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &target}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "user")}}, ValidationTarget: &target})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.BeforeRewrite.CandidateValidation == nil || e.AfterRewrite.CandidateValidation == nil || e.BeforeRewrite.CandidateValidation.FieldList == nil || e.AfterRewrite.CandidateValidation.FieldList == nil || e.BeforeRewrite.CandidateValidation.FieldList.Status != analysis.Valid || e.AfterRewrite.CandidateValidation.FieldList.Status != analysis.Invalid {
		t.Fatalf("canonical validation evidence missing: %+v", e)
	}
	var changedOutcome, diagnostic bool
	for _, delta := range e.Deltas {
		if delta.Category == "outcome" {
			changedOutcome = true
		}
		if delta.Category == "diagnostic" {
			diagnostic = true
		}
	}
	if !changedOutcome || !diagnostic || e.Classification != Affected {
		t.Fatalf("candidate validation delta hidden: %+v", e)
	}
	var provenance bool
	for _, pair := range e.Alignment.Pairs {
		if pair.Domain == "candidate" && pair.Basis == "audit_original_reference_provenance" {
			provenance = true
		}
		if pair.Domain == "candidate" && pair.Basis != "audit_original_reference_provenance" {
			t.Fatalf("shifted candidate aligned without audit: %+v", pair)
		}
	}
	if !provenance {
		t.Fatalf("M6 candidate provenance not carried: %+v", e.Alignment)
	}
}

func TestShiftedUnprovenCandidateReferenceRemainsUnmatched(t *testing.T) {
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "longer_name")}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src other"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.AfterRewrite.CandidateText != "table longer_name other" {
		t.Fatalf("candidate was not shifted: %q", e.AfterRewrite.CandidateText)
	}
	var unmatchedCandidate bool
	for _, value := range e.Alignment.Unmatched {
		if strings.HasPrefix(value, "candidate:") {
			unmatchedCandidate = true
		}
	}
	if !unmatchedCandidate {
		t.Fatalf("unproven shifted candidate reference hidden: %+v", e.Alignment)
	}
}

func TestMappingAlignmentAmbiguous(t *testing.T) {
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("a", "src", "user"), fieldRule("b", "src", "owner")}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.AfterRewrite == nil || e.AfterRewrite.Coverage.RewriteComplete || e.Classification != Indeterminate || len(e.Alignment.Ambiguous) == 0 {
		t.Fatalf("ambiguous preview or alignment overstated: %+v", e)
	}

	p, err = PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}})
	if err != nil {
		t.Fatal(err)
	}
	r, err = p.Compare(impactInput("| mystery | table src"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Entries[0].Classification != Indeterminate {
		t.Fatalf("equal incomplete evidence called unchanged: %+v", r.Entries[0])
	}
}

func TestImpactNeverWrites(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "q.spl")
	original := []byte("table src\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	input, err := corpusio.LoadDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "user")}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Compare(input); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("source changed: %q", got)
	}
}
