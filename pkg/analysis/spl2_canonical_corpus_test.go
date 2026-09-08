package analysis

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
)

type spl2CanonicalReference struct {
	OriginalName   string `json:"original_name"`
	NormalizedName string `json:"normalized_name"`
	Kind           string `json:"kind"`
	Role           string `json:"role"`
	Binding        string `json:"binding"`
	Start          int    `json:"start"`
	End            int    `json:"end"`
}
type spl2CanonicalField struct {
	Name        string `json:"name"`
	Conditional bool   `json:"conditional"`
}
type spl2CanonicalExpectation struct {
	Phase            string                   `json:"phase"`
	Scope            string                   `json:"scope"`
	Status           Status                   `json:"status"`
	SyntaxComplete   bool                     `json:"syntax_complete"`
	SemanticComplete bool                     `json:"semantic_complete"`
	ExpectedCodes    []string                 `json:"expected_codes"`
	References       []spl2CanonicalReference `json:"references"`
	Fields           []spl2CanonicalField     `json:"fields"`
	Removed          []string                 `json:"removed"`
	Open             bool                     `json:"open"`
	Uncertain        bool                     `json:"uncertain"`
	StageCommands    []string                 `json:"stage_commands"`
	StageComplete    []bool                   `json:"stage_complete"`
}

type spl2CanonicalAssertions struct {
	Status             Status   `json:"status"`
	SyntaxComplete     bool     `json:"syntax_complete"`
	SemanticComplete   bool     `json:"semantic_complete"`
	RequiredCodes      []string `json:"required_codes"`
	ForbiddenCodes     []string `json:"forbidden_codes"`
	RequiredReferences []struct {
		OriginalName   string  `json:"original_name"`
		NormalizedName string  `json:"normalized_name"`
		Kind           string  `json:"kind"`
		Role           string  `json:"role"`
		Binding        *string `json:"binding"`
		Start          int     `json:"start"`
		End            int     `json:"end"`
		ScopeStart     *int    `json:"scope_start"`
		ScopeEnd       *int    `json:"scope_end"`
	} `json:"required_references"`
	ForbiddenReferences []struct {
		OriginalName string  `json:"original_name"`
		Start        int     `json:"start"`
		End          int     `json:"end"`
		Role         *string `json:"role"`
		Binding      *string `json:"binding"`
	} `json:"forbidden_references"`
	RequiredFields      []spl2CanonicalField `json:"required_fields"`
	ForbiddenFieldNames []string             `json:"forbidden_field_names"`
	RequiredStages      []struct {
		Command          string `json:"command"`
		Start            int    `json:"start"`
		SemanticComplete bool   `json:"semantic_complete"`
		ScopeStart       *int   `json:"scope_start"`
		ScopeEnd         *int   `json:"scope_end"`
	} `json:"required_stages"`
	RequiredScopes []struct {
		Kind        string `json:"kind"`
		Start       int    `json:"start"`
		End         int    `json:"end"`
		OwnerStart  int    `json:"owner_start"`
		ParentStart int    `json:"parent_start"`
		ParentEnd   int    `json:"parent_end"`
	} `json:"required_scopes"`
	MaxScopes  int `json:"max_scopes"`
	FinalState *struct {
		Fields    []spl2CanonicalField `json:"fields"`
		Removed   []string             `json:"removed"`
		Open      bool                 `json:"open"`
		Uncertain bool                 `json:"uncertain"`
	} `json:"final_state"`
}

