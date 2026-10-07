package main

import (
	"fmt"
	"io"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

func runResolutionCLI(args []string, stdout, stderr io.Writer) int {
	options, err := parseToolingOptions("resolve", args)
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
	fail := func(err error) int { return writeResolutionCLIError(stderr, format, err) }
	if options["request"] == "" || options["request"] == "-" {
		return fail(fmt.Errorf("resolve requires --request with one local path (not '-')"))
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
	report, err := resolution.ResolveJSON(raw)
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
		payload = []byte(resolution.FormatReport(report))
	}
	if err := writeToolingOutput(payload, options["output"], inputs, stdout); err != nil {
		return fail(err)
	}
	return resolutionCLIExit(report)
}

func writeResolutionCLIError(stderr io.Writer, format string, err error) int {
	if detail, ok := resolution.RequestErrorDetails(err); ok {
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
func resolutionCLIExit(report *resolution.Report) int {
	if report.Counts.Failed > 0 {
		return 1
	}
	if report.Counts.Incomplete > 0 {
		return 3
	}
	return 0
}
