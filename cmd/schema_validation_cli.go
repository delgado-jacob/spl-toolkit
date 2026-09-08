package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func isSchemaCLIOption(name string) bool {
	switch name {
	case "schema", "schema-base-uri", "schema-resources", "ocsf-catalog", "ocsf-version", "ocsf-class", "ocsf-category", "ocsf-profile", "ocsf-extension":
		return true
	}
	return false
}

func setSchemaCLIOption(options *cliOptions, name, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("--%s must be valid UTF-8", name)
	}
	switch name {
	case "ocsf-profile":
		options.ocsfProfiles = append(options.ocsfProfiles, value)
	case "ocsf-extension":
		options.ocsfExtensions = append(options.ocsfExtensions, value)
	default:
		if options.schemaOptions == nil {
			options.schemaOptions = make(map[string]string)
		}
		if _, ok := options.schemaOptions[name]; ok {
			return fmt.Errorf("duplicate option --%s", name)
		}
		options.schemaOptions[name] = value
	}
	return nil
}

func runSchemaValidationCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	options, format, err := parseCLIOptions("validate-schema", args)
	if err != nil {
		return writeCLIError(stderr, format, err.Error(), 2)
	}
	if options.help {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}
	if err = validateSchemaCLIOptions(options); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	payload, status, err := computeSchemaCLIResult(options, stdin)
	if err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	if err = writeCLIResult(payload, options.output, stdout); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	return analysisStatusExitCode(status)
}

func validateSchemaCLIOptions(options cliOptions) error {
	flags := options.schemaOptions
	if (flags["schema"] != "") == (flags["ocsf-catalog"] != "") {
		return fmt.Errorf("validate-schema requires exactly one --schema or --ocsf-catalog")
	}
	for _, key := range []string{"schema", "schema-resources", "ocsf-catalog"} {
		if flags[key] == "-" {
			return fmt.Errorf("--%s requires a local path (not '-')", key)
		}
	}
	for key := range flags {
		if flags["schema"] != "" && strings.HasPrefix(key, "ocsf-") {
			return fmt.Errorf("--%s requires --ocsf-catalog", key)
		}
		if flags["ocsf-catalog"] != "" && strings.HasPrefix(key, "schema-") {
			return fmt.Errorf("--%s requires --schema", key)
		}
	}
	if flags["schema"] != "" && (len(options.ocsfProfiles) > 0 || len(options.ocsfExtensions) > 0) {
		return fmt.Errorf("OCSF options require --ocsf-catalog")
	}
	if options.hasConfig {
		return fmt.Errorf("validate-schema does not accept --config")
	}
	sources := 0
	for _, set := range []bool{options.hasQuery, options.hasFile, options.hasStdin, options.hasBatch} {
		if set {
			sources++
		}
	}
	if sources != 1 {
		return fmt.Errorf("validate-schema requires exactly one query, --file, --stdin, or --batch")
	}
	if options.hasBatch && (options.hasLanguage || options.hasProfile || options.hasCompatibilityVersion || options.hasSourceID) {
		return fmt.Errorf("batch documents must supply their own language, profile, version, and source_id; global document options are not allowed")
	}
	return nil
}

