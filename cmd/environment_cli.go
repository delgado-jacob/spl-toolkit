package main

import (
	"fmt"
	"io"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func runEnvironmentCLI(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}
	if len(args) == 0 || args[0] != "validate" {
		return writeCLIError(stderr, "text", "environment requires validate", 2)
	}
	o, err := parseToolingOptions("environment-validate", args[1:])
	if err != nil {
		return writeCLIError(stderr, "text", err.Error(), 2)
	}
	if o["help"] != "" {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}
	fail := func(err error) int { return writeCLIError(stderr, o["format"], err.Error(), 2) }
	if o["snapshot"] == "" && o["schemas"] == "" {
		return fail(fmt.Errorf("environment validate requires --snapshot or --schemas"))
	}
	if o["snapshot"] == "-" || o["schemas"] == "-" {
		return fail(fmt.Errorf("environment artifacts require local paths (not '-')"))
	}
	format := o["format"]
	if format == "" {
		format = "text"
	}
	if format != "text" && format != "json" {
		return fail(fmt.Errorf("unsupported format %q", format))
	}
	inputs := []string{}
	for _, key := range []string{"snapshot", "schemas"} {
		if o[key] != "" {
			inputs = append(inputs, o[key])
		}
	}
	if err := checkToolingOutput(o["output"], inputs); err != nil {
		return fail(err)
	}
	var snapshotJSON, schemaBundleJSON []byte
	if o["snapshot"] != "" {
		snapshotJSON, err = os.ReadFile(o["snapshot"])
		if err != nil {
			return fail(fmt.Errorf("read snapshot: %w", err))
		}
	}
	if o["schemas"] != "" {
		schemaBundleJSON, err = os.ReadFile(o["schemas"])
		if err != nil {
			return fail(fmt.Errorf("read schemas: %w", err))
		}
	}
	report, err := environment.ValidateArtifacts(snapshotJSON, schemaBundleJSON)
	if err != nil {
		return fail(err)
	}
	var payload []byte
	if format == "json" {
		payload, _, err = marshalCLILine(report)
		if err != nil {
			return fail(err)
		}
	} else {
		payload = []byte(environment.FormatReport(report))
	}
	if err := writeToolingOutput(payload, o["output"], inputs, stdout); err != nil {
		return fail(err)
	}
	switch report.Status {
	case "valid":
		return 0
	case "partial":
		return 3
	default:
		return 1
	}
}
