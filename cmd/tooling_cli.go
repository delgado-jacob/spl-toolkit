package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/internal/lsp"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpusio"
	"github.com/delgado-jacob/spl-toolkit/pkg/document"
	"github.com/delgado-jacob/spl-toolkit/pkg/graph"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/sarif"
)

const toolingHelp = `Developer tooling (offline, read-only sources):
  scan --directory DIR | --manifest FILE [--target FILE] [--format text|json|sarif] [--output FILE]
  graph --directory DIR | --manifest FILE [--target FILE] [--format json] [--output FILE]
  impact-schema --directory DIR | --manifest FILE --before-target FILE --after-target FILE [--format text|json] [--output FILE]
  impact-mapping --directory DIR | --manifest FILE --before-rules FILE --after-rules FILE [--before-target FILE] [--after-target FILE] [--format text|json] [--output FILE]
  document QUERY | --query QUERY | --file FILE | --stdin [--language spl|spl2] [--profile splunkd] [--compatibility-version current] [--source-id ID] [--format json] [--output FILE]
  lsp --stdio [--profile splunkd] [--compatibility-version current] [--target FILE]
Targets are canonical inline field_list/json_schema/ocsf JSON wrappers; rule files are versioned rule sets.
Manifest files are versioned documents with literal text or contained relative file paths.
Directory/manifest roots require physical paths without symlink ancestors.
Local source snapshots and reports are memory-resident; no bytes are silently truncated.
Exit: 0 valid, 1 invalid, 3 incomplete, 2 request/acquisition/write/internal failure.
Examples: spl-toolkit scan --directory queries --format sarif --output findings.sarif
          spl-toolkit document 'search host=web' --format json
`

func parseToolingOptions(command string, args []string) (map[string]string, error) {
	allowed := map[string]bool{"help": true}
	boolean := map[string]bool{"help": true, "stdio": true, "stdin": true}
	names := []string{"format", "output", "directory", "manifest"}
	switch command {
	case "scan", "graph":
		names = append(names, "target")
	case "impact-schema":
		names = append(names, "before-target", "after-target")
	case "impact-mapping":
		names = append(names, "before-target", "after-target", "before-rules", "after-rules")
	case "document":
		names = []string{"query", "file", "stdin", "language", "profile", "compatibility-version", "source-id", "format", "output"}
	case "lsp":
		names = []string{"stdio", "profile", "compatibility-version", "target"}
	}
	for _, n := range names {
		allowed[n] = true
	}
	o := map[string]string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" && command == "document" && i+2 == len(args) {
			a = args[i+1]
			i++
			if _, ok := o["query"]; ok {
				return nil, fmt.Errorf("duplicate query")
			}
			o["query"] = a
			continue
		}
		if !strings.HasPrefix(a, "--") {
			if command == "document" && !strings.HasPrefix(a, "-") {
				if _, ok := o["query"]; ok {
					return nil, fmt.Errorf("duplicate query")
				}
				o["query"] = a
				continue
			}
			return nil, fmt.Errorf("unexpected argument %q", a)
		}
		n, v, equal := strings.Cut(strings.TrimPrefix(a, "--"), "=")
		if !allowed[n] {
			return nil, fmt.Errorf("unknown option --%s", n)
		}
		if _, ok := o[n]; ok {
			return nil, fmt.Errorf("duplicate option --%s", n)
		}
		if boolean[n] {
			if equal {
				return nil, fmt.Errorf("--%s does not accept a value", n)
			}
			o[n] = "true"
			continue
		}
		if !equal {
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "--") {
				return nil, fmt.Errorf("missing value for --%s", n)
			}
			v = args[i]
		}
		if v == "" && n != "query" && n != "source-id" && n != "language" && n != "profile" && n != "compatibility-version" {
			return nil, fmt.Errorf("empty --%s", n)
		}
		if (n == "directory" || n == "manifest" || n == "file" || n == "target" || strings.HasSuffix(n, "-target") || strings.HasSuffix(n, "-rules")) && v == "-" {
			return nil, fmt.Errorf("--%s requires a local path", n)
		}
		o[n] = v
	}
	return o, nil
}

