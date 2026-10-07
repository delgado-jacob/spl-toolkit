package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpusio"
	"github.com/delgado-jacob/spl-toolkit/pkg/workflow"
)

func parseWorkflowOptions(command string, args []string) (map[string]string, []string, error) {
	allowed := map[string]bool{"output": true}
	switch command {
	case "assess":
		for _, k := range []string{"request", "directory", "manifest", "settings", "format"} {
			allowed[k] = true
		}
	case "compare":
		allowed["before"], allowed["after"], allowed["format"] = true, true, true
	case "evidence":
		allowed["report"], allowed["comparison"], allowed["include"] = true, true, true
	case "recheck":
		allowed["request"], allowed["format"] = true, true
	default:
		return nil, nil, fmt.Errorf("unknown workflow operation %q", command)
	}
	options, includes := map[string]string{}, []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return nil, nil, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, equals := strings.Cut(strings.TrimPrefix(arg, "--"), "=")
		if !allowed[name] {
			return nil, nil, fmt.Errorf("unknown option --%s", name)
		}
		if !equals {
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "--") {
				return nil, nil, fmt.Errorf("missing value for --%s", name)
			}
			i++
			value = args[i]
		}
		if value == "" {
			return nil, nil, fmt.Errorf("missing value for --%s", name)
		}
		if name == "include" {
			includes = append(includes, value)
			continue
		}
		if _, ok := options[name]; ok {
			return nil, nil, fmt.Errorf("duplicate option --%s", name)
		}
		options[name] = value
	}
	switch command {
	case "assess":
		sources := 0
		for _, k := range []string{"request", "directory", "manifest"} {
			if options[k] != "" {
				sources++
			}
		}
		if sources != 1 {
			return nil, nil, fmt.Errorf("assess requires exactly one of --request, --directory or --manifest")
		}
		if (options["request"] != "") == (options["settings"] != "") {
			return nil, nil, fmt.Errorf("--settings is required only for directory or manifest assessment")
		}
	case "compare":
		if options["before"] == "" || options["after"] == "" {
			return nil, nil, fmt.Errorf("compare requires --before and --after")
		}
	case "evidence":
		if (options["report"] == "") == (options["comparison"] == "") {
			return nil, nil, fmt.Errorf("evidence requires exactly one of --report or --comparison")
		}
	case "recheck":
		if options["request"] == "" {
			return nil, nil, fmt.Errorf("recheck requires --request")
		}
	}
	if format := options["format"]; format != "" && format != "text" && format != "json" && !(command == "assess" && (format == "sarif" || format == "graph" || format == "bom")) {
		return nil, nil, fmt.Errorf("unsupported format %q", format)
	}
	return options, includes, nil
}

func runWorkflowCLI(args []string, stdout, stderr io.Writer) int {
	fail := func(err error) int { return writeCLIError(stderr, "text", err.Error(), 2) }
	if len(args) == 0 {
		return fail(fmt.Errorf("workflow requires assess, compare, evidence or recheck"))
	}
	command := args[0]
	options, includes, err := parseWorkflowOptions(command, args[1:])
	if err != nil {
		return fail(err)
	}
	inputs := []string{}
	for _, k := range []string{"request", "settings", "manifest", "before", "after", "report", "comparison"} {
		if options[k] != "" {
			inputs = append(inputs, options[k])
		}
	}
	if err = checkToolingOutput(options["output"], inputs); err != nil {
		return fail(err)
	}
	format := options["format"]
	if format == "" {
		format = "text"
	}
	var value any
	code := 0
	switch command {
	case "assess":
		var report *workflow.Report
		if options["request"] != "" {
			raw, e := os.ReadFile(options["request"])
			if e != nil {
				return fail(e)
			}
			request, e := workflow.DecodeRequest(raw)
			if e != nil {
				return fail(e)
			}
			if options["format"] == "" && request.Format != "" {
				format = request.Format
			}
			request.Format = format
			report, err = workflow.Assess(request)
		} else {
			raw, e := os.ReadFile(options["settings"])
			if e != nil {
				return fail(e)
			}
			settings, e := workflow.DecodeSettings(raw)
			if e != nil {
				return fail(e)
			}
			prepared, e := workflow.Prepare(settings)
			if e != nil {
				return fail(e)
			}
			var input corpus.Input
			root := options["directory"]
			if root != "" {
				input, err = corpusio.LoadDirectory(root)
			} else {
				raw, e = os.ReadFile(options["manifest"])
				if e != nil {
					return fail(e)
				}
				manifest, e := corpusio.DecodeManifest(raw)
				if e != nil {
					return fail(e)
				}
				root = filepath.Join(filepath.Dir(options["manifest"]), manifest.Base)
				for _, entry := range manifest.Documents {
					if entry.Path != nil {
						inputs = append(inputs, filepath.Join(root, filepath.FromSlash(*entry.Path)))
					}
				}
				input, err = corpusio.LoadManifest(manifest, options["manifest"])
			}
			if err != nil {
				return fail(err)
			}
			for _, entry := range input.Entries {
				if entry.Origin.Kind == "file" {
					inputs = append(inputs, filepath.Join(root, filepath.FromSlash(entry.Origin.RelativePath)))
				}
			}
			if err = checkToolingOutput(options["output"], inputs); err != nil {
				return fail(err)
			}
			report, err = prepared.Assess(input)
		}
		if err == nil {
			code = report.CIExitCode
			value, err = workflow.RenderAssessment(report, format)
		}
	case "compare":
		before, e := os.ReadFile(options["before"])
		if e != nil {
			return fail(e)
		}
		after, e := os.ReadFile(options["after"])
		if e != nil {
			return fail(e)
		}
		raw, e := json.Marshal(map[string]any{"schema_version": 1, "before": json.RawMessage(before), "after": json.RawMessage(after)})
		if e != nil {
			return fail(e)
		}
		report, e := workflow.CompareJSON(raw)
		err = e
		if err == nil {
			code = report.CIExitCode
			value = report
			if format == "text" {
				value = workflow.FormatComparison(report)
			}
		}
	case "evidence":
		kind := "report"
		if options[kind] == "" {
			kind = "comparison"
		}
		source, e := os.ReadFile(options[kind])
		if e != nil {
			return fail(e)
		}
		raw, e := json.Marshal(map[string]any{"schema_version": 1, kind: json.RawMessage(source), "include": includes})
		if e != nil {
			return fail(e)
		}
		value, err = workflow.EvidenceJSON(raw)
	case "recheck":
		raw, e := os.ReadFile(options["request"])
		if e != nil {
			return fail(e)
		}
		report, e := workflow.RecheckJSON(raw)
		err = e
		if err == nil {
			code = report.Assessment.CIExitCode
			value = report
			if format == "text" {
				value = workflow.FormatRecheck(report)
			}
		}
	}
	if err != nil {
		return fail(err)
	}
	var payload []byte
	if text, ok := value.(string); ok {
		payload = []byte(text)
	} else {
		payload, err = json.Marshal(value)
		payload = append(payload, '\n')
	}
	if err != nil {
		return fail(err)
	}
	if err = writeToolingOutput(payload, options["output"], inputs, stdout); err != nil {
		return fail(err)
	}
	return code
}