func readSchemaCLITarget(options cliOptions) (validation.SchemaTarget, error) {
	flags := options.schemaOptions
	target := map[string]any{}
	read := func(key string) (json.RawMessage, error) {
		data, err := os.ReadFile(flags[key])
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", key, err)
		}
		return json.RawMessage(data), nil
	}
	if flags["schema"] != "" {
		target["kind"] = "json_schema"
		raw, err := read("schema")
		if err != nil {
			return validation.SchemaTarget{}, err
		}
		target["schema"] = raw
		if base, ok := flags["schema-base-uri"]; ok {
			target["base_uri"] = base
		}
		if _, ok := flags["schema-resources"]; ok {
			raw, err := read("schema-resources")
			if err != nil {
				return validation.SchemaTarget{}, err
			}
			target["resources"] = raw
		}
	} else {
		target["kind"] = "ocsf"
		raw, err := read("ocsf-catalog")
		if err != nil {
			return validation.SchemaTarget{}, err
		}
		target["catalog"] = raw
		profiles := append([]string{}, options.ocsfProfiles...)
		extensions := append([]string{}, options.ocsfExtensions...)
		selection := map[string]any{"version": flags["ocsf-version"], "profiles": profiles, "extensions": extensions}
		for _, key := range []string{"class", "category"} {
			if value, ok := flags["ocsf-"+key]; ok {
				numeric := value != ""
				for _, r := range value {
					if r < '0' || r > '9' {
						numeric = false
					}
				}
				if numeric {
					uid, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return validation.SchemaTarget{}, fmt.Errorf("--ocsf-%s UID must fit int64", key)
					}
					selection[key+"_uid"] = uid
				} else {
					selection[key] = value
				}
			}
		}
		target["selection"] = selection
	}
	raw, err := json.Marshal(target)
	if err != nil {
		return validation.SchemaTarget{}, err
	}
	return validation.DecodeSchemaTarget(raw)
}

func computeSchemaCLIResult(options cliOptions, stdin io.Reader) ([]byte, analysis.Status, error) {
	target, err := readSchemaCLITarget(options)
	if err != nil {
		return nil, "", err
	}
	if options.hasBatch {
		var data []byte
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
		report, err := validation.ValidateSchemaBatch(documents, target)
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
			fmt.Fprintf(&text, "\nDocument %d:\n%s", i+1, formatSchemaValidationText(item))
		}
		return []byte(text.String()), report.Status, nil
	}
	document := analysis.QueryDocument{Text: options.query, Language: options.language, Profile: options.profile, Version: options.compatibilityVersion, SourceID: options.sourceID}
	if options.hasFile || options.hasStdin {
		var data []byte
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
	report, err := validation.ValidateSchema(document, target)
	if err != nil {
		return nil, "", err
	}
	if options.format == "json" {
		payload, _, err := marshalCLILine(report)
		return payload, report.Status, err
	}
	return formatSchemaValidationText(report), report.Status, nil
}

func formatSchemaValidationText(report *validation.SchemaReport) []byte {
	var text strings.Builder
	target, _, _ := marshalCLILine(report.Target)
	fmt.Fprintf(&text, "Target: %s", target)
	fmt.Fprintf(&text, "Status: %s\nSource: %q\nSyntax coverage: %s\nSemantic coverage: %s\nSchema coverage: %s\n", report.Status, report.Analysis.Document.SourceID, coverageLabel(report.Coverage.SyntaxComplete), coverageLabel(report.Coverage.SemanticComplete), coverageLabel(report.Coverage.SchemaComplete))
	if len(report.Coverage.Reasons) > 0 {
		fmt.Fprintf(&text, "Coverage reasons: %s\n", strings.Join(report.Coverage.Reasons, ", "))
	}
	refs := make(map[string]analysis.Reference)
	for _, ref := range report.Analysis.References {
		refs[ref.ID] = ref
	}
	text.WriteString("Outcomes:\n")
	for _, outcome := range report.Outcomes {
		ref := refs[outcome.ReferenceID]
		fmt.Fprintf(&text, "  - %s %q: %s @ %s\n", outcome.ReferenceID, ref.OriginalName, outcome.Outcome, formatAnalysisLocation(ref.Location))
		evidence, _, _ := marshalCLILine(outcome)
		fmt.Fprintf(&text, "    Evidence: %s", evidence)
	}
	text.WriteString("Diagnostics:\n")
	for _, d := range report.Diagnostics {
		fmt.Fprintf(&text, "  - %s [%s/%s] @ %s: %s\n", d.Code, d.Severity, d.Category, formatAnalysisLocation(d.Location), d.Message)
	}
	return []byte(text.String())
}
