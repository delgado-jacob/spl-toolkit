package environment

import (
	"fmt"
	"strings"
)

func copyEnvironmentReport(value Report) Report {
	value.Coverage = append([]CoverageEntry{}, value.Coverage...)
	value.Diagnostics = append([]Diagnostic{}, value.Diagnostics...)
	return value
}

func mergeEnvironmentReport(into, part *Report) {
	if part == nil {
		return
	}
	if part.Status == "invalid" || (part.Status == "partial" && into.Status == "valid") {
		into.Status = part.Status
	}
	if part.SnapshotDigest != "" {
		into.SnapshotDigest = part.SnapshotDigest
	}
	if part.SchemaBundleDigest != "" {
		into.SchemaBundleDigest = part.SchemaBundleDigest
	}
	into.Coverage = append(into.Coverage, part.Coverage...)
	into.Diagnostics = append(into.Diagnostics, part.Diagnostics...)
}

// FormatReport renders the existing report without changing validation decisions.
func FormatReport(report *Report) string {
	if report == nil {
		return ""
	}
	var out strings.Builder
	fmt.Fprintf(&out, "Status: %s\n", report.Status)
	if report.SnapshotDigest != "" {
		fmt.Fprintf(&out, "Snapshot digest: %s\n", report.SnapshotDigest)
	}
	if report.SchemaBundleDigest != "" {
		fmt.Fprintf(&out, "Schema bundle digest: %s\n", report.SchemaBundleDigest)
	}
	for _, entry := range report.Coverage {
		fmt.Fprintf(&out, "Coverage: %s %s", entry.Artifact, entry.Kind)
		if entry.SchemaID != "" || entry.ObjectID != "" {
			fmt.Fprintf(&out, " schema=%q object=%q", entry.SchemaID, entry.ObjectID)
		}
		fmt.Fprintf(&out, " %s", entry.Coverage)
		if entry.Reason != "" {
			fmt.Fprintf(&out, " (%s)", entry.Reason)
		}
		out.WriteByte('\n')
	}
	for _, diagnostic := range report.Diagnostics {
		fmt.Fprintf(&out, "%s: %s %s", diagnostic.Severity, diagnostic.Artifact, diagnostic.Code)
		if diagnostic.Path != "" {
			fmt.Fprintf(&out, " at %s", diagnostic.Path)
		}
		fmt.Fprintf(&out, ": %s\n", diagnostic.Message)
	}
	return out.String()
}
