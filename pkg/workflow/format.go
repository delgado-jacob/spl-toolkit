package workflow

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FormatReport includes nested owner findings without asserting runtime behavior.
func FormatReport(report *Report) string {
	if report == nil {
		return ""
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return ""
	}
	revision, modified := report.Provenance.VCSRevision, "unavailable"
	if revision == "" {
		revision = "unavailable"
	}
	if report.Provenance.VCSModified != nil {
		modified = strconv.FormatBool(*report.Provenance.VCSModified)
	}
	return fmt.Sprintf("Status: %s\nSelection: %s (complete: %t)\nExecution complete: %t\nSelected: %d; assessed: %d; valid: %d; invalid: %d; incomplete: %d\nCI exit code: %d\nVCS revision: %s; modified: %s\n\nOffline evidence:\n%s\n", report.Status, report.Selection.Mode, report.Selection.Complete, report.ExecutionComplete, report.Counts.Selected, report.Counts.Assessed, report.Counts.Valid, report.Counts.Invalid, report.Counts.Incomplete, report.CIExitCode, revision, modified, raw)
}
