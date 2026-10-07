package resolution

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"strings"
	"testing"
)

func TestFormatReportPublicationAndEvidence(t *testing.T) {
	verified, failed, incomplete := "from verified", "from diagnostic_failed", "from diagnostic_incomplete"
	report := &Report{GeneratedCount: 3, TotalCombinations: "3", MaxVariants: 100, Counts: Counts{Verified: 1, Failed: 1, Incomplete: 1}, Variants: []Variant{
		{Ordinal: 1, Outcome: "verified", Selection: []analysis.ResolutionChoice{{Placeholder: "$events", Kind: "dataset", Value: "verified"}}, Changes: []analysis.ResolutionChange{{Before: "$events", After: "verified", OriginalLocation: analysis.Location{Start: analysis.Position{Offset: 5}, End: analysis.Position{Offset: 12}}}}, ResolvedQuery: &verified, Proof: analysis.ResolutionProofEvidence{Proven: true}, Compatibility: &compatibility.ResolutionReport{Outcome: "satisfied"}},
		{Ordinal: 2, Outcome: "failed", CandidateText: &failed, ResolvedQuery: &failed, Compatibility: &compatibility.ResolutionReport{Outcome: "unsatisfied", Reasons: []compatibility.ResolutionReason{{Evidence: compatibility.Reason{Code: "schema_field_missing", Message: "id is absent"}}}}},
		{Ordinal: 3, Outcome: "incomplete", CandidateText: &incomplete, Proof: analysis.ResolutionProofEvidence{Limitations: []analysis.ResolutionLimitation{{Code: "unproved_owner", Message: "owner cannot be proved"}}}},
	}}
	text := FormatReport(report)
	for _, want := range []string{"3 generated of 3", "Verified: 1; failed: 1; incomplete: 1", "Variant 1: verified", "Variant 2: failed", "Variant 3: incomplete", `$events (dataset) = "verified"`, `Change [5,12): "$events" -> "verified"`, "Compatibility finding schema_field_missing", "Proof limitation unproved_owner", "Verified query:\nfrom verified"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %s", want, text)
		}
	}
	if strings.Count(text, "Verified query:") != 1 || strings.Contains(text, failed) || strings.Contains(text, incomplete) {
		t.Fatal("diagnostic query published", text)
	}
	if FormatReport(nil) != "" {
		t.Fatal("nil report should be empty")
	}
}
