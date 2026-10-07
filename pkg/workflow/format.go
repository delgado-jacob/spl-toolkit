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

// FormatComparison retains both assessment states, meaningful changes and
// uncertainty. Static evidence equality does not assert runtime equivalence.
func FormatComparison(report *ComparisonReport) string {
	if report == nil {
		return ""
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return ""
	}
	text := fmt.Sprintf("Execution complete: %t\nSelected: %d; compared: %d; affected: %d; unchanged: %d; indeterminate: %d; failed: %d\nCI exit code: %d\n", report.ExecutionComplete, report.Counts.Selected, report.Counts.Compared, report.Counts.Affected, report.Counts.Unchanged, report.Counts.Indeterminate, report.Counts.Failed, report.CIExitCode)
	for _, entry := range report.Entries {
		text += fmt.Sprintf("%s: %s; before: %s; after: %s; meaningful changes: %d; uncertainty: %v\n", entry.ID, entry.Classification, entry.Before.Status, entry.After.Status, len(entry.Deltas), entry.Reasons)
	}
	return text + "\nOffline evidence comparison:\n" + string(raw) + "\n"
}

// FormatRecheck retains the fresh canonical assessment and exact source links.
func FormatRecheck(report *RecheckReport) string {
	if report == nil {
		return ""
	}
	return fmt.Sprintf("Detection: %s\nOriginal source hash: %s\nProposed source hash: %s\nSource links do not prove preserved detection intent.\n\n%s", report.DetectionID, report.OriginalSourceHash, report.ProposedSourceHash, FormatReport(&report.Assessment))
}
