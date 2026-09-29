package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

// runClosureCLI adapts local inputs to the canonical inline request. Source
// paths are transport details and never become part of the evaluated document.
func runClosureCLI(o map[string]string, stdin io.Reader, stdout, stderr io.Writer) int {
	if o["help"] != "" {
		return writeGeneratedCLIResult([]byte(toolingHelp), stdout, stderr)
	}
	fail := func(err error) int { return writeCLIError(stderr, o["format"], err.Error(), 2) }
	count := 0
	for _, key := range []string{"query", "file", "stdin"} {
		if _, ok := o[key]; ok {
			count++
		}
	}
	if count != 1 {
		return fail(fmt.Errorf("closure requires exactly one query, --file, or --stdin"))
	}
	if o["bundle"] == "" {
		return fail(fmt.Errorf("closure requires --bundle"))
	}
	format := o["format"]
	if format == "" {
		format = "text"
	}
	if format != "text" && format != "json" && format != "graph" && format != "bom" {
		return fail(fmt.Errorf("unsupported format %q for closure", format))
	}
	inputs := []string{o["bundle"]}
	if o["bindings"] != "" {
		inputs = append(inputs, o["bindings"])
	}
	if o["file"] != "" {
		inputs = append(inputs, o["file"])
	}
	if err := checkToolingOutput(o["output"], inputs); err != nil {
		return fail(err)
	}
	query := o["query"]
	if o["file"] != "" || o["stdin"] != "" {
		var raw []byte
		var err error
		if o["file"] != "" {
			raw, err = os.ReadFile(o["file"])
		} else {
			raw, err = io.ReadAll(stdin)
		}
		if err != nil {
			return fail(fmt.Errorf("read query: %w", err))
		}
		query = string(raw)
	}
	if !utf8.ValidString(query) {
		return fail(fmt.Errorf("query must be valid UTF-8"))
	}
	for _, name := range []string{"source-id", "language", "profile", "compatibility-version"} {
		if !utf8.ValidString(o[name]) {
			return fail(fmt.Errorf("--%s must be valid UTF-8", name))
		}
	}
	bundle, err := os.ReadFile(o["bundle"])
	if err != nil {
		return fail(fmt.Errorf("read bundle: %w", err))
	}
	var bindings json.RawMessage
	if o["bindings"] != "" {
		bindings, err = os.ReadFile(o["bindings"])
		if err != nil {
			return fail(fmt.Errorf("read bindings: %w", err))
		}
	}
	raw, err := json.Marshal(struct {
		SchemaVersion int                    `json:"schema_version"`
		Document      analysis.QueryDocument `json:"document"`
		Bundle        json.RawMessage        `json:"bundle"`
		Bindings      json.RawMessage        `json:"bindings,omitempty"`
	}{1, analysis.QueryDocument{Text: query, Language: o["language"], Profile: o["profile"], Version: o["compatibility-version"], SourceID: o["source-id"]}, bundle, bindings})
	if err != nil {
		return fail(err)
	}
	request, err := closure.DecodeRequest(raw)
	if err != nil {
		return fail(err)
	}
	report, err := closure.Evaluate(request)
	if err != nil {
		return fail(err)
	}
	var payload []byte
	switch format {
	case "json":
		payload, _, err = marshalCLILine(report)
	case "graph":
		payload, _, err = marshalCLILine(report.Graph)
	case "bom":
		payload, _, err = marshalCLILine(report.BOM)
	default:
		payload = []byte(closure.FormatInventory(report))
	}
	if err != nil {
		return fail(err)
	}
	if err := writeToolingOutput(payload, o["output"], inputs, stdout); err != nil {
		return fail(err)
	}
	return analysisStatusExitCode(report.Status)
}
