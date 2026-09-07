package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// handleAnalyzeQuery returns the canonical structured analysis report.
// @Summary Analyze an SPL query
// @Description Return structured syntax and semantic coverage, located references, lineage, dependencies, and diagnostics for an SPL query.
// @Tags query
// @Accept json
// @Produce json
// @Param request body analysis.QueryDocument true "Query document"
// @Success 200 {object} analysis.Result "Canonical analysis report for valid, invalid, and incomplete queries"
// @Failure 400 {object} ErrorResponse "Malformed JSON or unsupported document options"
// @Router /query/analyze [post]
func (s *Server) handleAnalyzeQuery(w http.ResponseWriter, r *http.Request) {
	document, err := parseAnalysisDocument(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := analysis.Analyze(document)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

// handleCapabilities returns the canonical analysis capability manifest.
// @Summary Get analysis capabilities
// @Description Return the supported analysis language, profile, compatibility version, commands, functions, and limitations.
// @Tags query
// @Produce json
// @Success 200 {object} analysis.CapabilityManifest "Analysis capability manifest"
// @Router /capabilities [get]
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	s.writeJSONResponse(w, http.StatusOK, analysis.Capabilities())
}

func parseAnalysisDocument(w http.ResponseWriter, r *http.Request) (analysis.QueryDocument, error) {
	var document analysis.QueryDocument
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "application/json" {
		return document, fmt.Errorf("content-type must be application/json")
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return document, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := jsoninput.ValidateUnicode(body); err != nil {
		return document, fmt.Errorf("invalid JSON: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var decoded *analysis.QueryDocument
	if err := decoder.Decode(&decoded); err != nil {
		return document, fmt.Errorf("invalid JSON: %w", err)
	}
	if decoded == nil {
		return document, fmt.Errorf("invalid JSON: expected a query document")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return document, fmt.Errorf("invalid JSON: request body must contain a single JSON document")
		}
		return document, fmt.Errorf("invalid JSON: %w", err)
	}
	return *decoded, nil
}
