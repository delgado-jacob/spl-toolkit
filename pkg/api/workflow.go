package api

import (
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"github.com/delgado-jacob/spl-toolkit/pkg/workflow"
)

// handleWorkflowAssess selects a rendering of canonical offline assessment.
// @Summary Assess explicitly selected inline detections against offline evidence
// @Description Strict inline JSON only; no path or URI acquisition. Body limit 8 MiB. Content and retained entry failures return 200. Format selects text, json (default), sarif, graph or bom without changing canonical assessment or CI outcome.
// @Tags workflow
// @Accept json
// @Produce json,plain
// @Param request body workflow.Request true "Inline documents and settings"
// @Success 200 {object} workflow.Report "Selected canonical assessment rendering"
// @Failure 400 {object} workflow.RequestErrorDetail "Shared input error; transport boundary errors use ErrorResponse"
// @Failure 500 {object} ErrorResponse "Unexpected internal error"
// @Router /workflow/assess [post]
func (s *Server) handleWorkflowAssess(w http.ResponseWriter, r *http.Request) {
	s.handleWorkflow(w, r, func(raw []byte) (any, error) { return workflow.AssessOutputJSON(raw) })
}

// handleWorkflowCompare compares saved evidence without reevaluation.
// @Summary Compare two saved workflow reports
// @Description Strict inline JSON only; body limit 8 MiB. Content and retained entry failures return 200.
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.CompareRequest true "Before and after saved reports"
// @Success 200 {object} workflow.ComparisonReport
// @Failure 400 {object} workflow.RequestErrorDetail
// @Failure 500 {object} ErrorResponse
// @Router /workflow/compare [post]
func (s *Server) handleWorkflowCompare(w http.ResponseWriter, r *http.Request) {
	s.handleWorkflow(w, r, func(raw []byte) (any, error) { return workflow.CompareJSON(raw) })
}

// handleWorkflowEvidence projects saved facts using explicit disclosure choices.
// @Summary Project minimal evidence from one saved report or comparison
// @Description Strict inline JSON only; body limit 8 MiB. Include must be an explicit array. Projection never executes queries or restores proof.
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.EvidenceRequest true "Exactly one saved evidence source and explicit include array"
// @Success 200 {object} workflow.EvidenceReport
// @Failure 400 {object} workflow.RequestErrorDetail
// @Failure 500 {object} ErrorResponse
// @Router /workflow/evidence [post]
func (s *Server) handleWorkflowEvidence(w http.ResponseWriter, r *http.Request) {
	s.handleWorkflow(w, r, func(raw []byte) (any, error) { return workflow.EvidenceJSON(raw) })
}

// handleWorkflowRecheck freshly evaluates a proposed change against inline context.
// @Summary Recheck a proposal against offline context
// @Description Strict inline JSON only; body limit 8 MiB. Content and retained entry failures return 200. No filesystem reads or URI dereference.
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.RecheckRequest true "Original context and proposal"
// @Success 200 {object} workflow.RecheckReport
// @Failure 400 {object} workflow.RequestErrorDetail
// @Failure 500 {object} ErrorResponse
// @Router /workflow/recheck [post]
func (s *Server) handleWorkflowRecheck(w http.ResponseWriter, r *http.Request) {
	s.handleWorkflow(w, r, func(raw []byte) (any, error) { return workflow.RecheckJSON(raw) })
}

func (s *Server) handleWorkflow(w http.ResponseWriter, r *http.Request, operation func([]byte) (any, error)) {
	w.Header().Set("Content-Type", "application/json")
	raw, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	value, err := operation(raw)
	if err != nil {
		if detail, ok := workflow.RequestErrorDetails(err); ok {
			s.writeJSONResponse(w, http.StatusBadRequest, detail)
			return
		}
		status := http.StatusInternalServerError
		if validation.IsInputError(err) {
			status = http.StatusBadRequest
		}
		s.writeErrorResponse(w, status, err.Error())
		return
	}
	if text, ok := value.(string); ok {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(text))
		return
	}
	s.writeJSONResponse(w, http.StatusOK, value)
}
