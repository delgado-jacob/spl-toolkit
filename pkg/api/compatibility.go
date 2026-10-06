package api

import (
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// handleCompatibility returns canonical content reports independently of outcome.
// @Summary Assess requirements against bound offline evidence
// @Description Accepts one inline compatibility request within the 8 MiB application/json boundary. All content outcomes return 200; request and configuration failures return structured 400 details.
// @Tags query
// @Accept json
// @Produce json
// @Param request body compatibility.Request true "Offline compatibility request"
// @Success 200 {object} compatibility.Report "Canonical assessment report"
// @Failure 400 {object} compatibility.RequestErrorDetail "Request/configuration error; body boundaries use ErrorResponse"
// @Failure 500 {object} ErrorResponse "Internal failure"
// @Router /query/compatibility [post]
func (s *Server) handleCompatibility(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := compatibility.CheckJSON(data)
	if err != nil {
		s.writeCompatibilityError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

func (s *Server) writeCompatibilityError(w http.ResponseWriter, err error) {
	if detail, ok := compatibility.RequestErrorDetails(err); ok {
		s.writeJSONResponse(w, http.StatusBadRequest, detail)
		return
	}
	status := http.StatusInternalServerError
	if validation.IsInputError(err) {
		status = http.StatusBadRequest
	}
	s.writeErrorResponse(w, status, err.Error())
}
