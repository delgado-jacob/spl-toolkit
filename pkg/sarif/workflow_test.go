package sarif

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

func TestAdditionalFindingsReuseCoordinatesRulesAndArtifacts(t *testing.T) {
	c := loadCases(t)
	source := "a😀\r\nz\n"
	e := analyzed("doc", source, corpus.Origin{Kind: "file", BaseURI: c.WindowsRoot, RelativePath: c.RelativePath}, diag("Z_RULE", "warning", "original", point(0, 1, 1), point(1, 1, 2)))
	locations := []analysis.Location{
		{Start: point(1, 1, 2), End: point(5, 1, 3)},
		{Start: point(5, 1, 3), End: point(7, 2, 1)},
		{Start: point(len(source), 3, 1), End: point(len(source), 3, 1)},
	}
	findings := []AdditionalFinding{}
	for _, loc := range locations {
		findings = append(findings, AdditionalFinding{DocumentID: "doc", RuleID: "A_RULE", Level: "warning", Message: "additional", Location: &loc, Properties: map[string]any{"nested": map[string]any{"value": "original"}}})
	}
	findings = append(findings, AdditionalFinding{DocumentID: "doc", RuleID: "Z_RULE", Level: "note", Message: "unknown"})
	before, _ := json.Marshal(findings)
	log, err := ExportAdditional(report(e), findings)
	if err != nil {
		t.Fatal(err)
	}
	run := log.Runs[0]
	if len(run.Artifacts) != 1 || run.Artifacts[0].Location.URI != c.EncodedRelativeURI || len(run.Tool.Driver.Rules) != 2 {
		t.Fatalf("run: %+v", run)
	}
	for _, r := range run.Results {
		if run.Tool.Driver.Rules[r.RuleIndex].ID != r.RuleID {
			t.Fatal("rule index mismatch")
		}
	}
	regions := []*Region{run.Results[1].Locations[0].PhysicalLocation.Region, run.Results[2].Locations[0].PhysicalLocation.Region, run.Results[3].Locations[0].PhysicalLocation.Region}
	if *regions[0] != (Region{StartLine: 1, StartColumn: 2, EndLine: 1, EndColumn: 3}) || *regions[1] != (Region{StartLine: 1, StartColumn: 3, EndLine: 2, EndColumn: 1}) || *regions[2] != (Region{StartLine: 2, StartColumn: 3, EndLine: 2, EndColumn: 3}) {
		t.Fatalf("regions: %+v %+v %+v", regions[0], regions[1], regions[2])
	}
	if len(run.Results[4].Locations) != 0 {
		t.Fatal("unknown location guessed")
	}
	run.Results[1].Properties["nested"].(map[string]any)["value"] = "changed"
	after, _ := json.Marshal(findings)
	if string(before) != string(after) {
		t.Fatal("findings mutated by export")
	}
}

func TestAdditionalFindingsRejectInvalidSourceAndLocation(t *testing.T) {
	e := analyzed("doc", "a😀", corpus.Origin{Kind: "inline"})
	for _, tc := range []struct {
		name string
		id   string
		loc  *analysis.Location
	}{
		{"unknown document", "missing", nil},
		{"split code point", "doc", &analysis.Location{Start: point(2, 1, 2), End: point(3, 1, 3)}},
		{"outside source", "doc", &analysis.Location{Start: point(0, 1, 1), End: point(100, 1, 2)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ExportAdditional(report(e), []AdditionalFinding{{DocumentID: tc.id, RuleID: "RULE", Level: "warning", Message: "message", Location: tc.loc}}); err == nil {
				t.Fatal("invalid finding admitted")
			}
		})
	}
	e.SourceHash = "wrong"
	if _, err := ExportAdditional(report(e), nil); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("hash: %v", err)
	}
	a := analyzed("one", "x", corpus.Origin{Kind: "file", BaseURI: "file:///one/", RelativePath: "x.spl"})
	b := analyzed("two", "x", corpus.Origin{Kind: "file", BaseURI: "file:///two/", RelativePath: "x.spl"})
	if _, err := ExportAdditional(report(a, b), nil); err == nil || !strings.Contains(err.Error(), "multiple file roots") {
		t.Fatalf("roots: %v", err)
	}
}
