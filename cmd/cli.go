package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

type cliOptions struct {
	schemaOptions        map[string]string
	ocsfProfiles         []string
	ocsfExtensions       []string
	fields               string
	file                 string
	batch                string
	hasFields            bool
	hasFile              bool
	hasBatch             bool
	hasStdin             bool
	config               string
	query                string
	format               string
	output               string
	language             string
	profile              string
	compatibilityVersion string
	sourceID             string
	help                 bool

	hasConfig               bool
	hasQuery                bool
	hasFormat               bool
	hasOutput               bool
	hasLanguage             bool
	hasProfile              bool
	hasCompatibilityVersion bool
	hasSourceID             bool
	hasHelp                 bool
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	return runCLIWithInput(args, os.Stdin, stdout, stderr)
}

func runCLIWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}

	command := args[0]
	switch command {
	case "help", "--help", "-h":
		if len(args) != 1 {
			return writeCLIError(stderr, "text", "help does not accept arguments", 2)
		}
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	case "version", "--version", "-v":
		if len(args) != 1 {
			return writeCLIError(stderr, "text", "version does not accept arguments", 2)
		}
		return writeGeneratedCLIResult([]byte(fmt.Sprintf("spl-toolkit %s\n", buildinfo.Version)), stdout, stderr)
	case "demo":
		if len(args) != 1 {
			return writeCLIError(stderr, "text", "demo does not accept arguments", 2)
		}
		var payload bytes.Buffer
		if err := runDemo(&payload); err != nil {
			return writeCLIError(stderr, "text", err.Error(), 1)
		}
		return writeGeneratedCLIResult(payload.Bytes(), stdout, stderr)
	case "validate-schema":
		return runSchemaValidationCLI(args[1:], stdin, stdout, stderr)
	case "validate-fields":
		return runValidationCLI(args[1:], stdin, stdout, stderr)
	case "map", "discover", "validate", "analyze", "capabilities":
		return runQueryCommand(command, args[1:], stdout, stderr)
	default:
		return writeCLIError(stderr, "text", fmt.Sprintf("unknown command %q", command), 2)
	}
}

func runQueryCommand(command string, args []string, stdout, stderr io.Writer) int {
	options, parseFormat, err := parseCLIOptions(command, args)
	if err != nil {
		return writeCLIError(stderr, parseFormat, err.Error(), 2)
	}
	if options.help {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}
	if err := validateCLIOptions(command, options); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}

	payload, code, err := computeCLIResult(command, options)
	if err != nil {
		return writeCLIError(stderr, options.format, err.Error(), code)
	}
	if err := writeCLIResult(payload, options.output, stdout); err != nil {
		return writeCLIError(stderr, options.format, err.Error(), 2)
	}
	return code
}

func parseCLIOptions(command string, args []string) (cliOptions, string, error) {
	options := cliOptions{format: "text"}
	errorFormat := "text"
	terminated := false

	for i := 0; i < len(args); i++ {
		argument := args[i]
		if !terminated && argument == "--" {
			terminated = true
			continue
		}
		if !terminated && strings.HasPrefix(argument, "--") {
			name, value, hasEquals := strings.Cut(strings.TrimPrefix(argument, "--"), "=")
			if (command == "validate-fields" || command == "validate-schema") && name == "stdin" {
				if hasEquals {
					return options, errorFormat, fmt.Errorf("option --stdin does not accept a value")
				}
				if options.hasStdin {
					return options, errorFormat, fmt.Errorf("duplicate option --stdin")
				}
				options.hasStdin = true
				continue
			}
			if name == "help" {
				if hasEquals {
					return options, errorFormat, fmt.Errorf("option --help does not accept a value")
				}
				if options.hasHelp {
					return options, errorFormat, fmt.Errorf("duplicate option --help")
				}
				options.help, options.hasHelp = true, true
				continue
			}
			if !(command == "validate-schema" && (isSchemaCLIOption(name) || name == "file" || name == "batch")) && !(command == "validate-fields" && (name == "fields" || name == "file" || name == "batch")) && name != "config" && name != "query" && name != "format" && name != "output" && name != "language" && name != "profile" && name != "compatibility-version" && name != "source-id" {
				return options, errorFormat, fmt.Errorf("unknown option --%s", name)
			}
			if !hasEquals {
				if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
					return options, errorFormat, fmt.Errorf("missing value for --%s", name)
				}
				i++
				value = args[i]
			}
			if value == "" && !analysisOptionMayBeEmpty(command, name) {
				return options, errorFormat, fmt.Errorf("missing value for --%s", name)
			}
			if err := setCLIOption(&options, name, value); err != nil {
				return options, errorFormat, err
			}
			if name == "format" {
				if value != "text" && value != "json" {
					return options, errorFormat, fmt.Errorf("unsupported format %q", value)
				}
				errorFormat = value
			}
			continue
		}
		if !terminated && strings.HasPrefix(argument, "-") {
			return options, errorFormat, fmt.Errorf("unknown option %s", argument)
		}
		if options.hasQuery {
			return options, errorFormat, fmt.Errorf("multiple queries supplied")
		}
		options.query, options.hasQuery = argument, true
	}

	return options, errorFormat, nil
}

