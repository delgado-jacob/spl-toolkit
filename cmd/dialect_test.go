package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/api"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func dialectHTTP(t *testing.T, method, path string, body []byte) []byte {
	t.Helper()
	r := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.NewServer().Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("HTTP %d: %s", w.Code, w.Body)
	}
	return w.Body.Bytes()
}

func dialectJSONEqual(t *testing.T, got []byte, want any) {
	t.Helper()
	expected, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var a, b any
	if err = json.Unmarshal(got, &a); err != nil {
		t.Fatalf("JSON: %v: %s", err, got)
	}
	if err = json.Unmarshal(expected, &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("full JSON differs\ngot: %s\nwant: %s", got, expected)
	}
}

func TestDialectCapabilitiesCLIHTTPParity(t *testing.T) {
	for _, language := range []string{"", "spl", "spl2"} {
		t.Run(language, func(t *testing.T) {
			want, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: language, Profile: "splunkd", Version: "current"})
			if err != nil {
				t.Fatal(err)
			}
			code, out, stderr := runCLITest("capabilities", "--language="+language, "--profile=splunkd", "--compatibility-version=current", "--format=json")
			if code != 0 || stderr != "" {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			dialectJSONEqual(t, []byte(out), want)
			dialectJSONEqual(t, dialectHTTP(t, "GET", "/capabilities?language="+language+"&profile=splunkd&version=current", nil), want)
			if language == "spl2" {
				if want.DocumentationSnapshot == "" {
					t.Fatal("missing documentation snapshot")
				}
				code, out, stderr = runCLITest("capabilities", "--language=spl2")
				if code != 0 || stderr != "" || !strings.Contains(out, "Documentation snapshot: "+want.DocumentationSnapshot) {
					t.Fatalf("text manifest: %d %s %s", code, out, stderr)
				}
			}
		})
	}
}

func TestDialectAnalysisCLIHTTPParity(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		status     analysis.Status
		diagnostic string
	}{
		{"complete", "FROM main | eval owner=host | table owner", analysis.Valid, ""},
		{"malformed", "FROM main | eval owner=", analysis.Invalid, "SPL_SYNTAX_ERROR"},
		{"deferred", "FROM main | mystery host", analysis.Incomplete, ""},
		{"module", "$q = FROM main;", analysis.Invalid, "SPL_UNSUPPORTED_MODULE"},
		{"profile", "FROM main | route output", analysis.Invalid, "SPL_PROFILE_MISMATCH"},
		{"sql", "SELECT host FROM main WHERE bytes>0", analysis.Valid, ""},
		{"unicode", "FROM main\r\n| eval label=\"café 😀\" | table label\r\n", analysis.Valid, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := analysis.QueryDocument{Text: tc.text, Language: "spl2", Profile: "splunkd", Version: "current", SourceID: " exact\r\n😀 "}
			want, err := analysis.Analyze(doc)
			if err != nil {
				t.Fatal(err)
			}
			if want.Status != tc.status {
				t.Fatalf("contract status=%s want=%s diagnostics=%+v", want.Status, tc.status, want.Diagnostics)
			}
			if tc.diagnostic != "" {
				found := false
				for _, d := range want.Diagnostics {
					found = found || d.Code == tc.diagnostic
				}
				if !found {
					t.Fatalf("missing %s: %+v", tc.diagnostic, want.Diagnostics)
				}
			}
			if tc.name == "sql" {
				phases := []string{"source", "filter", "evaluate", "project"}
				stages := []string{"stage-1", "stage-2", "stage-0", "stage-0"}
				if len(want.Lineage) != 4 || len(want.References) != 3 {
					t.Fatalf("SQL evidence: %+v", want)
				}
				for i, l := range want.Lineage {
					if l.Phase != phases[i] || l.StageID != stages[i] || l.ExecutionOrder == nil || *l.ExecutionOrder != i {
						t.Fatalf("SQL phase %d: %+v", i, l)
					}
				}
				for i, name := range []string{"host", "main", "bytes"} {
					ref := want.References[i]
					if ref.OriginalName != name || doc.Text[ref.Location.Start.Offset:ref.Location.End.Offset] != name {
						t.Fatalf("SQL reference: %+v", ref)
					}
				}
			}
			code, out, stderr := runCLITest(analysisCLIArgs(doc)...)
			if code != analysisExitCode(tc.status) || stderr != "" {
				t.Fatalf("CLI %d %s", code, stderr)
			}
			dialectJSONEqual(t, []byte(out), want)
			body, _ := json.Marshal(doc)
			dialectJSONEqual(t, dialectHTTP(t, "POST", "/query/analyze", body), want)
			output := filepath.Join(t.TempDir(), "report.json")
			code, out, stderr = runCLITest(append(analysisCLIArgs(doc), "--output", output)...)
			if code != analysisExitCode(tc.status) || out != "" || stderr != "" {
				t.Fatalf("file %d %s %s", code, out, stderr)
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			dialectJSONEqual(t, data, want)
		})
	}
}

