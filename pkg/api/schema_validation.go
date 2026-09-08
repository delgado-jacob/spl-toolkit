package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// handleValidateSchema returns the canonical local field validation report.
// @Summary Validate query fields against an offline JSON Schema or OCSF target
// @Description Strict wrapper requiring document (or nonempty documents for batch) and target. Query documents require text. Reject unknown/duplicate wrapper members, nulls and malformed Unicode. Target is a tagged json_schema or ocsf union: JSON Schema accepts object-or-boolean schema, optional base_uri and URI-keyed inline resources; OCSF requires an official compiled catalog object and exact selection version with one of class/class_uid/category/category_uid, optional unique profiles/extensions arrays (not null). Schema/catalog payloads allow their own vocabulary. No filesystem lookup, downloading or URL dereference. Body limit is 8 MiB. Valid, invalid and incomplete content all return 200 with the full canonical report; input errors return 400.
// @Tags query
// @Accept json
// @Produce json
// @Param request body SchemaValidationRequest true "Query document and explicitly supplied schema target"
// @Success 200 {object} validation.SchemaReport "Canonical report, including invalid or incomplete content"
// @Failure 400 {object} ErrorResponse "Input error, content type, or body limit"
// @Failure 500 {object} ErrorResponse "Unexpected internal error"
// @Router /query/validate-schema [post]
func (s *Server) handleValidateSchema(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	request, err := validation.DecodeSchemaRequest(data)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	report, err := validation.ValidateSchema(request.Document, request.Target)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

// handleValidateSchemaBatch preserves document order and individual reports.
// @Summary Validate an ordered batch against an offline JSON Schema or OCSF target
// @Description Strict wrapper requiring document (or nonempty documents for batch) and target. Query documents require text. Reject unknown/duplicate wrapper members, nulls and malformed Unicode. Target is a tagged json_schema or ocsf union: JSON Schema accepts object-or-boolean schema, optional base_uri and URI-keyed inline resources; OCSF requires an official compiled catalog object and exact selection version with one of class/class_uid/category/category_uid, optional unique profiles/extensions arrays (not null). Schema/catalog payloads allow their own vocabulary. No filesystem lookup, downloading or URL dereference. Body limit is 8 MiB. Valid, invalid and incomplete content all return 200 with the full canonical report; input errors return 400.
// @Tags query
// @Accept json
// @Produce json
// @Param request body SchemaValidationBatchRequest true "Nonempty query document array and explicitly supplied schema target"
// @Success 200 {object} validation.SchemaBatchReport "Canonical ordered batch report"
// @Failure 400 {object} ErrorResponse "Input error, content type, or body limit"
// @Failure 500 {object} ErrorResponse "Unexpected internal error"
// @Router /query/validate-schema/batch [post]
func (s *Server) handleValidateSchemaBatch(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	request, err := validation.DecodeSchemaBatchRequest(data)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	report, err := validation.ValidateSchemaBatch(request.Documents, request.Target)
	if err != nil {
		s.writeValidationError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

func readSchemaValidationBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	mediaType, _, parseErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if parseErr != nil || mediaType != "application/json" {
		return nil, fmt.Errorf("content-type must be application/json")
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<20))
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		return nil, fmt.Errorf("request body exceeds 8 MiB limit")
	}
	return data, err
}

// SchemaValidationRequest is a documentation-only shape. Runtime decoding uses
// validation.DecodeSchemaRequest; the reconciler supplies the strict target union.
type SchemaValidationRequest struct {
	Document analysis.QueryDocument `json:"document"`
	Target   SchemaValidationTarget `json:"target"`
}

// SchemaValidationBatchRequest documents the strict ordered batch wrapper.
type SchemaValidationBatchRequest struct {
	Documents []analysis.QueryDocument `json:"documents"`
	Target    SchemaValidationTarget   `json:"target"`
}

// SchemaValidationTarget makes inline JSON visible to pinned Swag without
// treating json.RawMessage as bytes. The reconciler applies union constraints.
type SchemaValidationTarget struct {
	Kind      string                     `json:"kind"`
	Identity  string                     `json:"identity,omitempty"`
	Schema    json.RawMessage            `json:"schema,omitempty" swaggertype:"object"`
	BaseURI   string                     `json:"base_uri,omitempty"`
	Resources map[string]json.RawMessage `json:"resources,omitempty" swaggertype:"object"`
	Catalog   json.RawMessage            `json:"catalog,omitempty" swaggertype:"object"`
	Selection *validation.OCSFSelection  `json:"selection,omitempty"`
}
