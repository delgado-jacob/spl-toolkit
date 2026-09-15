package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// handleAnalyzeQuery returns the canonical structured analysis report.
// @Summary Analyze an SPL or standalone SPL2 query
// @Description Strict query document requiring text, with optional language (spl or spl2), profile (splunkd), version (current), and source_id strings. Empty selectors use defaults spl/splunkd/current. Preserve text and identity exactly. Reject null, duplicate or unknown members, malformed Unicode, and trailing JSON. Body limit is 1 MiB. Return canonical coverage, located references, SQL execution phases, lineage, dependencies and diagnostics; all content statuses use 200.
// @Tags query
// @Accept json
// @Produce json
// @Param request body AnalysisRequest true "Query document"
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

// handleRequirementsQuery returns the canonical direct query requirements.
// @Summary Report direct external query requirements
// @Description Strict query document requiring text, with optional language (spl or spl2), profile (splunkd), version (current), and source_id strings. Empty selectors use defaults spl/splunkd/current. Preserve text and identity exactly. Reject null, duplicate or unknown members, malformed Unicode, and trailing JSON. Body limit is 1 MiB. Return canonical query identity, capability revision, query status, requirement coverage, direct items, gaps, and diagnostics; all content statuses use 200.
// @Tags query
// @Accept json
// @Produce json
// @Param request body AnalysisRequest true "Query document"
// @Success 200 {object} analysis.RequirementSet "Canonical requirements for valid, invalid, and incomplete queries"
// @Failure 400 {object} ErrorResponse "Malformed JSON or unsupported document options"
// @Router /query/requirements [post]
func (s *Server) handleRequirementsQuery(w http.ResponseWriter, r *http.Request) {
	document, err := parseAnalysisDocument(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	set, err := analysis.Requirements(document)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSONResponse(w, http.StatusOK, set)
}

// handleCapabilities returns the canonical analysis capability manifest.
// @Summary Get analysis capabilities
// @Description Select the canonical capability manifest including language, profile, compatibility version, commands, functions, limitations, and optional pinned documentation_snapshot. Only language/profile/version query parameters are accepted; duplicate keys are rejected even when equal.
// @Tags query
// @Produce json
// @Param language query string false "Query language: spl (default) or spl2; empty uses default" Enums(,spl,spl2)
// @Param profile query string false "Execution profile: splunkd (default); empty uses default" Enums(,splunkd)
// @Param version query string false "Compatibility version: current (default); empty uses default" Enums(,current)
// @Failure 400 {object} ErrorResponse "Unknown, duplicate, conflicting, malformed, or unsupported selectors"
// @Success 200 {object} analysis.CapabilityManifest "Analysis capability manifest"
// @Router /capabilities [get]
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	selectors, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "invalid capability selectors: "+err.Error())
		return
	}
	for key, values := range selectors {
		if key != "language" && key != "profile" && key != "version" {
			s.writeErrorResponse(w, http.StatusBadRequest, "unknown capability selector: "+key)
			return
		}
		if len(values) != 1 {
			s.writeErrorResponse(w, http.StatusBadRequest, "duplicate capability selector: "+key)
			return
		}
	}
	manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: selectors.Get("language"), Profile: selectors.Get("profile"), Version: selectors.Get("version")})
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSONResponse(w, http.StatusOK, manifest)
}

func parseAnalysisDocument(w http.ResponseWriter, r *http.Request) (analysis.QueryDocument, error) {
	body, err := readValidationBody(w, r)
	if err != nil {
		return analysis.QueryDocument{}, err
	}
	// Reuse the canonical document decoder, retaining its strict member and
	// Unicode policy. The HTTP route still accepts exactly one object.
	documents, err := validation.DecodeDocuments(append(append([]byte{'['}, body...), ']'))
	if err != nil {
		return analysis.QueryDocument{}, err
	}
	if len(documents) != 1 {
		return analysis.QueryDocument{}, fmt.Errorf("expected exactly one query document")
	}
	return documents[0], nil
}

// AnalysisRequest is documentation-only. Runtime decoding uses the strict
// canonical document decoder; the reconciler supplies selector constraints.
type AnalysisRequest analysis.QueryDocument
