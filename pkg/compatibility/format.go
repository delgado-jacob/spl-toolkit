package compatibility

import "encoding/json"

// FormatReport formats the detached public report for display.
func FormatReport(report *Report) string {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return ""
	}
	return string(raw)
}
