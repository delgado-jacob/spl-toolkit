package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaCLIReportsAndExitCodes(t *testing.T) {
	for _, tc := range []struct {
		schema, query, outcome string
		code                   int
	}{
		{`{"properties":{"host":{}},"additionalProperties":false}`, "table absent", "missing", 1},
		{`{}`, "table absent", "permitted_unspecified", 0},
		{`{"$ref":"https://offline.invalid/schema"}`, "table host", "indeterminate", 3},
		{`{}`, "table h*", "indeterminate", 3},
	} {
		t.Run(tc.query+tc.schema, func(t *testing.T) {
			path := validationFile(t, tc.schema)
			var out, stderr bytes.Buffer
			code := runCLIWithInput([]string{"validate-schema", "--schema", path, "--query", tc.query, "--format", "json"}, strings.NewReader(""), &out, &stderr)
			if code != tc.code || out.Len() == 0 || stderr.Len() != 0 {
				t.Fatalf("code=%d out=%s err=%s", code, &out, &stderr)
			}
			var got validation.SchemaReport
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			want, err := validation.ValidateSchema(analysis.QueryDocument{Text: tc.query}, validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(tc.schema)})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&got, want) {
				t.Fatalf("parity mismatch: %s", &out)
			}
			if len(got.Outcomes) == 0 || got.Outcomes[0].Outcome != tc.outcome {
				t.Fatalf("outcomes=%+v", got.Outcomes)
			}
		})
	}
}

func TestSchemaCLISourcesBatchAndOutput(t *testing.T) {
	schema := validationFile(t, `{}`)
	query := "search host=\"café\"\r\n| table host\r\n"
	path := validationFile(t, query)
	for _, source := range []string{"file", "stdin"} {
		for _, override := range []bool{false, true} {
			args := []string{"validate-schema", "--schema", schema, "--format=json"}
			id := "<stdin>"
			if source == "file" {
				args = append(args, "--file", path)
				id = path
			} else {
				args = append(args, "--stdin")
			}
			if override {
				args = append(args, "--source-id=")
				id = ""
			}
			var out, stderr bytes.Buffer
			code := runCLIWithInput(args, strings.NewReader(query), &out, &stderr)
			if code != 0 {
				t.Fatalf("%d %s", code, &stderr)
			}
			var got validation.SchemaReport
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Analysis.Document.Text != query || got.Analysis.Document.SourceID != id {
				t.Fatalf("document=%+v", got.Analysis.Document)
			}
		}
	}
	batch := `[{"text":"table host","source_id":"first"},{"text":"table h*","source_id":"second"}]`
	for _, path := range []string{"-", validationFile(t, batch)} {
		output := filepath.Join(t.TempDir(), "report.json")
		var out, stderr bytes.Buffer
		code := runCLIWithInput([]string{"validate-schema", "--schema", schema, "--batch", path, "--format=json", "--output", output}, strings.NewReader(batch), &out, &stderr)
		if code != 3 || out.Len() != 0 || stderr.Len() != 0 {
			t.Fatalf("%d %s %s", code, &out, &stderr)
		}
		raw, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var got validation.SchemaBatchReport
		if err = json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		if len(got.Reports) != 2 || got.Reports[0].Analysis.Document.SourceID != "first" || got.Reports[1].Analysis.Document.SourceID != "second" {
			t.Fatalf("%s", raw)
		}
	}
}

