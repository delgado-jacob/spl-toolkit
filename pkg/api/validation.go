package api

import (
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// handleValidateFields returns the canonical local field validation report.
// @Summary Validate query fields against a local catalog
// @Description Strict object with required document and catalog keys. Document requires string text and permits only language/profile/version/source_id string options. Catalog accepts a string array or an object requiring fields and permitting optional_fields/identity/version. Nulls, unknown or duplicate keys, malformed Unicode, duplicate names and ordinary/optional overlap are rejected. Catalog metadata never resolves files or URLs. Body limit is 1 MiB. Content statuses valid/invalid/incomplete all return 200 with the full canonical report.
// @Tags query
// @Accept json
// @Produce json
// @Param request body validation.Request true "Query document and catalog (catalog may also be a string array)"
// @Success 200 {object} validation.Report "Canonical report, including invalid or incomplete content"
// @Failure 400 {object} ErrorResponse "Input error, content type, or body limit"
// @Failure 500 {object} ErrorResponse "Unexpected internal error"
// @Router /query/validate-fields [post]
func (s *Server) handleValidateFields(w http.ResponseWriter, r *http.Request) {
	data, err := readValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	request, err := validation.DecodeRequest(data)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	report, err := validation.Validate(request.Document, request.Catalog)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

// handleValidateFieldsBatch preserves document order and individual reports.
// @Summary Validate a batch of query documents against a local catalog
// @Description Strict object requiring only documents and catalog. Documents is a nonempty array of strict query documents, each with its own options and identity. Catalog accepts a string array or an object requiring fields and permitting optional_fields/identity/version; names are case-sensitive. Reject nulls, unknown or duplicate keys and malformed Unicode. No catalog file or URL resolution. Body limit is 1 MiB. Batch status precedence is invalid, incomplete, valid; all content statuses return 200.
// @Tags query
// @Accept json
// @Produce json
// @Param request body validation.BatchRequest true "Nonempty query document array and catalog (catalog may also be a string array)"
// @Success 200 {object} validation.BatchReport "Canonical ordered batch report"
// @Failure 400 {object} ErrorResponse "Input error, content type, or body limit"
// @Failure 500 {object} ErrorResponse "Unexpected internal error"
// @Router /query/validate-fields/batch [post]
func (s *Server) handleValidateFieldsBatch(w http.ResponseWriter, r *http.Request) {
	data, err := readValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	request, err := validation.DecodeBatchRequest(data)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	report, err := validation.ValidateBatch(request.Documents, request.Catalog)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

func readValidationBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "application/json" {
		return nil, fmt.Errorf("content-type must be application/json")
	}
	return io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
}

func (s *Server) writeValidationError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if validation.IsInputError(err) {
		code = http.StatusBadRequest
	}
	s.writeErrorResponse(w, code, err.Error())
}
