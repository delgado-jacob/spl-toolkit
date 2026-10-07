package workflow

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"strings"
	"testing"
)

func TestCIExitPrecedenceAndCounts(t *testing.T) {
	for _, tc := range []struct {
		name     string
		complete bool
		statuses []analysis.Status
		failed   bool
		code     int
		status   analysis.Status
	}{
		{"all valid", true, []analysis.Status{analysis.Valid}, false, 0, analysis.Valid},
		{"invalid beats incomplete", true, []analysis.Status{analysis.Valid, analysis.Invalid, analysis.Incomplete}, false, 1, analysis.Invalid},
		{"incomplete content", true, []analysis.Status{analysis.Incomplete}, false, 3, analysis.Incomplete},
		{"selection incomplete beats invalid", false, []analysis.Status{analysis.Invalid}, false, 2, analysis.Invalid},
		{"operation failed beats invalid", true, []analysis.Status{analysis.Invalid, analysis.Incomplete}, true, 2, analysis.Invalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &Report{Status: analysis.Valid, ExecutionComplete: tc.complete}
			r.Counts.Selected = len(tc.statuses)
			for _, s := range tc.statuses {
				r.Entries = append(r.Entries, ReportEntry{Status: s})
			}
			if tc.failed {
				r.Entries[len(r.Entries)-1].Failure = &Failure{Phase: "internal"}
			}
			finalize(r)
			if r.CIExitCode != tc.code || r.Status != tc.status || r.Counts.Valid+r.Counts.Invalid+r.Counts.Incomplete != r.Counts.Selected {
				t.Fatalf("report: %+v", r)
			}
			if tc.failed && r.Counts.InternalFailed != 1 {
				t.Fatal("internal failure not counted")
			}
		})
	}
}

func TestCICompatibilityStatusDischargesOnlyEvidenceGaps(t *testing.T) {
	for _, tc := range []struct {
		original analysis.Status
		outcome  string
		want     analysis.Status
	}{
		{analysis.Incomplete, "satisfied", analysis.Valid},
		{analysis.Invalid, "satisfied", analysis.Invalid},
		{analysis.Valid, "unsatisfied", analysis.Invalid},
		{analysis.Incomplete, "not assessed", analysis.Incomplete},
		{analysis.Valid, "incomplete", analysis.Incomplete},
	} {
		if got := compatibilityStatus(&analysis.Result{Status: tc.original}, &compatibility.Report{Outcome: tc.outcome}); got != tc.want {
			t.Errorf("original=%s outcome=%s got=%s want=%s", tc.original, tc.outcome, got, tc.want)
		}
	}
}

func TestAssessTextAndProvenance(t *testing.T) {
	req := seedRequest(t, "from [{id:1}] | eval broken =")
	first, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if first.Provenance.SettingsDigest == "" || first.Provenance.EnvironmentDigest == "" || first.Provenance.ToolkitVersion == "" || first.Provenance.SettingsDigest != second.Provenance.SettingsDigest {
		t.Fatalf("provenance: %+v", first.Provenance)
	}
	text := FormatReport(first)
	for _, s := range []string{"Selection: inline (complete: true)", "invalid: 1", "CI exit code: 1", "SPL_SYNTAX_ERROR", "compatibility"} {
		if !strings.Contains(text, s) {
			t.Errorf("text missing %q", s)
		}
	}
	if FormatReport(nil) != "" {
		t.Fatal("nil text")
	}
	req.Settings.Snapshot.Objects[0].Name = "changed"
	changed, err := Prepare(req.Settings)
	if err == nil && changed.provenance.SettingsDigest == first.Provenance.SettingsDigest {
		t.Fatal("settings digest ignored mutation")
	}
}