func runToolingCLI(command string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	o, err := parseToolingOptions(command, args)
	if err != nil {
		return writeCLIError(stderr, "text", err.Error(), 2)
	}
	if o["help"] != "" {
		return writeGeneratedCLIResult([]byte(toolingHelp), stdout, stderr)
	}
	fail := func(err error) int { return writeCLIError(stderr, o["format"], err.Error(), 2) }
	if command == "lsp" {
		if o["stdio"] == "" {
			return fail(fmt.Errorf("lsp requires --stdio"))
		}
		target, err := readToolingTarget(o["target"])
		if err != nil {
			return fail(err)
		}
		opts := lsp.Options{Profile: o["profile"], Version: o["compatibility-version"], ValidationTarget: target}
		if _, err = corpus.Prepare(corpus.ScanOptions{ValidationTarget: target}); err != nil {
			return fail(err)
		}
		if _, err = capabilityselector.Normalize("", opts.Profile, opts.Version); err != nil {
			return fail(err)
		}
		err = lsp.Serve(context.Background(), stdin, stdout, stderr, opts)
		if errors.Is(err, lsp.ErrExitWithoutShutdown) {
			return 1
		}
		if err != nil {
			return fail(err)
		}
		return 0
	}
	format := o["format"]
	if format == "" {
		format = "text"
		if command == "graph" || command == "document" {
			format = "json"
		}
	}
	if format != "json" && !(format == "text" && command != "graph" && command != "document") && !(format == "sarif" && command == "scan") {
		return fail(fmt.Errorf("unsupported format %q for %s", format, command))
	}
	if command == "document" {
		return runDocumentCLI(o, stdin, stdout, stderr)
	}
	if (o["directory"] == "") == (o["manifest"] == "") {
		return fail(fmt.Errorf("exactly one --directory or --manifest is required"))
	}
	inputs := []string{}
	for _, k := range []string{"manifest", "target", "before-target", "after-target", "before-rules", "after-rules"} {
		if o[k] != "" {
			inputs = append(inputs, o[k])
		}
	}
	// Preflight shared targets and rules before the loader is called, even when
	// all selected files would fail acquisition.
	var scan *corpus.PreparedScan
	var schemas *impact.PreparedSchemaComparison
	var mappings *impact.PreparedMappingComparison
	switch command {
	case "scan", "graph":
		var target *corpus.ValidationTarget
		target, err = readToolingTarget(o["target"])
		if err == nil {
			scan, err = corpus.Prepare(corpus.ScanOptions{ValidationTarget: target})
		}
	case "impact-schema":
		if o["before-target"] == "" || o["after-target"] == "" {
			return fail(fmt.Errorf("both --before-target and --after-target are required"))
		}
		var before, after *corpus.ValidationTarget
		before, err = readToolingTarget(o["before-target"])
		if err == nil {
			after, err = readToolingTarget(o["after-target"])
		}
		if err == nil {
			schemas, err = impact.PrepareSchemas(*before, *after)
		}
	case "impact-mapping":
		var before, after impact.MappingSide
		before, err = readToolingMapping(o["before-rules"], o["before-target"])
		if err == nil {
			after, err = readToolingMapping(o["after-rules"], o["after-target"])
		}
		if err == nil {
			mappings, err = impact.PrepareMappings(before, after)
		}
	}
	if err != nil {
		return fail(err)
	}
	if err = checkToolingOutput(o["output"], inputs); err != nil {
		return fail(err)
	}
	var input corpus.Input
	root := o["directory"]
	if root != "" {
		input, err = corpusio.LoadDirectory(root)
	} else {
		var raw []byte
		raw, err = os.ReadFile(o["manifest"])
		if err == nil {
			var m corpusio.Manifest
			m, err = corpusio.DecodeManifest(raw)
			if err == nil {
				root = filepath.Join(filepath.Dir(o["manifest"]), m.Base)
				for _, e := range m.Documents {
					if e.Path != nil {
						inputs = append(inputs, filepath.Join(root, filepath.FromSlash(*e.Path)))
					}
				}
				input, err = corpusio.LoadManifest(m, o["manifest"])
			}
		}
	}
	if err != nil {
		return fail(err)
	}
	for _, e := range input.Entries {
		if e.Origin.Kind == "file" {
			inputs = append(inputs, filepath.Join(root, filepath.FromSlash(e.Origin.RelativePath)))
		}
	}
	var value any
	code := 0
	if scan != nil {
		var report *corpus.Report
		report, err = scan.Scan(input)
		if err == nil {
			code = analysisStatusExitCode(report.Status)
			if !report.ExecutionComplete {
				code = 2
			}
			value = report
			if command == "graph" {
				value, err = graph.Export(report)
			} else if format == "sarif" {
				value, err = sarif.Export(report)
			}
		}
	} else {
		var report *impact.Report
		if schemas != nil {
			report, err = schemas.Compare(input)
		} else {
			report, err = mappings.Compare(input)
		}
		if err == nil {
			value = report
			code = impactCLIExit(report)
		}
	}
	if err != nil {
		return fail(err)
	}
	payload, _, err := marshalCLILine(value)
	if err != nil {
		return fail(err)
	}
	if format == "text" {
		payload = formatToolingText(value)
	}
	if err = writeToolingOutput(payload, o["output"], inputs, stdout); err != nil {
		return fail(err)
	}
	return code
}

