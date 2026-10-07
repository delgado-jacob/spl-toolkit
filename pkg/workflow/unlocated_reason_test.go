package workflow

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
)

func TestSavedIncompleteReportAdmitsCanonicalUnlocatedInputReason(t *testing.T) {
	report, err := Assess(seedRequest(t, "from [{id:1}] | unknowable_command"))
	if err != nil {
		t.Fatal(err)
	}
	if report.CIExitCode != 3 || len(report.Entries[0].Analysis.InputCoverage.Reasons) == 0 {
		t.Fatalf("incomplete setup %+v", report)
	}
	if report.Entries[0].Analysis.InputCoverage.Reasons[0].Location != (analysis.Location{}) {
		t.Fatal("expected canonical unlocated reason")
	}
	raw, err := json.Marshal(CompareRequest{SchemaVersion: 1, Before: *report, After: *report})
	if err != nil {
		t.Fatal(err)
	}
	comparison, err := CompareJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if comparison.Entries[0].Classification != impact.Indeterminate || comparison.CIExitCode != 3 {
		t.Fatalf("comparison %+v", comparison)
	}
	raw, err = json.Marshal(EvidenceRequest{SchemaVersion: 1, Report: report, Include: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := EvidenceJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.SourceCIExitCode != 3 {
		t.Fatalf("projection %+v", evidence)
	}
}
func TestSavedInputReasonRejectsPartiallyPopulatedLocation(t *testing.T) {
	report, err := Assess(seedRequest(t, "from [{id:1}] | unknowable_command"))
	if err != nil {
		t.Fatal(err)
	}
	report.Entries[0].Analysis.InputCoverage.Reasons[0].Location.Start.Line = 1
	_, err = Compare(CompareRequest{SchemaVersion: 1, Before: *report, After: *report})
	detail, ok := RequestErrorDetails(err)
	if !ok || !strings.Contains(detail.Path, "/input_coverage/reasons/0/location") {
		t.Fatalf("detail=%+v err=%v", detail, err)
	}
	_, err = Evidence(EvidenceRequest{SchemaVersion: 1, Report: report, Include: []string{}})
	if err == nil {
		t.Fatal("evidence admitted malformed location")
	}
}
