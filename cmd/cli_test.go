package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeCLIConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mapping.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runCLITest(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := runCLI(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestCLIMapLoadsConfig(t *testing.T) {
	path := writeCLIConfig(t, `{"version":"1.0","mappings":[{"source":"src_ip","target":"source_ip"}]}`)
	code, stdout, stderr := runCLITest("map", "--config", path, "--query", "search src_ip=1")
	if code != 0 || stdout != "search source_ip=1\n" || stderr != "" {
		t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
	}
}

func TestCLIUsageAndContentErrors(t *testing.T) {
	valid := writeCLIConfig(t, `{"version":"1.0","mappings":[]}`)
	malformed := writeCLIConfig(t, `{`)
	invalidRegex := writeCLIConfig(t, `{"version":"1.0","mappings":[],"rules":[{"id":"bad","conditions":[{"type":"source","operator":"regex","value":"["}],"mappings":[{"source":"a","target":"b"}],"priority":1,"enabled":true}]}`)
	missing := filepath.Join(t.TempDir(), "missing.json")

	tests := []struct {
		name string
		args []string
		code int
	}{
		{name: "map requires config", args: []string{"map", "search a=1"}, code: 2},
		{name: "missing config file", args: []string{"map", "--config", missing, "search a=1"}, code: 2},
		{name: "malformed config", args: []string{"map", "--config", malformed, "search a=1"}, code: 1},
		{name: "invalid regex", args: []string{"map", "--config", invalidRegex, "search a=1"}, code: 1},
		{name: "invalid query", args: []string{"map", "--config", valid, "|"}, code: 1},
		{name: "query twice", args: []string{"map", "search a=1", "--query", "search b=2", "--config", valid}, code: 2},
		{name: "unknown flag", args: []string{"discover", "--unknown", "search a=1"}, code: 2},
		{name: "extra positional", args: []string{"discover", "search a=1", "search b=2"}, code: 2},
		{name: "unsupported format", args: []string{"discover", "--query", "search a=1", "--format", "yaml"}, code: 2},
		{name: "missing value", args: []string{"discover", "--query"}, code: 2},
		{name: "duplicate option", args: []string{"discover", "--query", "search a=1", "--format", "text", "--format=json"}, code: 2},
		{name: "validate requires one target", args: []string{"validate"}, code: 2},
		{name: "validate rejects both targets", args: []string{"validate", "--config", valid, "--query", "search a=1"}, code: 2},
		{name: "discover rejects config", args: []string{"discover", "--config", valid, "--query", "search a=1"}, code: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLITest(test.args...)
			if code != test.code || stdout != "" || stderr == "" {
				t.Fatalf("code=%d out=%q err=%q; want code=%d, empty stdout, non-empty stderr", code, stdout, stderr, test.code)
			}
		})
	}
}

func TestCLIOptionFormsAndTermination(t *testing.T) {
	path := writeCLIConfig(t, `{"version":"1.0","mappings":[{"source":"src_ip","target":"source_ip"}]}`)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "options before positional", args: []string{"map", "--config", path, "search src_ip=1", "--format", "text"}, want: "search source_ip=1\n"},
		{name: "options after positional", args: []string{"map", "search src_ip=1", "--config", path}, want: "search source_ip=1\n"},
		{name: "equals form", args: []string{"map", "--config=" + path, "--query=search src_ip=1", "--format=text"}, want: "search source_ip=1\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLITest(test.args...)
			if code != 0 || stdout != test.want || stderr != "" {
				t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
			}
		})
	}

	code, stdout, stderr := runCLITest("validate", "--", "-search")
	if code != 1 || stdout != "" || strings.Contains(stderr, "unknown option") {
		t.Fatalf("-- did not terminate options: code=%d out=%q err=%q", code, stdout, stderr)
	}
}

func TestCLIMapJSONAndUnchangedQuery(t *testing.T) {
	mapped := writeCLIConfig(t, `{"version":"1.0","mappings":[{"source":"src_ip","target":"source_ip"}]}`)
	code, stdout, stderr := runCLITest("map", "--config", mapped, "--query", "search src_ip=1", "--format", "json")
	if code != 0 || stdout != "{\"query\":\"search source_ip=1\"}\n" || stderr != "" {
		t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
	}

	empty := writeCLIConfig(t, `{"version":"1.0","mappings":[]}`)
	code, stdout, stderr = runCLITest("map", "--config", empty, "search src_ip=1")
	if code != 0 || stdout != "search src_ip=1\n" || stderr != "" {
		t.Fatalf("unchanged map: code=%d out=%q err=%q", code, stdout, stderr)
	}
}

func TestCLIValidateQueryAndConfiguration(t *testing.T) {
	config := writeCLIConfig(t, `{"version":"1.0","mappings":[]}`)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "query text", args: []string{"validate", "--query", "search a=1"}, want: "Valid\n"},
		{name: "query json", args: []string{"validate", "--query", "search a=1", "--format", "json"}, want: "{\"target\":\"query\",\"valid\":true}\n"},
		{name: "configuration text", args: []string{"validate", "--config", config}, want: "Valid\n"},
		{name: "configuration json", args: []string{"validate", "--config", config, "--format=json"}, want: "{\"target\":\"configuration\",\"valid\":true}\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLITest(test.args...)
			if code != 0 || stdout != test.want || stderr != "" {
				t.Fatalf("code=%d out=%q err=%q", code, stdout, stderr)
			}
		})
	}
}

