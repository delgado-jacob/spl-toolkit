package compatibility

import (
	"encoding/json"
	"fmt"
)

// FormatReport formats the detached public report for display.
func FormatReport(report *Report) string {
	if report == nil {
		return ""
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("Outcome: %s\nCorrelation: %s\n\nEvidence:\n%s\n", report.Outcome, report.Correlation.Outcome, raw)
}
