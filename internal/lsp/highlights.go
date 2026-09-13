package lsp

import (
	"fmt"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/document"
)

type Highlight struct {
	Range Range `json:"range"`
	Kind  int   `json:"kind"`
}

func documentHighlights(result *analysis.Result, position Position) ([]Highlight, error) {
	offset, err := offsetAt(result.Document.Text, position)
	if err != nil {
		return nil, err
	}
	view, err := document.New(result, document.RevisionContext{ToolVersion: buildinfo.Version, ContractVersion: "analysis-report-v1"})
	if err != nil {
		return nil, err
	}
	refs, err := view.ReferencesAt(offset)
	if err != nil {
		return nil, err
	}
	empty := []Highlight{}
	if len(refs) != 1 || !highlightable(refs[0]) {
		return empty, nil
	}
	byID := map[string]analysis.Reference{}
	edges := map[string][]string{}
	for _, r := range view.References {
		if _, ok := byID[r.ID]; ok {
			return nil, fmt.Errorf("duplicate canonical reference ID")
		}
		byID[r.ID] = r
	}
	for _, r := range view.References {
		if !highlightable(r) {
			continue
		}
		for _, id := range r.OriginReferenceIDs {
			origin, ok := byID[id]
			if ok && highlightable(origin) {
				edges[r.ID] = append(edges[r.ID], id)
				edges[id] = append(edges[id], r.ID)
			}
		}
	}
	seen := map[string]bool{}
	queue := []string{refs[0].ID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		queue = append(queue, edges[id]...)
	}
	out := []Highlight{}
	ranges := map[Range]bool{}
	for _, r := range view.References {
		if !seen[r.ID] {
			continue
		}
		start, err := positionAt(result.Document.Text, r.Location.Start.Offset)
		if err != nil {
			return nil, err
		}
		end, err := positionAt(result.Document.Text, r.Location.End.Offset)
		if err != nil {
			return nil, err
		}
		span := Range{start, end}
		if ranges[span] {
			continue
		}
		ranges[span] = true
		kind := 2
		switch r.Role {
		case "create", "rename", "output":
			kind = 3
		}
		out = append(out, Highlight{span, kind})
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Range.Start, out[j].Range.Start
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Character < b.Character
	})
	return out, nil
}
func highlightable(r analysis.Reference) bool {
	return r.Kind == "field" && r.Resolution == "exact" && r.Binding != "indeterminate" && r.Binding != "unavailable"
}
