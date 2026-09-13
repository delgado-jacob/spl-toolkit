package api

import (
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/document"
	"github.com/delgado-jacob/spl-toolkit/pkg/graph"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/sarif"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// Tooling routes accept inline canonical requests only. Published request and
// response contracts are reconciled into OpenAPI by update_validation_openapi.
func (s *Server) handleToolingScan(w http.ResponseWriter, r *http.Request) {
	s.handleCorpusExport(w, r, "scan")
}
func (s *Server) handleToolingGraph(w http.ResponseWriter, r *http.Request) {
	s.handleCorpusExport(w, r, "graph")
}
func (s *Server) handleToolingSARIF(w http.ResponseWriter, r *http.Request) {
	s.handleCorpusExport(w, r, "sarif")
}
func (s *Server) handleCorpusExport(w http.ResponseWriter, r *http.Request, format string) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, 400, err.Error())
		return
	}
	request, err := corpus.DecodeRequest(data)
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	report, err := corpus.Scan(request)
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	var value any = report
	switch format {
	case "graph":
		value, err = graph.Export(report)
	case "sarif":
		value, err = sarif.Export(report)
	}
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	s.writeJSONResponse(w, 200, value)
}
func (s *Server) handleToolingSchema(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, 400, err.Error())
		return
	}
	request, err := impact.DecodeSchemaRequest(data)
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	report, err := impact.CompareSchemas(request)
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	s.writeJSONResponse(w, 200, report)
}
func (s *Server) handleToolingMapping(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, 400, err.Error())
		return
	}
	request, err := impact.DecodeMappingRequest(data)
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	report, err := impact.CompareMappings(request)
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	s.writeJSONResponse(w, 200, report)
}
func (s *Server) handleDocument(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, 400, err.Error())
		return
	}
	documents, err := validation.DecodeDocuments(append(append([]byte{'['}, data...), ']'))
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	if len(documents) != 1 {
		s.writeErrorResponse(w, 400, "expected exactly one query document")
		return
	}
	result, err := analysis.Analyze(documents[0])
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	view, err := document.New(result, document.RevisionContext{ToolVersion: buildinfo.Version, ContractVersion: "1"})
	if err != nil {
		s.writeToolingError(w, err)
		return
	}
	s.writeJSONResponse(w, 200, view)
}
func (s *Server) writeToolingError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if corpus.IsInputError(err) || impact.IsInputError(err) || rewrite.IsInputError(err) || validation.IsInputError(err) {
		status = http.StatusBadRequest
	}
	s.writeErrorResponse(w, status, err.Error())
}
