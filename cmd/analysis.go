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
