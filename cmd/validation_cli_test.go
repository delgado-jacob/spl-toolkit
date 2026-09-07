package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func validationFile(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateFieldsStdinIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fields.json")
	if err := os.WriteFile(path, []byte("[\"host\"]"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := runCLIWithInput([]string{"validate-fields", "--fields", path, "--stdin", "--format", "json"}, strings.NewReader("search host=web\n"), &out, &errOut)
	var report validation.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if code != 0 || errOut.Len() != 0 || report.Analysis.Document.SourceID != "<stdin>" || report.Analysis.Document.Text != "search host=web\n" {
		t.Fatalf("code=%d report=%+v stderr=%s", code, report, errOut.String())
	}
}

func TestValidateFieldsSourcesAndCanonicalReports(t *testing.T) {
	fields := validationFile(t, `{"fields":["host"],"optional_fields":["user"],"identity":"catalog","version":"v1"}`)
	for _, query := range []string{"search host=web\r\n", "search missing=x", "| mystery", "", "search host=\"é😀\"\n"} {
		for _, mode := range []string{"positional", "query", "file", "stdin", "override"} {
			t.Run(mode+query, func(t *testing.T) {
				args := []string{"validate-fields", "--fields", fields, "--format=json"}
				source := ""
				switch mode {
				case "positional":
					args = append(args, "--", query)
				case "query":
					args = append(args, "--query", query)
				case "file":
					source = validationFile(t, query)
					args = append(args, "--file", source)
				case "stdin":
					source = "<stdin>"
					args = append(args, "--stdin")
				case "override":
					source = "mine"
					args = append(args, "--stdin", "--source-id", source, "--language=spl", "--profile=splunkd", "--compatibility-version=current")
				}
				var out, stderr bytes.Buffer
				code := runCLIWithInput(args, strings.NewReader(query), &out, &stderr)
				var got validation.Report
				if err := json.Unmarshal(out.Bytes(), &got); err != nil {
					t.Fatalf("%v stderr=%s", err, &stderr)
				}
				want, err := validation.Validate(analysis.QueryDocument{Text: query, SourceID: source}, validation.FieldCatalog{Fields: []string{"host"}, OptionalFields: []string{"user"}, Identity: "catalog", Version: "v1"})
				if err != nil {
					t.Fatal(err)
				}
				wantCode := 0
				if want.Status == analysis.Invalid {
					wantCode = 1
				}
				if want.Status == analysis.Incomplete {
					wantCode = 3
				}
				if code != wantCode || stderr.Len() != 0 || !reflect.DeepEqual(&got, want) {
					t.Fatalf("code=%d want=%d report parity=%t stderr=%s", code, wantCode, reflect.DeepEqual(&got, want), &stderr)
				}
			})
		}
	}
}

func TestValidateFieldsBatch(t *testing.T) {
	fields := validationFile(t, `["host"]`)
	for _, batch := range []string{`[{"text":"search host=x","source_id":"first"},{"text":"| mystery","source_id":"second"}]`, `[{"text":"| mystery"},{"text":"search missing=x"},{"text":"search host=x\n","source_id":"last"}]`} {
		for _, mode := range []string{"file", "stdin"} {
			path := "-"
			if mode == "file" {
				path = validationFile(t, batch)
			}
			var out, stderr bytes.Buffer
			code := runCLIWithInput([]string{"validate-fields", "--fields", fields, "--batch", path, "--format=json"}, strings.NewReader(batch), &out, &stderr)
			var got validation.BatchReport
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatalf("%v stderr=%s", err, &stderr)
			}
			docs, err := validation.DecodeDocuments([]byte(batch))
			if err != nil {
				t.Fatal(err)
			}
			want, err := validation.ValidateBatch(docs, validation.FieldCatalog{Fields: []string{"host"}})
			if err != nil {
				t.Fatal(err)
			}
			wantCode := 3
			if want.Status == analysis.Invalid {
				wantCode = 1
			}
			if code != wantCode || stderr.Len() != 0 || !reflect.DeepEqual(&got, want) {
				t.Fatalf("code=%d report=%+v stderr=%s", code, got, &stderr)
			}
		}
	}
}

func TestValidateFieldsRejectsInputs(t *testing.T) {
	fields := validationFile(t, `["host"]`)
	base := []string{"validate-fields", "--fields", fields}
	cases := [][]string{
		{}, {"--stdin", "--query=x"}, {"--stdin", "--stdin"}, {"--file=x", "--file=x"}, {"--batch=-", "--batch=-"}, {"--file=x", "x"}, {"--batch=-", "--stdin"}, {"--fields", fields, "x"},
		{"--batch=-", "--language=spl"}, {"--batch=-", "--profile="}, {"--batch=-", "--compatibility-version=current"}, {"--batch=-", "--source-id="},
		{"--language=spl2", "x"}, {"--profile=cloud", "x"}, {"--compatibility-version=9", "x"}, {"--source-id", string([]byte{255}), "x"},
		{"--query", string([]byte{255})}, {"--file=/nonexistent/spl-query"}, {"--batch=/nonexistent/spl-batch"}, {"--config=x", "x"}, {"--format=yaml", "x"}, {"--wat", "x"}, {"--stdin=true"},
	}
	for _, args := range cases {
		var out, stderr bytes.Buffer
		code := runCLIWithInput(append(append([]string{}, base...), args...), strings.NewReader(`[]`), &out, &stderr)
		if code != 2 || out.Len() != 0 || stderr.Len() == 0 {
			t.Errorf("args=%q code=%d out=%s err=%s", args, code, &out, &stderr)
		}
	}
	for _, args := range [][]string{{"validate-fields", "x"}, {"validate-fields", "--fields=-", "x"}, {"validate-fields", "--fields=/nonexistent/fields", "x"}} {
		var out, stderr bytes.Buffer
		if code := runCLIWithInput(args, strings.NewReader(""), &out, &stderr); code != 2 {
			t.Errorf("args=%v code=%d", args, code)
		}
	}
	for _, bad := range []string{`null`, `[]`, `[{"text":null}]`, `[{"text":"x","text":"y"}]`, `[{"text":"\ud800"}]`} {
		var out, stderr bytes.Buffer
		if code := runCLIWithInput(append(base, "--batch=-"), strings.NewReader(bad), &out, &stderr); code != 2 || out.Len() != 0 {
			t.Errorf("batch=%s code=%d", bad, code)
		}
	}
	for _, bad := range []string{`null`, `["host","host"]`, `{"fields":[],"fields":[]}`, `["\ud800"]`} {
		var out, stderr bytes.Buffer
		if code := runCLIWithInput([]string{"validate-fields", "--fields", validationFile(t, bad), "x"}, strings.NewReader(""), &out, &stderr); code != 2 || out.Len() != 0 {
			t.Errorf("catalog=%s code=%d", bad, code)
		}
	}
}

type validationBrokenIO struct{}

func (validationBrokenIO) Read([]byte) (int, error)  { return 0, errors.New("read failed") }
func (validationBrokenIO) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestValidateFieldsOutputAndIO(t *testing.T) {
	fields := validationFile(t, `["host"]`)
	var out, stderr bytes.Buffer
	path := filepath.Join(t.TempDir(), "report.json")
	args := []string{"validate-fields", "--fields", fields, "--query=search missing=x", "--format=json", "--output", path}
	if code := runCLIWithInput(args, strings.NewReader(""), &out, &stderr); code != 1 || out.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, &out, &stderr)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report validation.Report
	if err = json.Unmarshal(data, &report); err != nil || report.Status != analysis.Invalid {
		t.Fatalf("report=%s err=%v", data, err)
	}
	for _, input := range []string{"--stdin", "--batch=-"} {
		if code := runCLIWithInput([]string{"validate-fields", "--fields", fields, input}, validationBrokenIO{}, &out, &stderr); code != 2 {
			t.Errorf("read code=%d", code)
		}
	}
	if code := runCLIWithInput(args[:len(args)-2], strings.NewReader(""), validationBrokenIO{}, &stderr); code != 2 {
		t.Errorf("write code=%d", code)
	}
	args[len(args)-1] = t.TempDir()
	if code := runCLIWithInput(args, strings.NewReader(""), &out, &stderr); code != 2 {
		t.Errorf("output code=%d", code)
	}
	out.Reset()
	stderr.Reset()
	code := runCLIWithInput([]string{"validate-fields", "--fields", fields, "--source-id=sample", "search host=x missing=y"}, strings.NewReader(""), &out, &stderr)
	for _, fragment := range []string{"Status: invalid", "sample", "Schema coverage:", "matching", "missing", "host", "bytes", "Diagnostics:"} {
		if !strings.Contains(out.String(), fragment) {
			t.Errorf("missing %q in %s", fragment, &out)
		}
	}
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%s", code, &stderr)
	}
}

func TestValidateFieldsBatchText(t *testing.T) {
	fields := validationFile(t, `["host"]`)
	var out, stderr bytes.Buffer
	code := runCLIWithInput([]string{"validate-fields", "--fields", fields, "--batch=-"}, strings.NewReader(`[{"text":"search host=x","source_id":"first"},{"text":"| mystery","source_id":"second"}]`), &out, &stderr)
	if code != 3 || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%s", code, &stderr)
	}
	for _, fragment := range []string{"Batch status: incomplete", "Document 1:", "Source: \"first\"", "Document 2:", "Source: \"second\"", "matching", "Diagnostics:"} {
		if !strings.Contains(out.String(), fragment) {
			t.Errorf("missing %q in %s", fragment, &out)
		}
	}
	if strings.Index(out.String(), "first") > strings.Index(out.String(), "second") {
		t.Fatal("batch text reordered documents")
	}
}
