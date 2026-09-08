package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
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

func TestSPL2CorpusCanonical(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "spl2")
	data, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var m struct {
		CaseFiles []string `json:"case_files"`
	}
	if e = json.Unmarshal(data, &m); e != nil {
		t.Fatal(e)
	}
	count := 0
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
			if c.Canonical == nil {
				continue
			}
			count++
			t.Run(c.ID, func(t *testing.T) {
				r, e := Analyze(c.Document)
				if e != nil {
					t.Fatal(e)
				}
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
				if !reflect.DeepEqual(got, *c.Canonical) {
					g, _ := json.Marshal(got)
					w, _ := json.Marshal(c.Canonical)
					t.Fatalf("canonical got %s\nwant %s", g, w)
				}
				for _, ref := range r.References {
					if r.Document.Text[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
						t.Fatal("source identity changed")
					}
				}
			})
		}
	}
	if count == 0 {
		t.Fatal("no explicit canonical expectations")
	}
}
