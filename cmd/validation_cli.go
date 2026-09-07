package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func runValidationCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	options, format, err := parseCLIOptions("validate-fields", args)
	if err != nil {
		return writeCLIError(stderr, format, err.Error(), 2)
	}
	if options.help {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}
	if err := validateFieldCLIOptions(options); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	payload, status, err := computeValidationCLIResult(options, stdin)
	if err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	if err := writeCLIResult(payload, options.output, stdout); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	return analysisStatusExitCode(status)
}

func validateFieldCLIOptions(options cliOptions) error {
	if !options.hasFields || options.fields == "-" {
		return fmt.Errorf("validate-fields requires --fields with a local file path (not '-')")
	}
	if options.hasConfig {
		return fmt.Errorf("validate-fields does not accept --config")
	}
	sources := 0
	for _, supplied := range []bool{options.hasQuery, options.hasFile, options.hasStdin, options.hasBatch} {
		if supplied {
			sources++
		}
	}
	if sources != 1 {
		return fmt.Errorf("validate-fields requires exactly one query, --file, --stdin, or --batch")
	}
	if options.hasBatch && (options.hasLanguage || options.hasProfile || options.hasCompatibilityVersion || options.hasSourceID) {
		return fmt.Errorf("batch documents must supply their own language, profile, version, and source_id; global document options are not allowed")
	}
	return nil
}

func computeValidationCLIResult(options cliOptions, stdin io.Reader) ([]byte, analysis.Status, error) {
	data, err := os.ReadFile(options.fields)
	if err != nil {
		return nil, "", fmt.Errorf("read fields: %w", err)
	}
	catalog, err := validation.DecodeFieldCatalog(data)
	if err != nil {
		return nil, "", err
	}
	if options.hasBatch {
		if options.batch == "-" {
			data, err = io.ReadAll(stdin)
		} else {
			data, err = os.ReadFile(options.batch)
		}
		if err != nil {
			return nil, "", fmt.Errorf("read batch: %w", err)
		}
		documents, err := validation.DecodeDocuments(data)
		if err != nil {
			return nil, "", err
		}
		report, err := validation.ValidateBatch(documents, catalog)
		if err != nil {
			return nil, "", err
		}
		if options.format == "json" {
			payload, _, err := marshalCLILine(report)
			return payload, report.Status, err
		}
		var text strings.Builder
		fmt.Fprintf(&text, "Batch status: %s\n", report.Status)
		for i, item := range report.Reports {
			fmt.Fprintf(&text, "\nDocument %d:\n%s", i+1, formatValidationText(item))
		}
		return []byte(text.String()), report.Status, nil
	}
	document := analysis.QueryDocument{Text: options.query, Language: options.language, Profile: options.profile, Version: options.compatibilityVersion, SourceID: options.sourceID}
	if options.hasFile || options.hasStdin {
		if options.hasFile {
			data, err = os.ReadFile(options.file)
			document.SourceID = options.file
		} else {
			data, err = io.ReadAll(stdin)
			document.SourceID = "<stdin>"
		}
		if err != nil {
			return nil, "", fmt.Errorf("read query: %w", err)
		}
		document.Text = string(data)
		if options.hasSourceID {
			document.SourceID = options.sourceID
		}
	}
	report, err := validation.Validate(document, catalog)
	if err != nil {
		return nil, "", err
	}
	if options.format == "json" {
		payload, _, err := marshalCLILine(report)
		return payload, report.Status, err
	}
	return formatValidationText(report), report.Status, nil
}

func formatValidationText(report *validation.Report) []byte {
	var text strings.Builder
	fmt.Fprintf(&text, "Status: %s\nSource: %q\n", report.Status, report.Analysis.Document.SourceID)
	fmt.Fprintf(&text, "Syntax coverage: %s\nSemantic coverage: %s\nSchema coverage: %s\n", coverageLabel(report.Coverage.SyntaxComplete), coverageLabel(report.Coverage.SemanticComplete), coverageLabel(report.Coverage.SchemaComplete))
	if len(report.Coverage.Reasons) > 0 {
		fmt.Fprintf(&text, "Coverage reasons: %s\n", strings.Join(report.Coverage.Reasons, ", "))
	}
	references := make(map[string]analysis.Reference, len(report.Analysis.References))
	for _, reference := range report.Analysis.References {
		references[reference.ID] = reference
	}
	text.WriteString("Outcomes:\n")
	if len(report.Outcomes) == 0 {
		text.WriteString("  (none)\n")
	}
	for _, outcome := range report.Outcomes {
		reference := references[outcome.ReferenceID]
		fmt.Fprintf(&text, "  - %s %q: %s @ %s\n", outcome.ReferenceID, reference.OriginalName, outcome.Outcome, formatAnalysisLocation(reference.Location))
		for _, match := range outcome.Matches {
			fmt.Fprintf(&text, "    %q (%s): %s\n", match.Name, match.Binding, match.Outcome)
		}
	}
	text.WriteString("Diagnostics:\n")
	if len(report.Diagnostics) == 0 {
		text.WriteString("  (none)\n")
	}
	for _, diagnostic := range report.Diagnostics {
		fmt.Fprintf(&text, "  - %s [%s/%s] @ %s: %s\n", diagnostic.Code, diagnostic.Severity, diagnostic.Category, formatAnalysisLocation(diagnostic.Location), diagnostic.Message)
	}
	return []byte(text.String())
}