func TestCLIDiscoveryFormatsAllSevenFields(t *testing.T) {
	query := "search sourcetype=web source=access.log src_ip=1"
	code, stdout, stderr := runCLITest("discover", "--query", query, "--format", "json")
	wantJSON := "{\"datamodels\":[],\"datasets\":[],\"lookups\":[],\"macros\":[],\"sources\":[\"access.log\"],\"sourcetypes\":[\"web\"],\"input_fields\":[\"sourcetype\",\"source\",\"src_ip\"]}\n"
	if code != 0 || stdout != wantJSON || stderr != "" {
		t.Fatalf("json: code=%d out=%q err=%q", code, stdout, stderr)
	}

	code, stdout, stderr = runCLITest("discover", "--query", query)
	wantText := "Data models: (none)\nDatasets: (none)\nLookups: (none)\nMacros: (none)\nSources: access.log\nSource types: web\nInput fields: sourcetype, source, src_ip\n"
	if code != 0 || stdout != wantText || stderr != "" {
		t.Fatalf("text: code=%d out=%q err=%q", code, stdout, stderr)
	}
}

func TestCLIErrorFormatFollowsSuccessfullyParsedFormat(t *testing.T) {
	malformed := writeCLIConfig(t, `{`)
	code, stdout, stderr := runCLITest("map", "--format=json", "--config", malformed, "--query", "search a=1")
	if code != 1 || stdout != "" || !strings.HasPrefix(stderr, "{\"error\":") || !strings.HasSuffix(stderr, "}\n") {
		t.Fatalf("json content error: code=%d out=%q err=%q", code, stdout, stderr)
	}

	code, stdout, stderr = runCLITest("discover", "--format=json", "--unknown")
	if code != 2 || stdout != "" || !strings.HasPrefix(stderr, "{\"error\":") {
		t.Fatalf("json usage error after format: code=%d out=%q err=%q", code, stdout, stderr)
	}

	code, stdout, stderr = runCLITest("discover", "--unknown", "--format=json")
	if code != 2 || stdout != "" || strings.HasPrefix(stderr, "{") {
		t.Fatalf("early parse error should be text: code=%d out=%q err=%q", code, stdout, stderr)
	}
}

type failingCLIWriter struct{}

func (failingCLIWriter) Write([]byte) (int, error) {
	return 0, errors.New("write rejected")
}

func TestCLIOutputFileAndWriteFailures(t *testing.T) {
	config := writeCLIConfig(t, `{"version":"1.0","mappings":[{"source":"src_ip","target":"source_ip"}]}`)
	output := filepath.Join(t.TempDir(), "result.txt")
	code, stdout, stderr := runCLITest("map", "--config", config, "search src_ip=1", "--output", output)
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("file result: code=%d out=%q err=%q", code, stdout, stderr)
	}
	contents, err := os.ReadFile(output)
	if err != nil || string(contents) != "search source_ip=1\n" {
		t.Fatalf("output file = %q, %v", contents, err)
	}

	preserved := filepath.Join(t.TempDir(), "preserved.txt")
	if err := os.WriteFile(preserved, []byte("keep me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runCLITest("map", "--config", config, "--query", "|", "--output", preserved)
	contents, err = os.ReadFile(preserved)
	if code != 1 || stdout != "" || stderr == "" || err != nil || string(contents) != "keep me\n" {
		t.Fatalf("rejected input changed file: code=%d out=%q err=%q contents=%q readErr=%v", code, stdout, stderr, contents, err)
	}

	var errOut bytes.Buffer
	code = runCLI([]string{"validate", "--query", "search a=1"}, failingCLIWriter{}, &errOut)
	if code != 2 || !strings.Contains(errOut.String(), "write rejected") {
		t.Fatalf("stdout failure: code=%d err=%q", code, errOut.String())
	}

	badOutput := filepath.Join(t.TempDir(), "missing", "result.txt")
	code, stdout, stderr = runCLITest("validate", "--query", "search a=1", "--output", badOutput)
	if code != 2 || stdout != "" || stderr == "" {
		t.Fatalf("file failure: code=%d out=%q err=%q", code, stdout, stderr)
	}
}

func TestCLIHelpAndVersionNeedNoQuery(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"--help"}} {
		code, stdout, stderr := runCLITest(args...)
		if code != 0 || !strings.Contains(stdout, "Usage: spl-toolkit <command> [options]") || stderr != "" {
			t.Fatalf("help %v: code=%d out=%q err=%q", args, code, stdout, stderr)
		}
	}

	for _, args := range [][]string{{"version"}, {"--version"}, {"-v"}} {
		code, stdout, stderr := runCLITest(args...)
		if code != 0 || stdout != "spl-toolkit dev\n" || stderr != "" {
			t.Fatalf("version %v: code=%d out=%q err=%q", args, code, stdout, stderr)
		}
	}
}

func TestCLIBinaryProcessExitCodes(t *testing.T) {
	name := "spl-toolkit"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-mod=readonly", "-o", binary, ".")
	build.Dir = "."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	tests := []struct {
		name string
		args []string
		code int
	}{
		{name: "success", args: []string{"validate", "--query", "search a=1"}, code: 0},
		{name: "content rejection", args: []string{"validate", "--query", "|"}, code: 1},
		{name: "usage error", args: []string{"map", "--query", "search a=1"}, code: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := exec.Command(binary, test.args...)
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			err := command.Run()
			got := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				got = exitErr.ExitCode()
			}
			if got != test.code {
				t.Fatalf("exit=%d out=%q err=%q; want %d", got, stdout.String(), stderr.String(), test.code)
			}
		})
	}
}
