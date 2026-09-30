package closure

import "strings"

// FormatInventory renders the report's existing BOM and coverage without resolving again.
func FormatInventory(report *Report) string {
	if report == nil {
		return ""
	}
	var out strings.Builder
	out.WriteString("Status: ")
	out.WriteString(string(report.Status))
	out.WriteByte('\n')
	kind := ""
	for _, entry := range report.BOM {
		if entry.Kind != kind {
			kind = entry.Kind
			out.WriteString(kind)
			out.WriteString(":\n")
		}
		if entry.Direct {
			writeInventoryEntry(&out, "direct", entry)
		}
		if entry.Transitive {
			writeInventoryEntry(&out, "transitive", entry)
		}
	}
	if !report.Coverage.Complete {
		out.WriteString("Partial reasons:")
		for _, reason := range report.Coverage.Reasons {
			out.WriteString(" ")
			out.WriteString(reason)
		}
		out.WriteByte('\n')
	}
	return out.String()
}

func writeInventoryEntry(out *strings.Builder, scope string, entry BOMEntry) {
	out.WriteString("  ")
	out.WriteString(scope)
	out.WriteString(": ")
	out.WriteString(entry.Name)
	out.WriteString(" (")
	out.WriteString(entry.ObjectID)
	out.WriteByte(')')
	if entry.Incomplete {
		out.WriteString(" [incomplete]")
	}
	out.WriteByte('\n')
}
