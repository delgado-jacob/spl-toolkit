package rewrite

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// The corpus owns independent publication, exact audit, source binding and
// lineage assertions. Exported full reports are transport evidence only.
type conformanceCase struct {
	ID               string          `json:"id"`
	Groups           []string        `json:"groups"`
	Request          json.RawMessage `json:"request"`
	Target           json.RawMessage `json:"target"`
	Expected         map[string]any  `json:"expected"`
	CatalogFixture   string          `json:"catalog_fixture"`
	CatalogRawSHA256 string          `json:"catalog_raw_sha256"`
}

var rewriteRequiredGroups = []string{
	"preview", "apply", "aliases", "implicit",
	"conditions", "boolean", "contains", "fact-order",
	"scopes", "sql", "identity-kinds", "atom-path",
	"collisions", "chains-swaps", "dependent-skips", "bytes",
	"syntax", "validation", "no-op", "refusals",
	"lexical-controls", "batch", "all-false-unknown", "all-true-unknown",
	"and-compatible-repeat", "and-conflict", "and-conflict-contains", "any-false-unknown",
	"any-true-unknown", "atom-path-control", "chain", "coalescing",
	"common-or", "common-or-spl", "conditional-spl", "conditional-spl2",
	"conflict-independent", "conflict-or", "conflict-unknown", "contains-pattern-unknown",
	"contains-query-text-control", "dataset-component", "dependency-context-control", "dependency-index",
	"dependency-source", "dependency-sourcetype", "dependency-value-control", "dependent-collision",
	"empty-rules", "eval-alias", "field-target-invalid", "field-target-valid",
	"implicit-consumer", "implicit-refusal", "independent-child", "inherited-child",
	"json-schema-conditional", "json-schema-invalid", "json-schema-unresolved", "json-schema-valid",
	"later-fact", "literal-comment-control", "literal-target", "lookup-catalog-spl2",
	"macro-refusal", "metric-index", "navigation-refusal", "no-change",
	"no-match", "noncommon-or", "noncommon-or-spl", "not",
	"not-spl", "null-spl2", "ocsf-local", "overwritten-fact",
	"qualified-corender", "reference-present", "removed-fact", "rename-spl2",
	"sql-logical-order", "swap", "typed-boolean", "typed-contains-exact-star",
	"typed-null", "typed-number-string", "unicode-crlf", "wildcard-refusal",
}

// The matrix is separate from the full semantic corpus: each advertised role
// needs a real positive producer and a context in which mapping stays fixed.
func TestRewriteCapabilityMatrix(t *testing.T) {
	var fixture struct {
		Checks []struct {
			ID, Language, Kind, Role string
			Supported                bool
			Positive                 *struct{ Query, Source, Target, Candidate, SiteName string }
			Negative                 struct{ Query, Source, Target, Reason string }
		} `json:"capability_checks"`
	}
	raw, err := os.ReadFile("../../testdata/rewrite/forms.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	wanted := map[string]bool{}
	for _, language := range []string{"spl", "spl2"} {
		manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: language})
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range manifest.Rewrite.Forms {
			wanted[language+"/"+f.Kind+"/"+f.Role] = f.Supported
		}
	}
	seen, ids := map[string]bool{}, map[string]bool{}
	for _, c := range fixture.Checks {
		key := c.Language + "/" + c.Kind + "/" + c.Role
		if c.ID == "" || ids[c.ID] || seen[key] {
			t.Fatalf("duplicate or missing matrix identity %q", c.ID)
		}
		ids[c.ID], seen[key] = true, true
		advertised, exists := wanted[key]
		if !exists || advertised != c.Supported || c.Supported != (c.Positive != nil) {
			t.Fatalf("matrix disagrees with advertised %s", key)
		}
		t.Run(c.ID, func(t *testing.T) {
			if p := c.Positive; p != nil {
				doc := analysis.QueryDocument{Text: p.Query, Language: c.Language, SourceID: c.ID}
				session, err := analysis.PrepareRewrite(doc, nil)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, site := range session.Evidence().Sites {
					if site.Kind == c.Kind && site.Role == c.Role && site.Identity.Name != nil && *site.Identity.Name == p.SiteName {
						found = true
					}
				}
				if !found {
					t.Fatalf("advertised role has no positive producer: %s", key)
				}
				r := requireRewrite(t, Request{SchemaVersion: 1, Mode: Apply, Document: doc, Rules: []Rule{{ID: "matrix", Kind: c.Kind, Source: conditionIdentity(p.Source), Target: conditionIdentity(p.Target)}}})
				if !r.Committed || r.Text != p.Candidate || r.CandidateText != p.Candidate || !r.Coverage.RewriteComplete {
					t.Fatalf("positive candidate unproved: %+v", r)
				}
				assertConformanceBytes(t, r)
			}
			n := c.Negative
			if n.Query == "" || n.Reason == "" {
				t.Fatal("missing negative context")
			}
			if c.Kind == "field" && c.Supported && c.Role != "implicit_output" {
				session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: n.Query, Language: c.Language}, nil)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, site := range session.Evidence().Sites {
					if site.Kind == c.Kind && site.Role == c.Role && site.Identity.Name != nil && *site.Identity.Name == n.Source && site.Eligibility != "eligible" {
						found = true
					}
				}
				if !found {
					t.Fatalf("negative context does not exercise the held %s role", c.Role)
				}
			}
			source := conditionIdentity(n.Source)
			if c.Role == "navigation" {
				source = Identity{Path: []string{"actor", "name"}}
			}
			r := requireRewrite(t, Request{SchemaVersion: 1, Mode: Apply, Document: analysis.QueryDocument{Text: n.Query, Language: c.Language, SourceID: c.ID + "-negative"}, Rules: []Rule{{ID: "matrix", Kind: c.Kind, Source: source, Target: conditionIdentity(n.Target)}}})
			if r.Committed || r.Text != n.Query || r.CandidateText != n.Query {
				t.Fatal("negative context changed")
			}
			found := false
			for _, e := range r.RuleEvaluations {
				if e.Reason == n.Reason {
					found = true
				}
			}
			for _, e := range r.Changes {
				if e.Reason == n.Reason {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing negative reason %s: %+v / %+v", n.Reason, r.RuleEvaluations, r.Changes)
			}
		})
	}
	if len(seen) != len(wanted) {
		t.Fatalf("missing capability rows: got %d want %d", len(seen), len(wanted))
	}
}

