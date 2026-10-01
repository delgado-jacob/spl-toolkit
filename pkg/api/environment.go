package api

import (
	"encoding/json"
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// handleValidateEnvironment validates an inline snapshot and optional schema bundle.
// @Summary Validate offline environment artifacts
// @Description Requires schema_version 1 and at least one inline snapshot or schema_bundle. The 8 MiB application/json body is validated without filesystem access. Valid and partial reports return 200; invalid content returns 400 with the canonical report.
// @Tags environment
// @Accept json
// @Produce json
// @Param request body EnvironmentValidationRequest true "Inline environment artifacts"
// @Success 200 {object} environment.Report "Valid or partial canonical report"
// @Failure 400 {object} environment.Report "Invalid artifact or request content; content type and size errors use ErrorResponse"
// @Router /environment/validate [post]
func (s *Server) handleValidateEnvironment(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := environment.ValidateJSON(data)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	status := http.StatusOK
	if report.Status == "invalid" {
		status = http.StatusBadRequest
	}
	s.writeJSONResponse(w, status, report)
}

// EnvironmentValidationRequest is a documentation shape; ValidateJSON handles strict runtime decoding.
type EnvironmentValidationRequest struct {
	SchemaVersion int             `json:"schema_version"`
	Snapshot      json.RawMessage `json:"snapshot,omitempty" swaggertype:"object"`
	SchemaBundle  json.RawMessage `json:"schema_bundle,omitempty" swaggertype:"object"`
}