func analysisOptionMayBeEmpty(command, name string) bool {
	if command != "analyze" && command != "validate-fields" && command != "validate-schema" {
		return false
	}
	switch name {
	case "query", "language", "profile", "compatibility-version", "source-id":
		return true
	default:
		return false
	}
}

func setCLIOption(options *cliOptions, name, value string) error {
	if isSchemaCLIOption(name) {
		return setSchemaCLIOption(options, name, value)
	}
	switch name {
	case "fields":
		if options.hasFields {
			return fmt.Errorf("duplicate option --fields")
		}
		options.fields, options.hasFields = value, true
	case "file":
		if options.hasFile {
			return fmt.Errorf("duplicate option --file")
		}
		options.file, options.hasFile = value, true
	case "batch":
		if options.hasBatch {
			return fmt.Errorf("duplicate option --batch")
		}
		options.batch, options.hasBatch = value, true
	case "config":
		if options.hasConfig {
			return fmt.Errorf("duplicate option --config")
		}
		options.config, options.hasConfig = value, true
	case "query":
		if options.hasQuery {
			return fmt.Errorf("query supplied more than once")
		}
		options.query, options.hasQuery = value, true
	case "format":
		if options.hasFormat {
			return fmt.Errorf("duplicate option --format")
		}
		options.format, options.hasFormat = value, true
	case "output":
		if options.hasOutput {
			return fmt.Errorf("duplicate option --output")
		}
		options.output, options.hasOutput = value, true
	case "language":
		if options.hasLanguage {
			return fmt.Errorf("duplicate option --language")
		}
		options.language, options.hasLanguage = value, true
	case "profile":
		if options.hasProfile {
			return fmt.Errorf("duplicate option --profile")
		}
		options.profile, options.hasProfile = value, true
	case "compatibility-version":
		if options.hasCompatibilityVersion {
			return fmt.Errorf("duplicate option --compatibility-version")
		}
		options.compatibilityVersion, options.hasCompatibilityVersion = value, true
	case "source-id":
		if options.hasSourceID {
			return fmt.Errorf("duplicate option --source-id")
		}
		options.sourceID, options.hasSourceID = value, true
	}
	return nil
}

func validateCLIOptions(command string, options cliOptions) error {
	if command != "analyze" && (options.hasLanguage || options.hasProfile || options.hasCompatibilityVersion || options.hasSourceID) {
		return fmt.Errorf("%s does not accept analysis document options", command)
	}
	switch command {
	case "map":
		if !options.hasConfig {
			return fmt.Errorf("map requires --config")
		}
		if !options.hasQuery {
			return fmt.Errorf("map requires a query")
		}
	case "discover":
		if options.hasConfig {
			return fmt.Errorf("discover does not accept --config")
		}
		if !options.hasQuery {
			return fmt.Errorf("discover requires a query")
		}
	case "validate":
		if options.hasConfig == options.hasQuery {
			return fmt.Errorf("validate requires exactly one of a query or --config")
		}
	case "analyze":
		if options.hasConfig {
			return fmt.Errorf("analyze does not accept --config")
		}
		if !options.hasQuery {
			return fmt.Errorf("analyze requires a query")
		}
	case "capabilities":
		if options.hasConfig || options.hasQuery {
			return fmt.Errorf("capabilities does not accept a query or --config")
		}
	}
	return nil
}