func formatToolingText(value any) []byte {
	var text strings.Builder
	switch report := value.(type) {
	case *corpus.Report:
		fmt.Fprintf(&text, "Status: %s\nExecution complete: %t\nSelected: %d; analyzed: %d; acquisition failures: %d; traversal failures: %d\n", report.Status, report.ExecutionComplete, report.Counts.Selected, report.Counts.Analyzed, report.Counts.AcquisitionFailed, report.Counts.TraversalFailed)
		coverage, _, _ := marshalCLILine(report.Coverage)
		fmt.Fprintf(&text, "Coverage: %s", coverage)
		for _, failure := range report.Selection.TraversalFailures {
			fmt.Fprintf(&text, "Traversal failure %q [%s]: %s\n", failure.Path, failure.Code, failure.Message)
		}
		for _, entry := range report.Entries {
			fmt.Fprintf(&text, "\nDocument %q:\n", entry.ID)
			if entry.Failure != nil {
				fmt.Fprintf(&text, "Acquisition failure [%s]: %s\n", entry.Failure.Code, entry.Failure.Message)
				continue
			}
			e := entry.Evaluation
			switch {
			case e.Analysis != nil:
				text.Write(formatAnalysisText(e.Analysis))
			case e.FieldValidation != nil:
				text.Write(formatValidationText(e.FieldValidation))
			case e.SchemaValidation != nil:
				text.Write(formatSchemaValidationText(e.SchemaValidation))
			}
		}
	case *impact.Report:
		fmt.Fprintf(&text, "Impact: %s\nExecution complete: %t\nSelected: %d; compared: %d; affected: %d; unchanged: %d; indeterminate: %d; failed: %d\n", report.Kind, report.ExecutionComplete, report.Counts.Selected, report.Counts.Compared, report.Counts.Affected, report.Counts.Unchanged, report.Counts.Indeterminate, report.Counts.Failed)
		for _, failure := range report.Selection.TraversalFailures {
			fmt.Fprintf(&text, "Traversal failure %q [%s]: %s\n", failure.Path, failure.Code, failure.Message)
		}
		for _, entry := range report.Entries {
			fmt.Fprintf(&text, "\nDocument %q: %s (%s -> %s)\nReasons: %s\n", entry.ID, entry.Classification, entry.BeforeStatus, entry.AfterStatus, strings.Join(entry.Reasons, ", "))
			if entry.Failure != nil {
				fmt.Fprintf(&text, "Acquisition failure [%s]: %s\n", entry.Failure.Code, entry.Failure.Message)
			}
			for _, delta := range entry.Deltas {
				raw, _, _ := marshalCLILine(delta)
				fmt.Fprintf(&text, "  Evidence: %s", raw)
			}
		}
	}
	return []byte(text.String())
}