func spl2CheckCanonicalAssertions(r *Result, a *spl2CanonicalAssertions) error {
	if a == nil {
		return fmt.Errorf("snapshot lacks independent assertions")
	}
	if err := spl2CheckOriginalReferences(r); err != nil {
		return err
	}
	c := spl2CanonicalProjection(r)
	scopes := map[string]Scope{}
	stages := map[string]Stage{}
	for _, scope := range r.Scopes {
		scopes[scope.ID] = scope
	}
	for _, stage := range r.Stages {
		stages[stage.ID] = stage
	}
	scopeMatches := func(id string, start, end *int) bool {
		if start == nil && end == nil {
			return true
		}
		scope, ok := scopes[id]
		return ok && start != nil && end != nil && scope.Location.Start.Offset == *start && scope.Location.End.Offset == *end
	}
	if c.Status != a.Status || c.SyntaxComplete != a.SyntaxComplete || c.SemanticComplete != a.SemanticComplete {
		return fmt.Errorf("independent status/coverage mismatch: %s %v/%v", c.Status, c.SyntaxComplete, c.SemanticComplete)
	}
	contains := func(names []string, name string) bool {
		for _, n := range names {
			if n == name {
				return true
			}
		}
		return false
	}
	for _, code := range a.RequiredCodes {
		if !contains(c.ExpectedCodes, code) {
			return fmt.Errorf("missing required code %s", code)
		}
	}
	for _, code := range a.ForbiddenCodes {
		if contains(c.ExpectedCodes, code) {
			return fmt.Errorf("forbidden code %s", code)
		}
	}
	for _, required := range a.RequiredReferences {
		found := false
		for _, actual := range r.References {
			if actual.OriginalName == required.OriginalName && actual.NormalizedName == required.NormalizedName && actual.Kind == required.Kind && actual.Role == required.Role && actual.Location.Start.Offset == required.Start && actual.Location.End.Offset == required.End && (required.Binding == nil || actual.Binding == *required.Binding) && scopeMatches(actual.ScopeID, required.ScopeStart, required.ScopeEnd) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing required reference %+v", required)
		}
	}
	for _, forbidden := range a.ForbiddenReferences {
		for _, actual := range c.References {
			if actual.OriginalName == forbidden.OriginalName && actual.Start == forbidden.Start && actual.End == forbidden.End && (forbidden.Role == nil || actual.Role == *forbidden.Role) && (forbidden.Binding == nil || actual.Binding == *forbidden.Binding) {
				return fmt.Errorf("forbidden reference %+v", forbidden)
			}
		}
	}
	for _, required := range a.RequiredFields {
		found := false
		for _, actual := range c.Fields {
			if actual == required {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing required field %+v", required)
		}
	}
	for _, actual := range c.Fields {
		if contains(a.ForbiddenFieldNames, actual.Name) {
			return fmt.Errorf("forbidden field %s", actual.Name)
		}
	}
	for _, required := range a.RequiredStages {
		found := false
		for _, actual := range r.Stages {
			if actual.Command == required.Command && actual.Location.Start.Offset == required.Start && actual.SemanticComplete == required.SemanticComplete && scopeMatches(actual.ScopeID, required.ScopeStart, required.ScopeEnd) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing required stage %+v", required)
		}
	}
	for _, required := range a.RequiredScopes {
		found := false
		for _, actual := range r.Scopes {
			parent, hasParent := scopes[actual.ParentID]
			owner, hasOwner := stages[actual.StageID]
			if actual.Kind == required.Kind && actual.Location.Start.Offset == required.Start && actual.Location.End.Offset == required.End && hasParent && parent.Location.Start.Offset == required.ParentStart && parent.Location.End.Offset == required.ParentEnd && hasOwner && owner.Location.Start.Offset == required.OwnerStart && owner.ScopeID == parent.ID {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing required child scope %+v", required)
		}
	}
	if len(r.Scopes) > a.MaxScopes {
		return fmt.Errorf("scope count %d exceeds independent bound %d", len(r.Scopes), a.MaxScopes)
	}
	if f := a.FinalState; f != nil {
		if !reflect.DeepEqual(c.Fields, f.Fields) || !reflect.DeepEqual(c.Removed, f.Removed) || c.Open != f.Open || c.Uncertain != f.Uncertain {
			return fmt.Errorf("independent exact final state mismatch")
		}
	}
	return nil
}

func spl2CheckOriginalReferences(r *Result) error {
	for _, ref := range r.References {
		start, end := ref.Location.Start.Offset, ref.Location.End.Offset
		if start < 0 || start >= end || end > len(r.Document.Text) || r.Document.Text[start:end] != ref.OriginalName {
			return fmt.Errorf("non-original or empty reference span: %+v", ref)
		}
	}
	return nil
}

func TestSPL2IndependentAssertionMutations(t *testing.T) {
	const assertion = `{"status":"valid","syntax_complete":true,"semantic_complete":true,
"required_codes":[],"forbidden_codes":["SPL_SYNTAX_ERROR"],
"required_references":[{"original_name":"main","normalized_name":"main","kind":"dataset","role":"read","binding":"not_applicable","start":5,"end":9}],
"forbidden_references":[{"original_name":"host","start":18,"end":22,"role":"create"},{"original_name":"host","start":18,"end":22,"role":"read","binding":"derived"}],
"required_fields":[{"name":"host","conditional":false}],"forbidden_field_names":["invented"],
"required_stages":[{"command":"table","start":12,"semantic_complete":true}],"max_scopes":1,
"final_state":{"fields":[{"name":"host","conditional":false}],"removed":[],"open":false,"uncertain":false}}`
	mutations := map[string]func(*Result){
		"status":                             func(r *Result) { r.Status = Incomplete },
		"coverage":                           func(r *Result) { r.Coverage.SemanticComplete = false },
		"missing reference":                  func(r *Result) { r.References = r.References[1:] },
		"forbidden role with read preserved": func(r *Result) { ref := r.References[1]; ref.Role = "create"; r.References = append(r.References, ref) },
		"forbidden binding":                  func(r *Result) { r.References[1].Binding = "derived" },
		"empty original span": func(r *Result) {
			r.References[1].Location.End = r.References[1].Location.Start
			r.References[1].OriginalName = ""
		},
		"non-original span": func(r *Result) { r.References[1].Location.Start.Offset++ },
		"extra field": func(r *Result) {
			r.Lineage[len(r.Lineage)-1].After.Fields = append(r.Lineage[len(r.Lineage)-1].After.Fields, FieldBinding{Name: "extra"})
		},
		"false closedness": func(r *Result) { r.Lineage[len(r.Lineage)-1].After.Open = true },
		"extra scope":      func(r *Result) { r.Scopes = append(r.Scopes, r.Scopes[0]) },
		"stage ownership":  func(r *Result) { r.Stages[1].Location.Start.Offset++ },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			r, err := Analyze(QueryDocument{Text: "FROM main | table host", Language: "spl2"})
			if err != nil {
				t.Fatal(err)
			}
			var a spl2CanonicalAssertions
			if err := json.Unmarshal([]byte(assertion), &a); err != nil {
				t.Fatal(err)
			}
			if err := spl2CheckCanonicalAssertions(r, &a); err != nil {
				t.Fatalf("independent control: %v", err)
			}
			mutate(r)
			if err := spl2CheckCanonicalAssertions(r, &a); err == nil {
				t.Fatal("mutation escaped independent assertion")
			}
		})
	}
}

