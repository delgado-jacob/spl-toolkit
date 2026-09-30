package api

import (
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

// handleClosure returns the canonical closure report for any content status.
// @Summary Evaluate a caller-supplied knowledge-object closure
// @Tags query
// @Accept json
// @Produce json
// @Success 200 {object} closure.Report "Canonical closure report"
// @Failure 400 {object} ErrorResponse "Malformed input"
// @Router /query/closure [post]
func (s *Server) handleClosure(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	request, err := closure.DecodeRequest(data)
	if err != nil {
		s.writeClosureError(w, err)
		return
	}
	report, err := closure.Evaluate(request)
	if err != nil {
		s.writeClosureError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

func (s *Server) writeClosureError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if closure.IsInputError(err) {
		code = http.StatusBadRequest
	}
	s.writeErrorResponse(w, code, err.Error())
}
