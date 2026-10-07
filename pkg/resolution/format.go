package resolution

import (
	"fmt"
	"strings"
)

// FormatReport keeps diagnostic candidates separate from published query text.
func FormatReport(report *Report) string {
	if report == nil {
		return ""
	}
	var out strings.Builder
	fmt.Fprintf(&out, "Resolution variants: %d generated of %s (limit %d)\nVerified: %d; failed: %d; incomplete: %d\n", report.GeneratedCount, report.TotalCombinations, report.MaxVariants, report.Counts.Verified, report.Counts.Failed, report.Counts.Incomplete)
	for _, variant := range report.Variants {
		fmt.Fprintf(&out, "\nVariant %d: %s (%s)\n", variant.Ordinal, variant.Outcome, variant.ID)
		if len(variant.Selection) == 0 {
			out.WriteString("Selection: empty\n")
		}
		for _, choice := range variant.Selection {
			fmt.Fprintf(&out, "Selection: %s (%s) = %q\n", choice.Placeholder, choice.Kind, choice.Value)
		}
		for _, change := range variant.Changes {
			fmt.Fprintf(&out, "Change [%d,%d): %q -> %q\n", change.OriginalLocation.Start.Offset, change.OriginalLocation.End.Offset, change.Before, change.After)
		}
		for _, diagnostic := range variant.Diagnostics {
			fmt.Fprintf(&out, "Diagnostic %s: %s\n", diagnostic.Code, diagnostic.Message)
		}
		fmt.Fprintf(&out, "Substitution proof: %t\n", variant.Proof.Proven)
		for _, limitation := range variant.Proof.Limitations {
			fmt.Fprintf(&out, "Proof limitation %s [%d,%d): %s\n", limitation.Code, limitation.Location.Start.Offset, limitation.Location.End.Offset, limitation.Message)
		}
		if variant.Compatibility != nil {
			fmt.Fprintf(&out, "Compatibility: %s\n", variant.Compatibility.Outcome)
			for _, reason := range variant.Compatibility.Reasons {
				fmt.Fprintf(&out, "Compatibility finding %s: %s\n", reason.Evidence.Code, reason.Evidence.Message)
			}
		} else {
			out.WriteString("Compatibility: unavailable\n")
		}
		if variant.Outcome == "verified" && variant.ResolvedQuery != nil {
			fmt.Fprintf(&out, "Verified query:\n%s\n", *variant.ResolvedQuery)
		}
	}
	return out.String()
}