func TestDialectLegacyCLIRejection(t *testing.T) {
	config := filepath.Join(t.TempDir(), "mapping.json")
	if err := os.WriteFile(config, []byte(`{"version":"1.0","mappings":[{"source":"host","target":"server"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"map", "discover", "validate"} {
		base := []string{command, "--query=search host=x", "--format=json"}
		if command == "map" {
			base = append(base, "--config", config)
		}
		for _, sel := range []string{"--language=spl2", "--language=unknown", "--profile=cloud", "--compatibility-version=next", "--language=\xff", "--profile=\xff", "--compatibility-version=\xff"} {
			code, out, stderr := runCLITest(append(base, sel)...)
			if code != 2 || out != "" || stderr == "" {
				t.Fatalf("%s %q: %d %s %s", command, sel, code, out, stderr)
			}
			if sel == "--language=spl2" && (!strings.Contains(stderr, "unsupported_dialect_for_operation") || !strings.Contains(stderr, "analyze")) {
				t.Fatalf("missing canonical guidance: %s", stderr)
			}
		}
		code, want, stderr := runCLITest(base...)
		if code != 0 || stderr != "" {
			t.Fatalf("legacy baseline %d %s", code, stderr)
		}
		for _, sel := range [][]string{{"--language=spl", "--profile=splunkd", "--compatibility-version=current"}, {"--language=", "--profile=", "--compatibility-version="}} {
			code, out, stderr := runCLITest(append(base, sel...)...)
			if code != 0 || out != want || stderr != "" {
				t.Fatalf("default legacy changed: %d %s %s", code, out, stderr)
			}
		}
	}
}

func TestDialectValidationCLIHTTPParity(t *testing.T) {
	dir := t.TempDir()
	fields := filepath.Join(dir, "fields.json")
	schema := filepath.Join(dir, "schema.json")
	catalog := validation.FieldCatalog{Fields: []string{"host", "bytes", "broad.flat", "broadleaf"}, OptionalFields: []string{}}
	catalogJSON, _ := json.Marshal(catalog)
	schemaJSON := []byte(`{"type":"object","properties":{"host":{"type":"string"},"bytes":{"type":"number"},"broad.flat":{"type":"string"},"broadleaf":{"type":"string"}},"additionalProperties":false}`)
	for path, data := range map[string][]byte{fields: catalogJSON, schema: schemaJSON} {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	target := validation.SchemaTarget{Kind: "json_schema", Schema: schemaJSON}
	for _, tc := range []struct {
		name, text string
		status     analysis.Status
	}{
		{"complete", "FROM main | table host", analysis.Valid},
		{"malformed", "FROM main | eval x=", analysis.Invalid},
		{"deferred", "FROM main | mystery host", analysis.Incomplete},
		{"module", "$q = FROM main;", analysis.Invalid},
		{"profile", "FROM main | route output", analysis.Invalid},
		{"sql", "SELECT host FROM main WHERE bytes>0", analysis.Valid},
		{"null-test", "FROM main | where isnull(absent)", analysis.Valid},
		{"null-read", "FROM main | where isnull(absent) OR absent=1", analysis.Invalid},
		{"dotted-wildcard", "FROM main | eval broadlocal=1 | fields 'broad*'", analysis.Incomplete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := analysis.QueryDocument{Text: tc.text + "\r\n", Language: "spl2", Profile: "splunkd", Version: "current", SourceID: " exact\r\n😀 "}
			for _, command := range []string{"validate-fields", "validate-schema"} {
				status := tc.status
				if tc.name == "dotted-wildcard" && command == "validate-fields" {
					status = analysis.Valid
				}
				args := []string{command, "--format=json", "--language=spl2", "--profile=splunkd", "--compatibility-version=current", "--source-id", doc.SourceID}
				var want any
				if command == "validate-fields" {
					args = append(args, "--fields", fields)
					r, e := validation.Validate(doc, catalog)
					if e != nil {
						t.Fatal(e)
					}
					if r.Status != status {
						t.Fatalf("%s contract: %s want %s", command, r.Status, status)
					}
					want = r
					if tc.name == "null-test" {
						inspections := 0
						for _, ref := range r.Analysis.References {
							if ref.Role == "null_test" {
								inspections++
								if ref.OriginalName != "absent" || ref.Binding != "source" {
									t.Fatalf("null inspection evidence: %+v", ref)
								}
								for _, o := range r.Outcomes {
									if o.ReferenceID == ref.ID {
										t.Fatal("null_test acquired required-existence outcome")
									}
								}
							}
						}
						if inspections != 1 {
							t.Fatalf("null inspection count: %d", inspections)
						}
					}
				} else {
					args = append(args, "--schema", schema)
					r, e := validation.ValidateSchema(doc, target)
					if e != nil {
						t.Fatal(e)
					}
					if r.Status != status {
						t.Fatalf("%s contract: %s want %s", command, r.Status, status)
					}
					want = r
					if tc.name == "dotted-wildcard" {
						outcome := r.Outcomes[len(r.Outcomes)-1]
						names := []string{}
						for _, match := range outcome.Matches {
							names = append(names, match.Name)
						}
						if outcome.Outcome != "indeterminate" || outcome.MatchesComplete || !reflect.DeepEqual(names, []string{"broadleaf", "broadlocal"}) {
							t.Fatalf("schema wildcard evidence: %+v", outcome)
						}
						for _, field := range r.Analysis.Lineage[len(r.Analysis.Lineage)-1].After.Fields {
							if field.Name == "broad.flat" {
								t.Fatal("ambiguous dotted candidate materialized")
							}
						}
					}
				}
				queryfile := filepath.Join(dir, "query.spl2")
				if err := os.WriteFile(queryfile, []byte(doc.Text), 0600); err != nil {
					t.Fatal(err)
				}
				for _, source := range [][]string{{"--query", doc.Text}, {"--file", queryfile}, {"--stdin"}} {
					var out, stderr bytes.Buffer
					code := runCLIWithInput(append(args, source...), strings.NewReader(doc.Text), &out, &stderr)
					if code != analysisExitCode(status) || stderr.Len() != 0 {
						t.Fatalf("%s %v: %d %s", command, source, code, &stderr)
					}
					dialectJSONEqual(t, out.Bytes(), want)
				}

				output := filepath.Join(dir, "report.json")
				code, out, stderr := runCLITest(append(args, "--query", doc.Text, "--output", output)...)
				if code != analysisExitCode(status) || out != "" || stderr != "" {
					t.Fatalf("validation output file: %d %s %s", code, out, stderr)
				}
				raw, err := os.ReadFile(output)
				if err != nil {
					t.Fatal(err)
				}
				dialectJSONEqual(t, raw, want)
				body := map[string]any{"document": doc}
				if command == "validate-fields" {
					body["catalog"] = catalog
				} else {
					body["target"] = target
				}
				data, _ := json.Marshal(body)
				dialectJSONEqual(t, dialectHTTP(t, "POST", "/query/"+command, data), want)
			}
		})
	}
	docs := []analysis.QueryDocument{{Text: "table host", SourceID: "legacy"}, {Text: "FROM main | table host", Language: "spl2", SourceID: "spl2"}, {Text: "FROM main | mystery host", Language: "spl2", SourceID: "deferred"}}
	batch, _ := json.Marshal(docs)
	for _, command := range []string{"validate-fields", "validate-schema"} {
		args := []string{command, "--format=json", "--batch=-"}
		body := map[string]any{"documents": docs}
		var want any
		if command == "validate-fields" {
			args = append(args, "--fields", fields)
			body["catalog"] = catalog
			want, _ = validation.ValidateBatch(docs, catalog)
		} else {
			args = append(args, "--schema", schema)
			body["target"] = target
			want, _ = validation.ValidateSchemaBatch(docs, target)
		}
		var out, stderr bytes.Buffer
		code := runCLIWithInput(args, bytes.NewReader(batch), &out, &stderr)
		if code != 3 || stderr.Len() != 0 {
			t.Fatalf("batch %d %s", code, &stderr)
		}
		dialectJSONEqual(t, out.Bytes(), want)
		data, _ := json.Marshal(body)
		dialectJSONEqual(t, dialectHTTP(t, "POST", "/query/"+command+"/batch", data), want)
		for _, flag := range []string{"--language=spl2", "--profile=", "--compatibility-version=current", "--source-id="} {
			out.Reset()
			stderr.Reset()
			code = runCLIWithInput(append(args, flag), bytes.NewReader(batch), &out, &stderr)
			if code != 2 || out.Len() != 0 || !strings.Contains(stderr.String(), "batch documents") {
				t.Fatalf("batch global selector %q: %d %s %s", flag, code, &out, &stderr)
			}
		}
		single := args[:2]
		if command == "validate-fields" {
			single = append(single, "--fields", fields)
		} else {
			single = append(single, "--schema", schema)
		}
		single = append(single, "--query=FROM main | table host")
		for _, flag := range []string{"--language=unknown", "--profile=cloud", "--compatibility-version=next", "--language=\xff", "--profile=\xff", "--compatibility-version=\xff", "--source-id=\xff"} {
			code, out, stderr := runCLITest(append(single, flag)...)
			if code != 2 || out != "" || stderr == "" {
				t.Fatalf("selector %s: %d %s %s", flag, code, out, stderr)
			}
		}
	}
}

func TestDialectCLIHelpAndSelectorErrors(t *testing.T) {
	code, out, stderr := runCLITest("help")
	if code != 0 || stderr != "" {
		t.Fatalf("help %d %s", code, stderr)
	}
	for _, fragment := range []string{"spl or spl2 (default: spl)", "capabilities accept language/profile/version selectors", "Legacy map/discover/validate reject spl2"} {
		if !strings.Contains(out, fragment) {
			t.Errorf("help missing %q", fragment)
		}
	}
	for _, command := range []string{"analyze", "capabilities"} {
		args := []string{command, "--format=json"}
		if command == "analyze" {
			args = append(args, "--query=FROM main | table host")
		}
		for _, flags := range [][]string{{"--language=unknown"}, {"--profile=cloud"}, {"--compatibility-version=next"}, {"--language=\xff"}, {"--profile=\xff"}, {"--compatibility-version=\xff"}, {"--language=spl2", "--language=spl"}, {"--profile=splunkd", "--profile=splunkd"}, {"--compatibility-version=current", "--compatibility-version=current"}} {
			code, out, stderr := runCLITest(append(args, flags...)...)
			if code != 2 || out != "" || stderr == "" {
				t.Fatalf("%s %v: %d %s %s", command, flags, code, out, stderr)
			}
		}
	}
}
