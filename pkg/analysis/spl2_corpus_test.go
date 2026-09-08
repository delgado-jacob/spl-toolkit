package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type spl2RecoveryClassification struct {
	Kind    string `json:"kind"`
	Command string `json:"command"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	StageID string `json:"stage_id"`
}
type spl2CorpusCase struct {
	Recovery         *spl2RecoveryClassification `json:"recovery_classification"`
	Canonical        *spl2CanonicalExpectation   `json:"canonical"`
	ID               string                      `json:"id"`
	ObligationIDs    []string                    `json:"obligation_ids"`
	SourceKeys       []string                    `json:"source_keys"`
	Document         QueryDocument               `json:"document"`
	FormIDs          []string                    `json:"form_ids"`
	SyntaxComplete   bool                        `json:"syntax_complete"`
	SemanticComplete bool                        `json:"semantic_complete"`
	Status           string                      `json:"status"`
	ExpectedCodes    []string                    `json:"expected_codes"`
	ForbiddenCodes   []string                    `json:"forbidden_codes"`
	Assertions       struct {
		Kinds         map[string]int `json:"kinds"`
		ShapeContains []string       `json:"shape_contains"`
		Shapes        []struct {
			Kind  string `json:"kind"`
			Shape string `json:"shape"`
		} `json:"shapes"`
		DiagnosticExcerpts []string `json:"diagnostic_excerpts"`
		Excerpts           []struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		} `json:"excerpts"`
	} `json:"assertions"`
}

func TestSPL2CorpusSyntax(t *testing.T) {
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
	seen := map[string]bool{}
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
			t.Run(c.ID, func(t *testing.T) {
				if seen[c.ID] {
					t.Fatal("duplicate case ID")
				}
				seen[c.ID] = true
				if len(c.ObligationIDs) == 0 || len(c.SourceKeys) == 0 || len(c.FormIDs) == 0 {
					t.Fatal("missing evidence identity")
				}
				p := parseSPL2Document(c.Document.Text)
				if p.source.text != c.Document.Text {
					t.Fatal("source changed")
				}
				if p.syntaxComplete != c.SyntaxComplete || p.semanticComplete != c.SemanticComplete {
					t.Errorf("coverage got %v/%v want %v/%v: %+v", p.syntaxComplete, p.semanticComplete, c.SyntaxComplete, c.SemanticComplete, p.diagnostics)
				}
				codes := map[string]bool{}
				status := "incomplete"
				for _, d := range p.diagnostics {
					codes[d.Code] = true
					if d.Severity == "error" {
						status = "invalid"
					}
					if d.Location.Start.Offset < 0 || d.Location.End.Offset > len(c.Document.Text) {
						t.Fatal("diagnostic outside source")
					}
				}
				if status != c.Status {
					t.Errorf("status %s want %s: %+v", status, c.Status, p.diagnostics)
				}
				for _, code := range c.ExpectedCodes {
					if !codes[code] {
						t.Errorf("missing %s: %+v", code, p.diagnostics)
					}
				}
				for _, code := range c.ForbiddenCodes {
					if codes[code] {
						t.Errorf("forbidden %s: %+v", code, p.diagnostics)
					}
				}
				for kind, want := range c.Assertions.Kinds {
					if got := len(spl2Nodes(p.syntax, kind)); got != want {
						t.Errorf("%s count %d want %d", kind, got, want)
					}
				}
				for _, want := range c.Assertions.ShapeContains {
					if !strings.Contains(p.syntax.shape(), want) {
						t.Errorf("missing shape %s in %s", want, p.syntax.shape())
					}
				}
				for _, expected := range c.Assertions.Shapes {
					found := false
					for _, node := range spl2Nodes(p.syntax, expected.Kind) {
						if node.shape() == expected.Shape {
							found = true
						}
					}
					if !found {
						t.Errorf("missing exact %s shape %s", expected.Kind, expected.Shape)
					}
				}
				for _, expected := range c.Assertions.DiagnosticExcerpts {
					found := false
					for _, diagnostic := range p.diagnostics {
						location := diagnostic.Location
						if c.Document.Text[location.Start.Offset:location.End.Offset] == expected {
							found = true
						}
					}
					if !found {
						t.Errorf("missing diagnostic source excerpt %q", expected)
					}
				}
				for _, assert := range c.Assertions.Excerpts {
					found := false
					for _, node := range spl2Nodes(p.syntax, assert.Kind) {
						loc := node.Location
						if c.Document.Text[loc.Start.Offset:loc.End.Offset] == assert.Text {
							found = true
						}
					}
					if !found {
						t.Errorf("missing located %s %q", assert.Kind, assert.Text)
					}
				}
			})
		}
	}
	if len(seen) == 0 {
		t.Fatal("empty corpus")
	}
}