// A published supported role must have an actual producer. A renderer switch
// alone does not establish that the selected dialect can emit that role.
func TestRewriteCapabilityNullInspectionRole(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		t.Run(language, func(t *testing.T) {
			manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: language})
			if err != nil {
				t.Fatal(err)
			}
			advertised := false
			for _, form := range manifest.Rewrite.Forms {
				if form.Kind == "field" && form.Role == "null_test" && form.Supported {
					advertised = true
				}
			}
			if !advertised {
				t.Fatal("null_test positive expectation requires the advertised supported role")
			}
			session, err := analysis.PrepareRewrite(analysis.QueryDocument{Language: language, Text: "search src=x | where isnull(src)"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			evidence := session.Evidence()
			if !evidence.Analysis.Coverage.SyntaxComplete || !evidence.Analysis.Coverage.SemanticComplete {
				t.Fatal("positive control must have complete canonical syntax and semantics")
			}
			found := false
			for _, site := range evidence.Sites {
				if site.Kind == "field" && site.Identity.Name != nil && *site.Identity.Name == "src" && site.Location.Start.Offset == 28 {
					found = true
					if site.Role != "null_test" || site.Eligibility != "eligible" {
						t.Fatalf("advertised null_test role has no matching producer: %+v", site)
					}
				}
			}
			if !found {
				t.Fatal("missing null-inspection operand")
			}
		})
	}
}

func readConformance(t *testing.T) []conformanceCase {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/rewrite/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []conformanceCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	ids, groups := map[string]bool{}, map[string]bool{}
	for i := range cases {
		c := &cases[i]
		if c.CatalogFixture != "" {
			if c.CatalogFixture != "ocsf/1.6.0/base.json.gz" || c.CatalogRawSHA256 != "9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137" {
				t.Fatalf("%s: unknown local catalog fixture", c.ID)
			}
			// Reuse the accepted local catalog loader and its raw-byte hash check.
			catalog := ocsfRewriteTarget(t).SchemaTarget.Catalog
			var request map[string]json.RawMessage
			var target map[string]json.RawMessage
			if err := json.Unmarshal(c.Request, &request); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(request["validation_target"], &target); err != nil {
				t.Fatal(err)
			}
			target["catalog"] = catalog
			var err error
			request["validation_target"], err = json.Marshal(target)
			if err != nil {
				t.Fatal(err)
			}
			c.Request, err = json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			c.Target = request["validation_target"]
		}
		if c.ID == "" || ids[c.ID] {
			t.Fatalf("missing or duplicate corpus ID %q", c.ID)
		}
		ids[c.ID] = true
		for _, g := range c.Groups {
			groups[g] = true
		}
		if len(c.Target) == 0 {
			t.Fatalf("%s: target decision must be explicit", c.ID)
		}
		for _, k := range []string{"status", "candidate_text", "text", "committed", "coverage", "changes", "rule_evaluations", "original_references", "candidate_references", "original_lineage", "candidate_lineage"} {
			if _, ok := c.Expected[k]; !ok {
				t.Fatalf("%s: missing independent %s expectation", c.ID, k)
			}
		}
		if string(c.Target) != "null" && c.Expected["validation_summary"] == nil {
			t.Fatalf("%s: target requires independent validation outcomes", c.ID)
		}
	}
	for _, group := range rewriteRequiredGroups {
		if !groups[group] {
			t.Errorf("empty required rewrite group %s", group)
		}
	}
	return cases
}