func readToolingTarget(path string) (*corpus.ValidationTarget, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read target: %w", err)
	}
	return corpus.DecodeValidationTarget(raw)
}
func readToolingMapping(rulesPath, targetPath string) (impact.MappingSide, error) {
	if rulesPath == "" {
		return impact.MappingSide{}, fmt.Errorf("both --before-rules and --after-rules are required")
	}
	raw, err := os.ReadFile(rulesPath)
	if err != nil {
		return impact.MappingSide{}, err
	}
	rules, err := rewrite.DecodeRuleSet(raw)
	if err != nil {
		return impact.MappingSide{}, err
	}
	target, err := readToolingTarget(targetPath)
	return impact.MappingSide{Rules: rules, ValidationTarget: target}, err
}

func impactCLIExit(r *impact.Report) int {
	if !r.ExecutionComplete {
		return 2
	}
	status := analysis.Valid
	for _, e := range r.Entries {
		for _, s := range []analysis.Status{e.BeforeStatus, e.AfterStatus} {
			if s == analysis.Invalid {
				status = analysis.Invalid
			} else if s == analysis.Incomplete && status != analysis.Invalid {
				status = analysis.Incomplete
			}
		}
	}
	if status == analysis.Invalid {
		return 1
	}
	if r.Counts.Indeterminate > 0 {
		return 3
	}
	return analysisStatusExitCode(status)
}

func runDocumentCLI(o map[string]string, stdin io.Reader, stdout, stderr io.Writer) int {
	fail := func(err error) int { return writeCLIError(stderr, o["format"], err.Error(), 2) }
	count := 0
	for _, key := range []string{"query", "file", "stdin"} {
		if _, ok := o[key]; ok {
			count++
		}
	}
	if count != 1 {
		return fail(fmt.Errorf("document requires exactly one query, --file or --stdin"))
	}
	d := analysis.QueryDocument{Text: o["query"], Language: o["language"], Profile: o["profile"], Version: o["compatibility-version"], SourceID: o["source-id"]}
	inputs := []string{}
	if o["file"] != "" || o["stdin"] != "" {
		var raw []byte
		var err error
		if o["file"] != "" {
			inputs = append(inputs, o["file"])
			raw, err = os.ReadFile(o["file"])
			d.SourceID = o["file"]
		} else {
			raw, err = io.ReadAll(stdin)
			d.SourceID = "<stdin>"
		}
		if err != nil {
			return fail(err)
		}
		d.Text = string(raw)
		if id, ok := o["source-id"]; ok {
			d.SourceID = id
		}
	}
	result, err := analysis.Analyze(d)
	if err != nil {
		return fail(err)
	}
	view, err := document.New(result, document.RevisionContext{ToolVersion: buildinfo.Version, ContractVersion: "1"})
	if err != nil {
		return fail(err)
	}
	payload, _, err := marshalCLILine(view)
	if err == nil {
		err = writeToolingOutput(payload, o["output"], inputs, stdout)
	}
	if err != nil {
		return fail(err)
	}
	return analysisStatusExitCode(result.Status)
}

// Resolve existing aliases and compare file identities, then publish by rename
// instead of truncating an opened output that might be a source hardlink.
func checkToolingOutput(output string, inputs []string) error {
	if output == "" {
		return nil
	}
	outAbs, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	outReal, err := filepath.EvalSymlinks(outAbs)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err != nil {
		parent, e := filepath.EvalSymlinks(filepath.Dir(outAbs))
		if e != nil {
			return e
		}
		outReal = filepath.Join(parent, filepath.Base(outAbs))
	}
	outInfo, err := os.Stat(output)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, input := range inputs {
		abs, e := filepath.Abs(input)
		if e != nil {
			return e
		}
		real, e := filepath.EvalSymlinks(abs)
		if e != nil {
			real = abs
		}
		info, e := os.Stat(input)
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if outAbs == abs || outReal == real || (outInfo != nil && info != nil && os.SameFile(outInfo, info)) {
			return fmt.Errorf("output must not overwrite input %q", input)
		}
	}
	return nil
}
func writeToolingOutput(payload []byte, output string, inputs []string, stdout io.Writer) error {
	if output == "" {
		return writeCLIResult(payload, "", stdout)
	}
	if err := checkToolingOutput(output, inputs); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(output), ".spl-toolkit-output-*")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.Write(payload); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = checkToolingOutput(output, inputs); err != nil {
		return err
	}
	return os.Rename(name, output)
}