func TestSPL2IndependentScopeAssertions(t *testing.T) {
	const query = `FROM main | append [FROM child | table item] | where tail>0`
	start, end := strings.Index(query, "["), strings.Index(query, "]")+1
	item := strings.Index(query, "item")
	assertion := fmt.Sprintf(`{"status":"incomplete","syntax_complete":true,"semantic_complete":false,
"required_codes":["SPL_UNSUPPORTED_SEMANTICS"],"forbidden_codes":[],
"required_references":[{"original_name":"item","normalized_name":"item","kind":"field","role":"read","binding":"source","start":%d,"end":%d,"scope_start":%d,"scope_end":%d}],
"forbidden_references":[],"required_fields":[],"forbidden_field_names":[],
"required_stages":[{"command":"table","start":%d,"semantic_complete":true,"scope_start":%d,"scope_end":%d}],"max_scopes":2,"final_state":null,
"required_scopes":[{"kind":"search","start":%d,"end":%d,"owner_start":12,"parent_start":0,"parent_end":%d}]}`,
		item, item+4, start, end, strings.Index(query, "table"), start, end, start, end, len(query))
	for name, mutate := range map[string]func(*Result){
		"missing child":       func(r *Result) { r.Scopes = r.Scopes[:1] },
		"wrong owner":         func(r *Result) { r.Scopes[1].StageID = "stage-0" },
		"wrong parent bounds": func(r *Result) { r.Scopes[0].Location.End.Offset-- },
		"child stage moved": func(r *Result) {
			for i := range r.Stages {
				if r.Stages[i].Command == "table" {
					r.Stages[i].ScopeID = "scope-0"
				}
			}
		},
		"child reference moved": func(r *Result) {
			for i := range r.References {
				if r.References[i].NormalizedName == "item" {
					r.References[i].ScopeID = "scope-0"
				}
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			var a spl2CanonicalAssertions
			if err := json.Unmarshal([]byte(assertion), &a); err != nil {
				t.Fatal(err)
			}
			if err := spl2CheckCanonicalAssertions(r, &a); err != nil {
				t.Fatalf("independent scope control: %v", err)
			}
			mutate(r)
			if spl2CheckCanonicalAssertions(r, &a) == nil {
				t.Fatal("scope mutation escaped")
			}
		})
	}
}