func TestSchemaCLIRejectsBadOptionsAndInputs(t *testing.T) {
	schema := validationFile(t, `{}`)
	for _, tail := range [][]string{
		{"--schema", schema, "--schema", schema, "x"}, {"--schema=-", "x"}, {"--schema=/nonexistent", "x"}, {"--schema", validationFile(t, `null`), "x"},
		{"--schema", schema, "x", "--stdin"}, {"--schema", schema, "--batch=-", "--profile="}, {"--schema", schema, "--ocsf-version=1.6.0", "x"},
		{"--schema", schema, "--ocsf-profile=p", "x"}, {"--schema", schema, "--ocsf-catalog=c", "x"}, {"--schema", schema, "--schema-resources=-", "x"},
		{"--schema", schema, "--bogus=x", "x"}, {"--schema", schema, "--config=c", "x"}, {"--schema", schema}, {"--schema=", "x"}, {"--query=x"},
		{"--ocsf-catalog=c", "--schema-base-uri=urn:x", "x"}, {"--ocsf-catalog=c", "--ocsf-version=1.6.0", "--ocsf-class=1", "--ocsf-category=2", "x"},
	} {
		var out, stderr bytes.Buffer
		code := runCLIWithInput(append([]string{"validate-schema"}, tail...), strings.NewReader("[]"), &out, &stderr)
		if code != 2 || out.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("%v: %d %s %s", tail, code, &out, &stderr)
		}
	}
	for _, batch := range []string{`[]`, `null`, `[{"text":"x","unknown":true}]`, `[`, "\xff"} {
		var out, stderr bytes.Buffer
		code := runCLIWithInput([]string{"validate-schema", "--schema", schema, "--batch=-"}, strings.NewReader(batch), &out, &stderr)
		if code != 2 || out.Len() != 0 {
			t.Fatalf("%q: %d %s", batch, code, &out)
		}
	}
}

func TestSchemaCLIResourcesAndText(t *testing.T) {
	schema := validationFile(t, `{"$ref":"urn:local"}`)
	resources := validationFile(t, `{"urn:local":{"properties":{"host":{}},"additionalProperties":false}}`)
	var out, stderr bytes.Buffer
	code := runCLIWithInput([]string{"validate-schema", "--schema", schema, "--schema-resources", resources, "--schema-base-uri=urn:root", "table host"}, strings.NewReader(""), &out, &stderr)
	if code != 0 {
		t.Fatalf("%d %s", code, &stderr)
	}
	for _, s := range []string{"Target:", "Schema coverage:", "host", "@", "resource_uri", "urn:local"} {
		if !strings.Contains(out.String(), s) {
			t.Fatalf("missing %q: %s", s, &out)
		}
	}
}

func TestSchemaCLIRejectsMalformedUnicodeWithoutReplacement(t *testing.T) {
	schema := validationFile(t, `{}`)
	for _, key := range []string{"schema-base-uri", "ocsf-version", "ocsf-class", "ocsf-profile", "ocsf-extension"} {
		options, _, err := parseCLIOptions("validate-schema", []string{"--schema", schema, "--" + key + "=\xff", "x"})
		if err == nil {
			_, err = readSchemaCLITarget(options)
		}
		if err == nil {
			t.Errorf("invalid UTF-8 accepted for %s", key)
		}
	}
	for _, raw := range []string{"{\"title\":\"\xff\"}", `{"title":"\ud800"}`, `{"title":"a","title":"b"}`} {
		var out, stderr bytes.Buffer
		code := runCLIWithInput([]string{"validate-schema", "--schema", validationFile(t, raw), "table host"}, strings.NewReader(""), &out, &stderr)
		if code != 2 || out.Len() != 0 {
			t.Errorf("malformed target accepted: %q %d %s", raw, code, &out)
		}
	}
}

