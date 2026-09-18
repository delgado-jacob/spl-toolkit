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
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func rewriteFile(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func rewriteRulesFile(t *testing.T) string {
	t.Helper()
	return rewriteFile(t, "rules.json", `{"schema_version":1,"rules":[{"id":"rename","kind":"field","source":{"name":"src"},"target":{"name":"user"}}]}`)
}

func TestRewriteCLICanonicalSingleSourcesAndModes(t *testing.T) {
	rulesPath := rewriteRulesFile(t)
	query := "search src=alice\r\n| table src\n"
	for _, source := range []string{"positional", "query", "file", "stdin", "override"} {
		for _, apply := range []bool{false, true} {
			t.Run(source+map[bool]string{false: "-preview", true: "-apply"}[apply], func(t *testing.T) {
				args := []string{"rewrite", "--rules", rulesPath, "--format=json"}
				mode := rewrite.Preview
				if apply {
					args = append(args, "--apply")
					mode = rewrite.Apply
				}
				sourceID := ""
				switch source {
				case "positional":
					args = append(args, "--", query)
				case "query":
					args = append(args, "--query", query)
				case "file":
					path := rewriteFile(t, "input.spl", query)
					sourceID = path
					args = append(args, "--file", path)
				case "stdin":
					sourceID = "<stdin>"
					args = append(args, "--stdin")
				case "override":
					sourceID = " exact id "
					args = append(args, "--stdin", "--source-id", sourceID)
				}
				var stdout, stderr bytes.Buffer
				code := runCLIWithInput(args, strings.NewReader(query), &stdout, &stderr)
				var got rewrite.Result
				if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
					t.Fatalf("decode stdout: %v; stderr=%s", err, &stderr)
				}
				rules, err := rewrite.DecodeRuleSet([]byte(`{"schema_version":1,"rules":[{"id":"rename","kind":"field","source":{"name":"src"},"target":{"name":"user"}}]}`))
				if err != nil {
					t.Fatal(err)
				}
				want, err := rewrite.Rewrite(rewrite.Request{SchemaVersion: 1, Mode: mode, Document: analysis.QueryDocument{Text: query, SourceID: sourceID}, Rules: rules.Rules})
				if err != nil {
					t.Fatal(err)
				}
				if code != analysisStatusExitCode(want.Status) || stderr.Len() != 0 || !reflect.DeepEqual(&got, want) {
					t.Fatalf("code=%d want=%d parity=%t stderr=%s", code, analysisStatusExitCode(want.Status), reflect.DeepEqual(&got, want), &stderr)
				}
			})
		}
	}
}

