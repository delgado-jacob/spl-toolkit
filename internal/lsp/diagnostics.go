package lsp

import (
	"fmt"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Code     string `json:"code"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}
type evaluation struct {
	analysis    *analysis.Result
	diagnostics []Diagnostic
}
type analyzeFunc func(snapshot) (evaluation, error)

func canonicalAnalysis(s snapshot) (evaluation, error) {
	report, err := s.Config.prepared.Scan(corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true}, Entries: []corpus.Entry{{ID: s.URI, Origin: corpus.Origin{Kind: "inline"}, Document: &s.Document}}})
	if err != nil {
		return evaluation{}, err
	}
	e := report.Entries[0].Evaluation
	var result *analysis.Result
	var diagnostics []analysis.Diagnostic
	switch e.Kind {
	case "analysis":
		result = e.Analysis
		diagnostics = result.Diagnostics
	case "field_list":
		result = e.FieldValidation.Analysis
		diagnostics = e.FieldValidation.Diagnostics
	default:
		result = e.SchemaValidation.Analysis
		diagnostics = e.SchemaValidation.Diagnostics
	}
	out := evaluation{analysis: result, diagnostics: []Diagnostic{}}
	for _, d := range diagnostics {
		start, err := positionAt(s.Document.Text, d.Location.Start.Offset)
		if err != nil {
			return evaluation{}, fmt.Errorf("diagnostic start: %w", err)
		}
		end, err := positionAt(s.Document.Text, d.Location.End.Offset)
		if err != nil || d.Location.End.Offset < d.Location.Start.Offset {
			return evaluation{}, fmt.Errorf("invalid diagnostic range")
		}
		severity := 3
		switch d.Severity {
		case "error":
			severity = 1
		case "warning":
			severity = 2
		case "hint":
			severity = 4
		}
		out.diagnostics = append(out.diagnostics, Diagnostic{Range{start, end}, severity, d.Code, "spl-toolkit", d.Message})
	}
	return out, nil
}

func (s *server) publish(doc snapshot, diagnostics []Diagnostic) error {
	params := struct {
		URI         string       `json:"uri"`
		Version     *int32       `json:"version,omitempty"`
		Diagnostics []Diagnostic `json:"diagnostics"`
	}{URI: doc.URI, Diagnostics: diagnostics}
	if s.versionSupport {
		v := doc.Version
		params.Version = &v
	}
	return s.writer.notify("textDocument/publishDiagnostics", params)
}