func TestRewriteConformanceRequiredGroups(t *testing.T) {
	groups := map[string]bool{}
	for _, c := range readConformance(t) {
		for _, g := range c.Groups {
			groups[g] = true
		}
	}
	for _, g := range rewriteRequiredGroups {
		if !groups[g] {
			t.Errorf("empty required rewrite group %s", g)
		}
	}
}

// Fault injection is a package-private proof seam, not a public API feature.
// It deliberately supplies a stale successful summary to the real final gate.
func TestRewriteConformanceNegativeProof(t *testing.T) {
	var fixture struct {
		Cases []struct {
			conformanceCase
			Fault     string `json:"fault"`
			ProofCode string `json:"proof_code"`
		} `json:"safety_cases"`
	}
	raw, err := os.ReadFile("../../testdata/rewrite/forms.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	groups, ids := map[string]bool{}, map[string]bool{}
	for _, c := range fixture.Cases {
		if c.ID == "" || ids[c.ID] {
			t.Fatal("missing or duplicate safety ID")
		}
		ids[c.ID], groups[c.Fault] = true, true
		t.Run(c.ID, func(t *testing.T) {
			request, err := DecodeRequest(c.Request)
			if err != nil {
				t.Fatal(err)
			}
			if request.Mode != Apply || len(request.Rules) != 1 || string(c.Target) != "null" {
				t.Fatal("invalid internal proof fixture")
			}
			var corrupt func([]analysis.RewriteTextEdit)
			switch c.Fault {
			case "candidate-syntax":
				corrupt = func(edits []analysis.RewriteTextEdit) { edits[0].After = "'" }
			case "candidate-binding":
			default:
				t.Fatalf("unknown proof fault %s", c.Fault)
			}
			pending := internalProposedEdit(t, request.Document, *request.Rules[0].Source.Name, *request.Rules[0].Target.Name, corrupt)
			proof := pending.original.Verify(pending.candidate.Session, pending.candidate.Rendering)
			if proof.Proven || len(proof.Limitations) == 0 || proof.Limitations[0].Code != c.ProofCode {
				t.Fatalf("wrong proof failure: %+v", proof)
			}
			r := finishRewrite(pending, Apply, nil)
			got := conformanceProjection(r)
			if !reflect.DeepEqual(got, c.Expected) {
				a, _ := json.Marshal(got)
				b, _ := json.Marshal(c.Expected)
				t.Fatalf("negative proof report\ngot %s\nwant %s", a, b)
			}
			assertConformanceBytes(t, r)
		})
	}
	for _, group := range []string{"candidate-syntax", "candidate-binding"} {
		if !groups[group] {
			t.Errorf("empty required proof group %s", group)
		}
	}
}

func conformanceProjection(r *Result) map[string]any {
	raw, _ := json.Marshal(r)
	var report map[string]any
	_ = json.Unmarshal(raw, &report)
	out := map[string]any{}
	for _, k := range []string{"status", "candidate_text", "text", "committed", "coverage", "changes", "rule_evaluations"} {
		out[k] = report[k]
	}
	if wrapper, ok := report["candidate_validation"].(map[string]any); ok {
		key := "schema"
		if wrapper["kind"] == "field_list" {
			key = "field_list"
		}
		validation := wrapper[key].(map[string]any)
		outcomes := []any{}
		for _, raw := range validation["outcomes"].([]any) {
			item := raw.(map[string]any)
			outcomes = append(outcomes, []any{item["reference_id"], item["outcome"]})
		}
		out["validation_summary"] = []any{wrapper["kind"], validation["status"], outcomes}
	}
	for _, key := range []string{"changes", "rule_evaluations"} {
		for _, value := range out[key].([]any) {
			entry := value.(map[string]any)
			for _, locKey := range []string{"location", "original_location", "candidate_location"} {
				if value, ok := entry[locKey]; ok {
					loc := value.(map[string]any)
					entry[locKey] = []any{loc["start"].(map[string]any)["offset"], loc["end"].(map[string]any)["offset"]}
				}
			}
		}
	}
	for _, prefix := range []string{"original", "candidate"} {
		a := report[prefix+"_analysis"].(map[string]any)
		refs, lineage := []any{}, []any{}
		for _, value := range a["references"].([]any) {
			r := value.(map[string]any)
			refs = append(refs, []any{r["id"], r["normalized_name"], r["kind"], r["role"], r["binding"], r["origin_reference_ids"]})
		}
		for _, value := range a["lineage"].([]any) {
			l := value.(map[string]any)
			fields, transitions := []any{}, []any{}
			for _, f := range l["after"].(map[string]any)["fields"].([]any) {
				field := f.(map[string]any)
				fields = append(fields, []any{field["name"], field["conditional"]})
			}
			for _, tr := range l["transitions"].([]any) {
				transitions = append(transitions, tr)
			}
			phase := l["phase"]
			if phase == nil {
				phase = ""
			}
			lineage = append(lineage, []any{l["stage_id"], l["scope_id"], phase, fields, transitions})
		}
		out[prefix+"_references"], out[prefix+"_lineage"] = refs, lineage
	}
	return out
}

func assertConformanceBytes(t *testing.T, r *Result) {
	t.Helper()
	for _, a := range []*analysis.Result{r.OriginalAnalysis, r.CandidateAnalysis} {
		for _, ref := range a.References {
			start, end := ref.Location.Start.Offset, ref.Location.End.Offset
			if start < 0 || end < start || end > len(a.Document.Text) || a.Document.Text[start:end] != ref.OriginalName {
				t.Fatalf("invalid independent source slice: %+v", ref)
			}
		}
	}
	originalEnd, candidateEnd := 0, 0
	for _, c := range r.Changes {
		if !c.CandidateApplied {
			continue
		}
		if c.OriginalLocation == nil || c.CandidateLocation == nil {
			t.Fatal("included edit lacks locations")
		}
		a, b := c.OriginalLocation, c.CandidateLocation
		if a.Start.Offset < originalEnd || b.Start.Offset < candidateEnd || a.End.Offset > len(r.OriginalText) || b.End.Offset > len(r.CandidateText) {
			t.Fatal("invalid edit ranges")
		}
		if r.OriginalText[originalEnd:a.Start.Offset] != r.CandidateText[candidateEnd:b.Start.Offset] || r.OriginalText[a.Start.Offset:a.End.Offset] != c.OldText || r.CandidateText[b.Start.Offset:b.End.Offset] != c.NewText {
			t.Fatal("audit changed untouched bytes or mislocated a replacement")
		}
		originalEnd, candidateEnd = a.End.Offset, b.End.Offset
	}
	if r.OriginalText[originalEnd:] != r.CandidateText[candidateEnd:] {
		t.Fatal("tail bytes changed")
	}
}

func TestRewriteConformance(t *testing.T) {
	cases := readConformance(t)
	entries := []any{}
	results := map[string]*Result{}
	coveredForms := map[string]bool{}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			request, err := DecodeRequest(c.Request)
			if err != nil {
				t.Fatal(err)
			}
			if request.Document.Language == "" || request.Document.SourceID != c.ID {
				t.Fatal("explicit document identity is required")
			}
			r, err := Rewrite(request)
			if err != nil {
				t.Fatal(err)
			}
			if r.Document != request.Document || r.OriginalText != request.Document.Text || r.CandidateAnalysis.Document.SourceID != c.ID {
				t.Fatal("source identity changed")
			}
			got := conformanceProjection(r)
			for key, want := range c.Expected {
				if !reflect.DeepEqual(got[key], want) {
					actual, _ := json.Marshal(got[key])
					expected, _ := json.Marshal(want)
					t.Errorf("%s\ngot  %s\nwant %s", key, actual, expected)
				}
			}
			assertConformanceBytes(t, r)
			// Capability presence is not surface proof. Every advertised positive
			// role must occur in an actual candidate-applied corpus edit.
			session, err := analysis.PrepareRewrite(request.Document, nil)
			if err != nil {
				t.Fatal(err)
			}
			applied := map[string]bool{}
			for _, change := range r.Changes {
				if change.CandidateApplied {
					for _, id := range change.OriginalReferenceIDs {
						applied[id] = true
					}
				}
			}
			for _, site := range session.Evidence().Sites {
				changed := applied[site.ReferenceID]
				// An implicit definition is re-rendered by its input edit, not an
				// additional overlapping edit. Require its actual candidate name
				// to change, alongside the independently asserted lineage/report.
				if site.Role == "implicit_output" && site.Identity.Name != nil && len(applied) > 0 {
					for _, ref := range r.CandidateAnalysis.References {
						if ref.ID == site.ReferenceID && ref.NormalizedName != *site.Identity.Name {
							changed = true
						}
					}
				}
				if changed {
					coveredForms[request.Document.Language+"/"+site.Kind+"/"+site.Role] = true
				}
			}
			results[c.ID] = r
			entries = append(entries, map[string]any{"id": c.ID, "request": json.RawMessage(c.Request), "report": r})
		})
	}
	if t.Failed() {
		return
	}
	for _, language := range []string{"spl", "spl2"} {
		capabilities, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: language})
		if err != nil {
			t.Fatal(err)
		}
		for _, form := range capabilities.Rewrite.Forms {
			if form.Supported && !coveredForms[language+"/"+form.Kind+"/"+form.Role] {
				t.Errorf("missing full-report positive form %s/%s/%s", language, form.Kind, form.Role)
			}
		}
	}
	if t.Failed() {
		return
	}
	// Exercise the actual batch entrypoint as well as single-query transport.
	groups := map[string][]conformanceCase{}
	for _, c := range cases {
		var shared map[string]json.RawMessage
		if err := json.Unmarshal(c.Request, &shared); err != nil {
			t.Fatal(err)
		}
		delete(shared, "document")
		key, err := json.Marshal(shared)
		if err != nil {
			t.Fatal(err)
		}
		groups[string(key)] = append(groups[string(key)], c)
	}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	batches := []any{}
	mixed := false
	for _, key := range keys {
		group := groups[key]
		// Schema targets use the strict wire decoder, not struct unmarshalling.
		single, err := DecodeRequest(group[0].Request)
		if err != nil {
			t.Fatal(err)
		}
		request := BatchRequest{SchemaVersion: 1, Mode: single.Mode, Rules: single.Rules, ValidationTarget: single.ValidationTarget}
		ids, reports := []string{}, []*Result{}
		status := analysis.Valid
		statuses := map[analysis.Status]bool{}
		for i := len(group) - 1; i >= 0; i-- {
			c := group[i]
			request.Documents = append(request.Documents, results[c.ID].Document)
			ids = append(ids, c.ID)
			reports = append(reports, results[c.ID])
			statuses[results[c.ID].Status] = true
			if results[c.ID].Status == analysis.Invalid {
				status = analysis.Invalid
			} else if results[c.ID].Status == analysis.Incomplete && status != analysis.Invalid {
				status = analysis.Incomplete
			}
		}
		mixed = mixed || len(statuses) == 3
		got, err := RewriteBatch(request)
		if err != nil {
			t.Fatal(err)
		}
		want := &BatchResult{SchemaVersion: 1, Status: status, Reports: reports}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("batch differs from independently asserted ordered singles: %v", ids)
		}
		batches = append(batches, map[string]any{"ids": ids, "report": got})
	}
	if !mixed {
		t.Fatal("empty required valid/incomplete/invalid mixed batch")
	}
	if path := os.Getenv("SPL_REWRITE_GO_REPORTS"); path != "" {
		root, _ := filepath.Abs("../..")
		hashes := map[string]string{}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relative, _ := filepath.Rel(root, path)
			relative = filepath.ToSlash(relative)
			if d.IsDir() {
				if relative != "." && !(relative == "pkg" || relative == "internal" || relative == "parser" || relative == "grammar" || strings.HasPrefix(relative, "pkg/") || strings.HasPrefix(relative, "internal/") || strings.HasPrefix(relative, "parser/")) {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(relative, ".go") || strings.HasSuffix(relative, ".g4") || relative == "go.mod" || relative == "go.sum" {
				raw, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				hashes[relative] = fmt.Sprintf("%x", sha256.Sum256(raw))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		fixtures := map[string]string{}
		paths, err := filepath.Glob("../../testdata/rewrite/*.json")
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(paths)
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			fixtures[filepath.Base(path)] = fmt.Sprintf("%x", sha256.Sum256(raw))
		}
		artifact := map[string]any{"schema_version": 1, "kind": "rewrite-go-transport", "conformance_credit": 0, "source_hashes": hashes, "fixture_hashes": fixtures, "reports": entries, "batches": batches}
		raw, err := json.Marshal(artifact)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