// Optional output is raw regression evidence. Independent assertions must be
// frozen before this test is invoked for capture; it never writes corpus files.
func TestSPL2SnapshotAssertions(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "spl2")
	data, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		CaseFiles []string `json:"case_files"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	reports := map[string]struct {
		Report    *Result                  `json:"report"`
		Canonical spl2CanonicalExpectation `json:"canonical"`
	}{}
	for _, file := range manifest.CaseFiles {
		data, err = os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		var cases []spl2CorpusCase
		if err = json.Unmarshal(data, &cases); err != nil {
			t.Fatal(err)
		}
		for _, c := range cases {
			if c.CanonicalEvidence != "regression_snapshot" {
				continue
			}
			t.Run(c.ID, func(t *testing.T) {
				r, err := Analyze(c.Document)
				if err != nil {
					t.Fatal(err)
				}
				reports[c.ID] = struct {
					Report    *Result                  `json:"report"`
					Canonical spl2CanonicalExpectation `json:"canonical"`
				}{r, spl2CanonicalProjection(r)}
				if err = spl2CheckCanonicalAssertions(r, c.CanonicalAssertions); err != nil {
					t.Error(err)
				}
				assertCorpusIntegrity(t, r)
			})
		}
	}
	if path := os.Getenv("SPL_SPL2_CAPTURE"); path != "" {
		if len(reports) == 0 {
			t.Fatal("no independently asserted snapshots to capture")
		}
		data, err := json.MarshalIndent(reports, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if _, err = file.Write(append(data, '\n')); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSPL2CorpusCanonical(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "spl2")
	data, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var m struct {
		CaseFiles          []string `json:"case_files"`
		EnforceFinalFloors bool     `json:"enforce_final_floors"`
	}
	if e = json.Unmarshal(data, &m); e != nil {
		t.Fatal(e)
	}
	count, total := 0, 0
	reports := []spl2TransportReport{}
	for _, file := range m.CaseFiles {
		data, e = os.ReadFile(filepath.Join(root, file))
		if e != nil {
			t.Fatal(e)
		}
		var cases []spl2CorpusCase
		if e = json.Unmarshal(data, &cases); e != nil {
			t.Fatal(e)
		}
		for _, c := range cases {
			total++
			if c.Canonical == nil {
				if m.EnforceFinalFloors {
					t.Errorf("final closure is missing canonical obligation %s", c.ID)
				}
				continue
			}
			count++
			t.Run(c.ID, func(t *testing.T) {
				r, e := Analyze(c.Document)
				if e != nil {
					t.Fatal(e)
				}
				if recovery := c.Recovery; recovery != nil {
					parsed := parseSPL2Document(c.Document.Text)
					original := false
					for _, token := range parsed.tokens.GetAllTokens() {
						location := parsed.source.location(token.GetStart(), token.GetStop()+1)
						if location.Start.Offset == recovery.Start && location.End.Offset == recovery.End && token.GetText() == recovery.Command && token.GetTokenIndex() >= 0 {
							original = true
						}
					}
					if !original {
						t.Fatal("recovery classification is not an original token")
					}
					found := false
					for _, stage := range r.Stages {
						if stage.ID == recovery.StageID && stage.Command == recovery.Command && stage.Location.Start.Offset == recovery.Start && stage.Location.End.Offset >= recovery.End {
							for _, d := range r.Diagnostics {
								if d.StageID == stage.ID && d.Code == CodeUnsupportedSemantics && d.Severity == "warning" && d.Location.Start.Offset <= recovery.Start && d.Location.End.Offset >= recovery.End {
									found = true
								}
							}
							for _, ref := range r.References {
								if ref.StageID == stage.ID {
									t.Fatal("unknown command fabricated a reference")
								}
							}
						}
					}
					if !found {
						t.Fatal("recovery classification lacks actual owning stage and located unsupported diagnostic")
					}
				}
				if c.CanonicalAssertions != nil || c.CanonicalEvidence == "regression_snapshot" {
					if err := spl2CheckCanonicalAssertions(r, c.CanonicalAssertions); err != nil {
						t.Fatal(err)
					}
				}
				got := spl2CanonicalProjection(r)
				if !reflect.DeepEqual(got, *c.Canonical) {
					g, _ := json.Marshal(got)
					w, _ := json.Marshal(c.Canonical)
					t.Fatalf("canonical got %s\nwant %s", g, w)
				}
				if err := spl2CheckOriginalReferences(r); err != nil {
					t.Fatal(err)
				}
				reports = append(reports, spl2TransportReport{c.ID, c.Document, r})
			})
		}
	}
	if count == 0 {
		t.Fatal("no explicit canonical expectations")
	}
	if path := os.Getenv("SPL_SPL2_GO_REPORTS"); path != "" {
		if t.Failed() || len(reports) != total {
			t.Fatal("incomplete conformance cannot emit transport evidence")
		}
		spl2WriteTransport(t, root, path, reports)
	}
}

// This artifact is full Go transport evidence, separate from semantic authority.
type spl2TransportReport struct {
	ID       string        `json:"id"`
	Document QueryDocument `json:"document"`
	Report   *Result       `json:"report"`
}

func spl2WriteTransport(t *testing.T, fixtures, path string, reports []spl2TransportReport) {
	t.Helper()
	root := filepath.Join("..", "..")
	hashes := func(paths []string, base string) map[string]string {
		out := map[string]string{}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			relative, err := filepath.Rel(base, path)
			if err != nil {
				t.Fatal(err)
			}
			out[filepath.ToSlash(relative)] = fmt.Sprintf("%x", sha256.Sum256(data))
		}
		return out
	}
	fixturePaths, err := filepath.Glob(filepath.Join(fixtures, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	sourcePaths := []string{filepath.Join(root, "go.mod"), filepath.Join(root, "go.sum")}
	for _, directory := range []string{"pkg/analysis", "parser", "grammar"} {
		err := filepath.WalkDir(filepath.Join(root, directory), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && (filepath.Ext(path) == ".go" || filepath.Ext(path) == ".g4") {
				sourcePaths = append(sourcePaths, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	artifact := struct {
		SchemaVersion     int                   `json:"schema_version"`
		Kind              string                `json:"kind"`
		ConformanceCredit int                   `json:"conformance_credit"`
		SourceHashes      map[string]string     `json:"source_hashes"`
		FixtureHashes     map[string]string     `json:"fixture_hashes"`
		Reports           []spl2TransportReport `json:"reports"`
	}{1, "spl2-go-transport", 0, hashes(sourcePaths, root), hashes(fixturePaths, fixtures), reports}
	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err = file.Write(append(data, '\n')); err != nil {
		t.Fatal(err)
	}
}

func spl2CanonicalProjection(r *Result) spl2CanonicalExpectation {
	got := spl2CanonicalExpectation{Phase: "analysis", Scope: "canonical-result", Status: r.Status, SyntaxComplete: r.Coverage.SyntaxComplete, SemanticComplete: r.Coverage.SemanticComplete, ExpectedCodes: []string{}, References: []spl2CanonicalReference{}, Fields: []spl2CanonicalField{}, Removed: []string{}, StageCommands: []string{}, StageComplete: []bool{}}
	codes := map[string]bool{}
	for _, d := range r.Diagnostics {
		codes[d.Code] = true
	}
	for code := range codes {
		got.ExpectedCodes = append(got.ExpectedCodes, code)
	}
	sort.Strings(got.ExpectedCodes)
	for _, ref := range r.References {
		got.References = append(got.References, spl2CanonicalReference{ref.OriginalName, ref.NormalizedName, ref.Kind, ref.Role, ref.Binding, ref.Location.Start.Offset, ref.Location.End.Offset})
	}
	for _, stage := range r.Stages {
		got.StageCommands = append(got.StageCommands, stage.Command)
		got.StageComplete = append(got.StageComplete, stage.SemanticComplete)
	}
	if len(r.Lineage) > 0 {
		after := r.Lineage[len(r.Lineage)-1].After
		got.Removed = after.Removed
		got.Open = after.Open
		got.Uncertain = after.Uncertain
		for _, f := range after.Fields {
			got.Fields = append(got.Fields, spl2CanonicalField{f.Name, f.Conditional})
		}
	}
	return got
}