func computeCLIResult(command string, options cliOptions) ([]byte, int, error) {
	switch command {
	case "map":
		config, code, err := readCLIConfig(options.config)
		if err != nil {
			return nil, code, err
		}
		mapped, err := mapper.NewWithConfig(config).MapQuery(options.query)
		if err != nil {
			return nil, 1, err
		}
		if options.format == "json" {
			return marshalCLILine(struct {
				Query string `json:"query"`
			}{Query: mapped})
		}
		return []byte(mapped + "\n"), 0, nil
	case "discover":
		info, err := mapper.New().DiscoverQuery(options.query)
		if err != nil {
			return nil, 1, err
		}
		if options.format == "json" {
			return marshalCLILine(info)
		}
		return formatDiscoveryText(info), 0, nil
	case "validate":
		target := "query"
		if options.hasConfig {
			target = "configuration"
			if _, code, err := readCLIConfig(options.config); err != nil {
				return nil, code, err
			}
		} else if err := mapper.NewParser().ValidateQuery(options.query); err != nil {
			return nil, 1, err
		}
		if options.format == "json" {
			return marshalCLILine(struct {
				Target string `json:"target"`
				Valid  bool   `json:"valid"`
			}{Target: target, Valid: true})
		}
		return []byte("Valid\n"), 0, nil
	case "analyze":
		report, err := analysis.Analyze(analysis.QueryDocument{
			Text: options.query, Language: options.language, Profile: options.profile,
			Version: options.compatibilityVersion, SourceID: options.sourceID,
		})
		if err != nil {
			return nil, 2, err
		}
		code := analysisStatusExitCode(report.Status)
		if options.format == "json" {
			payload, _, err := marshalCLILine(report)
			return payload, code, err
		}
		return formatAnalysisText(report), code, nil
	case "capabilities":
		manifest := analysis.Capabilities()
		if options.format == "json" {
			return marshalCLILine(manifest)
		}
		return formatCapabilitiesText(manifest), 0, nil
	default:
		return nil, 2, fmt.Errorf("unknown query command %q", command)
	}
}

func readCLIConfig(path string) (*mapper.MappingConfig, int, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, 2, fmt.Errorf("read configuration: %w", err)
	}
	config, err := mapper.LoadMappingConfig(contents)
	if err != nil {
		return nil, 1, err
	}
	return config, 0, nil
}

func marshalCLILine(value any) ([]byte, int, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, 1, err
	}
	return append(payload, '\n'), 0, nil
}

func formatDiscoveryText(info *mapper.QueryInfo) []byte {
	var payload strings.Builder
	writeDiscoveryLine := func(label string, values []string) {
		value := "(none)"
		if len(values) > 0 {
			value = strings.Join(values, ", ")
		}
		fmt.Fprintf(&payload, "%s: %s\n", label, value)
	}
	writeDiscoveryLine("Data models", info.DataModels)
	writeDiscoveryLine("Datasets", info.Datasets)
	writeDiscoveryLine("Lookups", info.Lookups)
	writeDiscoveryLine("Macros", info.Macros)
	writeDiscoveryLine("Sources", info.Sources)
	writeDiscoveryLine("Source types", info.SourceTypes)
	writeDiscoveryLine("Input fields", info.InputFields)
	return []byte(payload.String())
}

func writeCLIError(w io.Writer, format, message string, code int) int {
	if format == "json" {
		payload, err := json.Marshal(struct {
			Error string `json:"error"`
		}{Error: message})
		if err == nil {
			_, _ = w.Write(append(payload, '\n'))
		}
		return code
	}
	_, _ = fmt.Fprintf(w, "Error: %s\n", message)
	return code
}

func writeCLIResult(payload []byte, output string, stdout io.Writer) error {
	if output != "" {
		if err := os.WriteFile(output, payload, 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		return nil
	}
	_, err := stdout.Write(payload)
	return err
}

func helpPayload() []byte {
	var payload bytes.Buffer
	if err := showHelp(&payload); err != nil {
		return nil
	}
	return payload.Bytes()
}

func writeGeneratedCLIResult(payload []byte, stdout, stderr io.Writer) int {
	if err := writeCLIResult(payload, "", stdout); err != nil {
		return writeCLIError(stderr, "text", err.Error(), 2)
	}
	return 0
}