func TestSchemaCLIOCSFSelectors(t *testing.T) {
	f, err := os.Open("../testdata/schemas/ocsf/1.6.0/base.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137" {
		t.Fatal("fixture hash mismatch")
	}
	catalog := validationFile(t, string(raw))
	uid := int64(3002)
	for _, language := range []string{"spl", "spl2"} {
		query := "table time"
		if language == "spl2" {
			query = "FROM main SELECT time"
		}
		for _, selector := range []string{"authentication", "003002"} {
			args := []string{"validate-schema", "--ocsf-catalog", catalog, "--ocsf-version=1.6.0", "--ocsf-class", selector, "--ocsf-profile=host", "--ocsf-profile=datetime", "--profile=splunkd", "--compatibility-version=current", "--format=json", "--language=" + language, "--source-id=ocsf-query", query}
			var out, stderr bytes.Buffer
			code := runCLIWithInput(args, strings.NewReader(""), &out, &stderr)
			if code != 0 || stderr.Len() != 0 {
				t.Fatalf("%s %s %d %s %s", language, selector, code, &out, &stderr)
			}
			selection := &validation.OCSFSelection{Version: "1.6.0", Class: "authentication", Profiles: []string{"host", "datetime"}, Extensions: []string{}}
			if selector != "authentication" {
				selection.Class = ""
				selection.ClassUID = &uid
			}
			want, err := validation.ValidateSchema(analysis.QueryDocument{Text: query, Language: language, Profile: "splunkd", Version: "current", SourceID: "ocsf-query"}, validation.SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: selection})
			if err != nil {
				t.Fatal(err)
			}
			var got validation.SchemaReport
			if err = json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&got, want) {
				t.Fatalf("parity: %s", &out)
			}
		}
	}
	// Target construction preserves category keys versus numeric UIDs and repeated extension arguments.
	for _, selector := range []string{"iam", "3"} {
		options, _, err := parseCLIOptions("validate-schema", []string{"--ocsf-catalog", catalog, "--ocsf-version=1.6.0", "--ocsf-category", selector, "--ocsf-extension=z", "--ocsf-extension=a", "x"})
		if err != nil {
			t.Fatal(err)
		}
		target, err := readSchemaCLITarget(options)
		if err != nil {
			t.Fatal(err)
		}
		if selector == "iam" && target.Selection.Category != "iam" {
			t.Fatal("category key lost")
		}
		if selector == "3" && (target.Selection.CategoryUID == nil || *target.Selection.CategoryUID != 3) {
			t.Fatal("category UID lost")
		}
		if !reflect.DeepEqual(target.Selection.Extensions, []string{"a", "z"}) {
			t.Fatal("extensions lost")
		}
	}
	for _, tail := range [][]string{
		{"--ocsf-class=authentication", "--ocsf-class=authentication"},
		{"--ocsf-class=authentication", "--ocsf-profile=host", "--ocsf-profile=host"},
		{"--ocsf-class=authentication", "--ocsf-extension=win", "--ocsf-extension=win"},
		{"--ocsf-class=9223372036854775808"}, {"--ocsf-class=authentication", "--ocsf-category=3"}, {"--ocsf-class=authentication", "--ocsf-version=1.6.0"}, {},
	} {
		args := append([]string{"validate-schema", "--ocsf-catalog", catalog, "--ocsf-version=1.6.0", "--query=table time"}, tail...)
		var out, stderr bytes.Buffer
		if code := runCLIWithInput(args, strings.NewReader(""), &out, &stderr); code != 2 || out.Len() != 0 {
			t.Fatalf("%v %d %s", tail, code, &out)
		}
	}
}

func TestSchemaCLIHelpAndIOErrors(t *testing.T) {
	var out, stderr bytes.Buffer
	if code := runCLIWithInput([]string{"validate-schema", "--help"}, strings.NewReader(""), &out, &stderr); code != 0 || !strings.Contains(out.String(), "--ocsf-catalog") {
		t.Fatalf("help: %d %s %s", code, &out, &stderr)
	}
	schema := validationFile(t, `{}`)
	for _, input := range []string{"--stdin", "--batch=-"} {
		out.Reset()
		stderr.Reset()
		if code := runCLIWithInput([]string{"validate-schema", "--schema", schema, input}, validationBrokenIO{}, &out, &stderr); code != 2 || out.Len() != 0 {
			t.Fatalf("read error %d %s", code, &out)
		}
	}
	out.Reset()
	stderr.Reset()
	if code := runCLIWithInput([]string{"validate-schema", "--schema", schema, "--query=table host", "--output", filepath.Join(t.TempDir(), "missing", "out")}, strings.NewReader(""), &out, &stderr); code != 2 || out.Len() != 0 {
		t.Fatalf("write error %d %s", code, &out)
	}
}
