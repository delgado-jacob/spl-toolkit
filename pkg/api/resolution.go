package api

import (
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

// handleResolution returns canonical content reports independently of outcome.
// @Summary Resolve named inputs using bound offline evidence
// @Description Accepts one inline resolution request within the 8 MiB application/json boundary. All content outcomes return 200; request and configuration failures return structured 400 details.
// @Tags query
// @Accept json
// @Produce json
// @Param request body resolution.Request true "Offline resolution request"
// @Success 200 {object} resolution.Report "Canonical assessment report"
// @Failure 400 {object} resolution.RequestErrorDetail "Request/configuration error; body boundaries use ErrorResponse"
// @Failure 500 {object} ErrorResponse "Internal failure"
// @Router /query/resolve [post]
func (s *Server) handleResolution(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := resolution.ResolveJSON(data)
	if err != nil {
		s.writeResolutionError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

func (s *Server) writeResolutionError(w http.ResponseWriter, err error) {
	if detail, ok := resolution.RequestErrorDetails(err); ok {
		s.writeJSONResponse(w, http.StatusBadRequest, detail)
		return
	}
	s.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
}
