package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func runRewriteCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	options, format, err := parseCLIOptions("rewrite", args)
	if err != nil {
		return writeCLIError(stderr, format, err.Error(), 2)
	}
	if options.help {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}
	if err := validateRewriteCLIOptions(options); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	payload, status, err := computeRewriteCLIResult(options, stdin)
	if err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	if err := writeCLIResult(payload, options.output, stdout); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	return analysisStatusExitCode(status)
}

func validateRewriteCLIOptions(options cliOptions) error {
	if !options.hasRules || options.rules == "-" {
		return fmt.Errorf("rewrite requires --rules with a local file path (not '-')")
	}
	if options.hasConfig {
		return fmt.Errorf("rewrite does not accept --config")
	}
	sources := 0
	for _, set := range []bool{options.hasQuery, options.hasFile, options.hasStdin, options.hasBatch} {
		if set {
			sources++
		}
	}
	if sources != 1 {
		return fmt.Errorf("rewrite requires exactly one query, --file, --stdin, or --batch")
	}
	if options.hasBatch && (options.hasLanguage || options.hasProfile || options.hasCompatibilityVersion || options.hasSourceID) {
		return fmt.Errorf("batch documents must supply their own language, profile, version, and source_id; global document options are not allowed")
	}
	flags := options.schemaOptions
	hasSchema := flags["schema"] != ""
	hasOCSF := flags["ocsf-catalog"] != ""
	families := 0
	for _, set := range []bool{options.hasFields, hasSchema, hasOCSF} {
		if set {
			families++
		}
	}
	if families > 1 {
		return fmt.Errorf("rewrite accepts at most one --fields, --schema, or --ocsf-catalog target")
	}
	if options.hasFields && options.fields == "-" {
		return fmt.Errorf("--fields requires a local path (not '-')")
	}
	for _, key := range []string{"schema", "schema-resources", "ocsf-catalog"} {
		if flags[key] == "-" {
			return fmt.Errorf("--%s requires a local path (not '-')", key)
		}
	}
	for key := range flags {
		if strings.HasPrefix(key, "schema-") && !hasSchema {
			return fmt.Errorf("--%s requires --schema", key)
		}
		if strings.HasPrefix(key, "ocsf-") && !hasOCSF {
			return fmt.Errorf("--%s requires --ocsf-catalog", key)
		}
	}
	if len(options.ocsfProfiles) > 0 || len(options.ocsfExtensions) > 0 {
		if !hasOCSF {
			return fmt.Errorf("OCSF options require --ocsf-catalog")
		}
	}
	return nil
}

func readRewriteCLITarget(options cliOptions) (*rewrite.ValidationTarget, error) {
	if options.hasFields {
		data, err := os.ReadFile(options.fields)
		if err != nil {
			return nil, fmt.Errorf("read fields: %w", err)
		}
		catalog, err := validation.DecodeFieldCatalog(data)
		if err != nil {
			return nil, err
		}
		return &rewrite.ValidationTarget{Kind: "field_list", Catalog: &catalog}, nil
	}
	if options.schemaOptions["schema"] != "" || options.schemaOptions["ocsf-catalog"] != "" {
		target, err := readSchemaCLITarget(options)
		if err != nil {
			return nil, err
		}
		return &rewrite.ValidationTarget{Kind: target.Kind, SchemaTarget: &target}, nil
	}
	return nil, nil
}

func computeRewriteCLIResult(options cliOptions, stdin io.Reader) ([]byte, analysis.Status, error) {
	rawRules, err := os.ReadFile(options.rules)
	if err != nil {
		return nil, "", fmt.Errorf("read rules: %w", err)
	}
	rules, err := rewrite.DecodeRuleSet(rawRules)
	if err != nil {
		return nil, "", err
	}
	target, err := readRewriteCLITarget(options)
	if err != nil {
		return nil, "", err
	}
	mode := rewrite.Preview
	if options.apply {
		mode = rewrite.Apply
	}
	if options.hasBatch {
		var raw []byte
		if options.batch == "-" {
			raw, err = io.ReadAll(stdin)
		} else {
			raw, err = os.ReadFile(options.batch)
		}
		if err != nil {
			return nil, "", fmt.Errorf("read batch: %w", err)
		}
		documents, err := validation.DecodeDocuments(raw)
		if err != nil {
			return nil, "", err
		}
		report, err := rewrite.RewriteBatch(rewrite.BatchRequest{SchemaVersion: 1, Mode: mode, Documents: documents, Rules: rules.Rules, ValidationTarget: target})
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
			fmt.Fprintf(&text, "\nDocument %d:\n%s", i+1, formatRewriteText(item))
		}
		return []byte(text.String()), report.Status, nil
	}
	document := analysis.QueryDocument{Text: options.query, Language: options.language, Profile: options.profile, Version: options.compatibilityVersion, SourceID: options.sourceID}
	if options.hasFile || options.hasStdin {
		var raw []byte
		if options.hasFile {
			raw, err = os.ReadFile(options.file)
			document.SourceID = options.file
		} else {
			raw, err = io.ReadAll(stdin)
			document.SourceID = "<stdin>"
		}
		if err != nil {
			return nil, "", fmt.Errorf("read query: %w", err)
		}
		document.Text = string(raw)
		if options.hasSourceID {
			document.SourceID = options.sourceID
		}
	}
	report, err := rewrite.Rewrite(rewrite.Request{SchemaVersion: 1, Mode: mode, Document: document, Rules: rules.Rules, ValidationTarget: target})
	if err != nil {
		return nil, "", err
	}
	if options.format == "json" {
		payload, _, err := marshalCLILine(report)
		return payload, report.Status, err
	}
	return formatRewriteText(report), report.Status, nil
}

func formatRewriteText(report *rewrite.Result) []byte {
	var text strings.Builder
	fmt.Fprintf(&text, "Status: %s\nMode: %s\nSource: %q\nCommitted: %t\n", report.Status, report.Mode, report.Document.SourceID, report.Committed)
	fmt.Fprintf(&text, "Original text: %s\nCandidate text: %s\nReturned text: %s\n", report.OriginalText, report.CandidateText, report.Text)
	fmt.Fprintf(&text, "Syntax coverage: %s\nSemantic coverage: %s\nRewrite coverage: %s\n", coverageLabel(report.Coverage.SyntaxComplete), coverageLabel(report.Coverage.SemanticComplete), coverageLabel(report.Coverage.RewriteComplete))
	if len(report.Coverage.Reasons) > 0 {
		fmt.Fprintf(&text, "Coverage reasons: %s\n", strings.Join(report.Coverage.Reasons, ", "))
	}
	text.WriteString("Changes:\n")
	if len(report.Changes) == 0 {
		text.WriteString("  (none)\n")
	}
	for _, change := range report.Changes {
		fmt.Fprintf(&text, "  - %s: %q -> %q (candidate_applied=%t committed=%t; %s)\n", change.Outcome, change.OldText, change.NewText, change.CandidateApplied, change.Committed, change.Reason)
	}
	text.WriteString("Rule evaluations:\n")
	if len(report.RuleEvaluations) == 0 {
		text.WriteString("  (none)\n")
	}
	for _, evaluation := range report.RuleEvaluations {
		fmt.Fprintf(&text, "  - %s: %s (%s)\n", evaluation.RuleID, evaluation.Outcome, evaluation.Reason)
	}
	return []byte(text.String())
}