func TestRewriteCLIBatchTargetOutputAndIncompleteCommit(t *testing.T) {
	rules := rewriteRulesFile(t)
	batch := `[{"text":"search src=x","source_id":"first"},{"text":"| mystery src","source_id":"second"}]`
	fields := rewriteFile(t, "fields.json", `["user"]`)
	output := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer
	code := runCLIWithInput([]string{"rewrite", "--rules", rules, "--fields", fields, "--batch=-", "--apply", "--format=json", "--output", output}, strings.NewReader(batch), &stdout, &stderr)
	if code != 3 || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	docs, err := validation.DecodeDocuments([]byte(batch))
	if err != nil {
		t.Fatal(err)
	}
	ruleSet, err := rewrite.DecodeRuleSet([]byte(`{"schema_version":1,"rules":[{"id":"rename","kind":"field","source":{"name":"src"},"target":{"name":"user"}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := validation.DecodeFieldCatalog([]byte(`["user"]`))
	if err != nil {
		t.Fatal(err)
	}
	want, err := rewrite.RewriteBatch(rewrite.BatchRequest{SchemaVersion: 1, Mode: rewrite.Apply, Documents: docs, Rules: ruleSet.Rules, ValidationTarget: &rewrite.ValidationTarget{Kind: "field_list", Catalog: &catalog}})
	if err != nil {
		t.Fatal(err)
	}
	expected, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSpace(data), expected) || !want.Reports[0].Committed {
		t.Fatalf("parity=%t committed=%v", bytes.Equal(bytes.TrimSpace(data), expected), want.Reports[0].Committed)
	}
}

func TestRewriteCLISingleTransportsIncompleteCommittedPartialRewrite(t *testing.T) {
	rules := rewriteFile(t, "rules.json", `{"schema_version":1,"rules":[{"id":"src-user","kind":"field","source":{"name":"src"},"target":{"name":"user"}},{"id":"other-a","kind":"field","source":{"name":"other"},"target":{"name":"a"}},{"id":"other-b","kind":"field","source":{"name":"other"},"target":{"name":"b"}}]}`)
	query := "search src=x other=y | table src other"
	var stdout, stderr bytes.Buffer
	code := runCLIWithInput([]string{"rewrite", "--rules", rules, "--query", query, "--apply", "--format=json"}, strings.NewReader(""), &stdout, &stderr)
	var got rewrite.Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout: %v; stderr=%s", err, &stderr)
	}
	if code != 3 || stderr.Len() != 0 || got.Status != analysis.Incomplete || !got.Committed || got.Text != "search user=x other=y | table user other" || got.Text == got.OriginalText {
		t.Fatalf("code=%d report=%+v stderr=%s", code, got, &stderr)
	}
	applied, ambiguous := 0, 0
	for _, change := range got.Changes {
		switch change.Outcome {
		case "applied":
			applied++
			if !change.CandidateApplied || !change.Committed {
				t.Errorf("applied change lost commit evidence: %+v", change)
			}
		case "ambiguous":
			ambiguous++
			if change.CandidateApplied || change.Committed {
				t.Errorf("ambiguous change was transported as applied: %+v", change)
			}
		}
	}
	if applied != 2 || ambiguous != 4 {
		t.Fatalf("partial audit lost: applied=%d ambiguous=%d", applied, ambiguous)
	}
}

func TestRewriteCLIRejectsInvalidCombinations(t *testing.T) {
	rules := rewriteRulesFile(t)
	fields := rewriteFile(t, "fields.json", `["user"]`)
	schema := rewriteFile(t, "schema.json", `{}`)
	base := []string{"rewrite", "--rules", rules}
	cases := [][]string{
		{}, {"--query=x", "y"}, {"--stdin", "--file=x"}, {"--batch=-", "--query=x"},
		{"--batch=-", "--language=spl"}, {"--batch=-", "--profile="}, {"--batch=-", "--compatibility-version=current"}, {"--batch=-", "--source-id="},
		{"--config=x", "x"}, {"--rules=-", "x"}, {"--rules", rules, "x"}, {"--apply=true", "x"}, {"--apply", "--apply", "x"},
		{"--fields", fields, "--schema", schema, "x"}, {"--schema-base-uri=urn:x", "x"}, {"--ocsf-version=1.6.0", "x"},
	}
	for _, suffix := range cases {
		var stdout, stderr bytes.Buffer
		code := runCLIWithInput(append(append([]string{}, base...), suffix...), strings.NewReader(`[]`), &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Errorf("args=%q code=%d stdout=%s stderr=%s", suffix, code, &stdout, &stderr)
		}
	}
	for _, bad := range []string{`null`, `{"schema_version":1,"rules":[],"mode":"apply"}`, `{"schema_version":1,"rules":null}`} {
		var stdout, stderr bytes.Buffer
		code := runCLIWithInput([]string{"rewrite", "--rules", rewriteFile(t, "bad.json", bad), "x"}, strings.NewReader(""), &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 {
			t.Errorf("rules=%s code=%d", bad, code)
		}
	}
}

type rewriteBrokenIO struct{}

func (rewriteBrokenIO) Read([]byte) (int, error)  { return 0, errors.New("read failed") }
func (rewriteBrokenIO) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestRewriteCLITextAndIOFailures(t *testing.T) {
	rules := rewriteRulesFile(t)
	var stdout, stderr bytes.Buffer
	if code := runCLIWithInput([]string{"rewrite", "--rules", rules, "--query=search src=x"}, strings.NewReader(""), &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "Candidate text: search user=x") || !strings.Contains(stdout.String(), "Returned text: search src=x") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runCLIWithInput([]string{"rewrite", "--rules", rules, "--stdin"}, rewriteBrokenIO{}, &stdout, &stderr); code != 2 || stdout.Len() != 0 {
		t.Fatalf("read code=%d", code)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runCLIWithInput([]string{"rewrite", "--rules", rules, "--query=search src=x"}, strings.NewReader(""), rewriteBrokenIO{}, &stderr); code != 2 || !strings.Contains(stderr.String(), "write failed") {
		t.Fatalf("write code=%d stderr=%s", code, &stderr)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runCLIWithInput([]string{"rewrite", "--rules", rules, "--query=search src=x", "--output", t.TempDir()}, strings.NewReader(""), &stdout, &stderr); code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "write output") {
		t.Fatalf("output code=%d stderr=%s", code, &stderr)
	}
}

func TestRewriteCLINoopAndFailedApplyAudit(t *testing.T) {
	noRules := rewriteFile(t, "rules.json", `{"schema_version":1,"rules":[]}`)
	var stdout, stderr bytes.Buffer
	code := runCLIWithInput([]string{"rewrite", "--rules", noRules, "--query=search host=x", "--format=json"}, strings.NewReader(""), &stdout, &stderr)
	var noOp rewrite.Result
	if err := json.Unmarshal(stdout.Bytes(), &noOp); err != nil {
		t.Fatal(err)
	}
	if code != 0 || stderr.Len() != 0 || len(noOp.Changes) != 0 || noOp.CandidateText != noOp.OriginalText || noOp.Text != noOp.OriginalText {
		t.Fatalf("code=%d report=%+v stderr=%s", code, noOp, &stderr)
	}

	stdout.Reset()
	stderr.Reset()
	rules := rewriteRulesFile(t)
	fields := rewriteFile(t, "fields.json", `["src"]`)
	code = runCLIWithInput([]string{"rewrite", "--rules", rules, "--fields", fields, "--query=search src=x", "--apply", "--format=json"}, strings.NewReader(""), &stdout, &stderr)
	var failed rewrite.Result
	if err := json.Unmarshal(stdout.Bytes(), &failed); err != nil {
		t.Fatal(err)
	}
	if code != 1 || stderr.Len() != 0 || failed.Committed || failed.Text != failed.OriginalText || len(failed.Changes) != 1 || !failed.Changes[0].CandidateApplied || failed.Changes[0].Committed {
		t.Fatalf("code=%d report=%+v stderr=%s", code, failed, &stderr)
	}
}

func TestRewriteCLIJSONInputError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLIWithInput([]string{"rewrite", "--format=json", "--query=x"}, strings.NewReader(""), &stdout, &stderr)
	if code != 2 || stdout.Len() != 0 || !strings.HasPrefix(stderr.String(), `{"error":"`) || !strings.HasSuffix(stderr.String(), "\n") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
}

func TestCapabilitiesTextExposesOptionalRewriteManifest(t *testing.T) {
	manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Rewrite == nil {
		t.Fatal("missing test precondition")
	}
	got := string(formatCapabilitiesText(manifest))
	if !strings.Contains(got, "Rewrite forms:\n") || !strings.Contains(got, "field/expression_atom: supported=true") || !strings.Contains(got, "field/navigation: supported=false") ||
		!strings.Contains(got, " safe_rewriting=") || !strings.Contains(got, "Commands:\n") || !strings.Contains(got, "Functions:\n") {
		t.Fatalf("rewrite capability omitted from text:\n%s", got)
	}
}
