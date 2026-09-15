package main

import (
	"fmt"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func analysisStatusExitCode(status analysis.Status) int {
	switch status {
	case analysis.Valid:
		return 0
	case analysis.Invalid:
		return 1
	case analysis.Incomplete:
		return 3
	default:
		return 2
	}
}

func requirementsExitCode(set *analysis.RequirementSet) int {
	if set == nil {
		return 2
	}
	switch set.QueryStatus {
	case analysis.Invalid:
		return 1
	case analysis.Incomplete:
		return 3
	case analysis.Valid:
		if !set.Coverage.Complete {
			return 3
		}
		return 0
	default:
		return 2
	}
}

func formatAnalysisText(report *analysis.Result) []byte {
	var payload strings.Builder
	fmt.Fprintf(&payload, "Status: %s\n", report.Status)
	fmt.Fprintf(&payload, "Syntax coverage: %s\n", coverageLabel(report.Coverage.SyntaxComplete))
	fmt.Fprintf(&payload, "Semantic coverage: %s\n", coverageLabel(report.Coverage.SemanticComplete))
	if len(report.Coverage.Reasons) > 0 {
		fmt.Fprintf(&payload, "Coverage reasons: %s\n", strings.Join(report.Coverage.Reasons, ", "))
	}
	payload.WriteString("References:\n")
	if len(report.References) == 0 {
		payload.WriteString("  (none)\n")
	}
	for _, reference := range report.References {
		fmt.Fprintf(&payload, "  - %s %s/%s %q (%s, %s) @ %s\n",
			reference.ID, reference.Kind, reference.Role, reference.OriginalName,
			reference.Binding, reference.Resolution, formatAnalysisLocation(reference.Location))
	}
	payload.WriteString("Diagnostics:\n")
	if len(report.Diagnostics) == 0 {
		payload.WriteString("  (none)\n")
	}
	for _, diagnostic := range report.Diagnostics {
		fmt.Fprintf(&payload, "  - %s [%s/%s] @ %s: %s\n",
			diagnostic.Code, diagnostic.Severity, diagnostic.Category,
			formatAnalysisLocation(diagnostic.Location), diagnostic.Message)
	}
	return []byte(payload.String())
}

func formatRequirementsText(set *analysis.RequirementSet) []byte {
	var payload strings.Builder
	fmt.Fprintf(&payload, "Query status: %s\n", set.QueryStatus)
	fmt.Fprintf(&payload, "Requirement coverage: %s\n", coverageLabel(set.Coverage.Complete))
	if len(set.Coverage.Reasons) > 0 {
		fmt.Fprintf(&payload, "Coverage reasons: %s\n", strings.Join(set.Coverage.Reasons, ", "))
	}
	payload.WriteString("Items:\n")
	if len(set.Items) == 0 {
		payload.WriteString("  (none)\n")
	}
	for _, item := range set.Items {
		fmt.Fprintf(&payload, "  - %s %s/%s %q (%s, %s, %s)\n",
			item.ID, item.Kind, item.Role, item.Identity,
			item.Necessity, item.Origin, item.Resolution)
		for _, occurrence := range item.Occurrences {
			fmt.Fprintf(&payload, "    occurrence %s %q (%s) stage=%s scope=%s @ %s\n",
				occurrence.ReferenceID, occurrence.OriginalName, occurrence.Binding,
				occurrence.StageID, occurrence.ScopeID, formatAnalysisLocation(occurrence.Location))
		}
	}
	payload.WriteString("Gaps:\n")
	if len(set.Gaps) == 0 {
		payload.WriteString("  (none)\n")
	}
	for _, gap := range set.Gaps {
		fmt.Fprintf(&payload, "  - %s: %s (references: %s; diagnostics: %s)\n",
			gap.Code, gap.Message, joinedRequirementEvidence(gap.ReferenceIDs), joinedRequirementEvidence(gap.DiagnosticCodes))
	}
	payload.WriteString("Diagnostics:\n")
	if len(set.Diagnostics) == 0 {
		payload.WriteString("  (none)\n")
	}
	for _, diagnostic := range set.Diagnostics {
		fmt.Fprintf(&payload, "  - %s [%s/%s] @ %s: %s\n",
			diagnostic.Code, diagnostic.Severity, diagnostic.Category,
			formatAnalysisLocation(diagnostic.Location), diagnostic.Message)
	}
	return []byte(payload.String())
}

func joinedRequirementEvidence(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

func coverageLabel(complete bool) string {
	if complete {
		return "complete"
	}
	return "incomplete"
}

func formatAnalysisLocation(location analysis.Location) string {
	return fmt.Sprintf("%d:%d-%d:%d, bytes %d-%d",
		location.Start.Line, location.Start.Column,
		location.End.Line, location.End.Column,
		location.Start.Offset, location.End.Offset)
}

func formatCapabilitiesText(manifest analysis.CapabilityManifest) []byte {
	var payload strings.Builder
	fmt.Fprintf(&payload, "Schema version: %d\n", manifest.SchemaVersion)
	fmt.Fprintf(&payload, "Language: %s\n", manifest.Language)
	fmt.Fprintf(&payload, "Profile: %s\n", manifest.Profile)
	fmt.Fprintf(&payload, "Compatibility version: %s\n", manifest.Version)
	if manifest.DocumentationSnapshot != "" {
		fmt.Fprintf(&payload, "Documentation snapshot: %s\n", manifest.DocumentationSnapshot)
	}
	if manifest.Rewrite != nil {
		fmt.Fprintf(&payload, "Rewrite schema version: %d\n", manifest.Rewrite.SchemaVersion)
		payload.WriteString("Rewrite forms:\n")
		for _, form := range manifest.Rewrite.Forms {
			fmt.Fprintf(&payload, "  - %s/%s: supported=%t identities=%s", form.Kind, form.Role, form.Supported, strings.Join(form.IdentityForms, ","))
			writeCapabilityLimitations(&payload, form.Limitations)
		}
	}
	payload.WriteString("Commands:\n")
	for _, capability := range manifest.Commands {
		fmt.Fprintf(&payload, "  - %s: syntax=%t semantic=%t", capability.Name, capability.SyntaxSupported, capability.SemanticSupported)
		writeCapabilityLimitations(&payload, capability.Limitations)
	}
	payload.WriteString("Functions:\n")
	for _, capability := range manifest.Functions {
		fmt.Fprintf(&payload, "  - %s: syntax=%t semantic=%t", capability.Name, capability.SyntaxSupported, capability.SemanticSupported)
		writeCapabilityLimitations(&payload, capability.Limitations)
	}
	return []byte(payload.String())
}

func writeCapabilityLimitations(payload *strings.Builder, limitations []string) {
	if len(limitations) > 0 {
		fmt.Fprintf(payload, " (%s)", strings.Join(limitations, "; "))
	}
	payload.WriteByte('\n')
}
