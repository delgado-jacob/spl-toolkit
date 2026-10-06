package main

import (
	"fmt"
	"io"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
)

func runCompatibilityCLI(args []string, stdout, stderr io.Writer) int {
	options, err := parseToolingOptions("compatibility", args)
	if err != nil {
		return writeCLIError(stderr, "text", err.Error(), 2)
	}
	if options["help"] != "" {
		return writeGeneratedCLIResult(helpPayload(), stdout, stderr)
	}
	format := options["format"]
	if format == "" {
		format = "text"
	}
	fail := func(err error) int { return writeCompatibilityCLIError(stderr, format, err) }
	if options["request"] == "" || options["request"] == "-" {
		return fail(fmt.Errorf("compatibility requires --request with one local path (not '-')"))
	}
	if format != "text" && format != "json" {
		return fail(fmt.Errorf("unsupported format %q", format))
	}
	inputs := []string{options["request"]}
	if err := checkToolingOutput(options["output"], inputs); err != nil {
		return fail(err)
	}
	raw, err := os.ReadFile(options["request"])
	if err != nil {
		return fail(fmt.Errorf("read request: %w", err))
	}
	report, err := compatibility.CheckJSON(raw)
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
		payload = []byte(compatibility.FormatReport(report))
	}
	if err := writeToolingOutput(payload, options["output"], inputs, stdout); err != nil {
		return fail(err)
	}
	return compatibilityCLIExit(report, options["require-connected"] != "")
}

func writeCompatibilityCLIError(stderr io.Writer, format string, err error) int {
	if detail, ok := compatibility.RequestErrorDetails(err); ok {
		if format == "json" {
			payload, _, encodeErr := marshalCLILine(detail)
			if encodeErr == nil {
				_, _ = stderr.Write(payload)
			}
			return 2
		}
		message := detail.Code + ": " + detail.Message
		if detail.Path != "" {
			message = detail.Code + " at " + detail.Path + ": " + detail.Message
		}
		return writeCLIError(stderr, "text", message, 2)
	}
	return writeCLIError(stderr, format, err.Error(), 2)
}

// Policy selects an exit only after the complete canonical report is written.
func compatibilityCLIExit(report *compatibility.Report, requireConnected bool) int {
	if report.Outcome == "unsatisfied" || (requireConnected && report.Correlation.Outcome == "disconnected") {
		return 1
	}
	if report.Outcome == "incomplete" || report.Outcome == "not assessed" || (requireConnected && report.Correlation.Outcome == "indeterminate") {
		return 3
	}
	return 0
}
