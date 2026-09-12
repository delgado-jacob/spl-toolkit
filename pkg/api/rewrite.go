package api

import (
	"encoding/json"
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// handleRewrite returns one complete canonical rewrite report.
// @Summary Preview or apply explicit safe rewrite rules
// @Description Strict versioned request with one query document, explicit inline rules, optional inline validation target, and preview mode by default. No filesystem paths, URL retrieval, or mapper configuration. Query statuses valid, invalid, and incomplete all return HTTP 200. Body limit is 8 MiB.
// @Tags query
// @Accept json
// @Produce json
// @Param request body RewriteRequest true "Versioned single-query rewrite request"
// @Success 200 {object} rewrite.Result "Canonical rewrite report"
// @Failure 400 {object} ErrorResponse "Input error, content type, or body limit"
// @Failure 500 {object} ErrorResponse "Unexpected internal error"
// @Router /query/rewrite [post]
func (s *Server) handleRewrite(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	request, err := rewrite.DecodeRequest(data)
	if err != nil {
		s.writeRewriteError(w, err)
		return
	}
	report, err := rewrite.Rewrite(request)
	if err != nil {
		s.writeRewriteError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

// handleRewriteBatch returns ordered complete canonical rewrite reports.
// @Summary Preview or apply explicit safe rewrite rules to an ordered batch
// @Description Strict versioned request with nonempty query documents, explicit inline rules, optional inline validation target, and preview mode by default. No filesystem paths, URL retrieval, or mapper configuration. Query statuses valid, invalid, and incomplete all return HTTP 200. Body limit is 8 MiB.
// @Tags query
// @Accept json
// @Produce json
// @Param request body RewriteBatchRequest true "Versioned ordered-batch rewrite request"
// @Success 200 {object} rewrite.BatchResult "Canonical ordered rewrite reports"
// @Failure 400 {object} ErrorResponse "Input error, content type, or body limit"
// @Failure 500 {object} ErrorResponse "Unexpected internal error"
// @Router /query/rewrite/batch [post]
func (s *Server) handleRewriteBatch(w http.ResponseWriter, r *http.Request) {
	data, err := readSchemaValidationBody(w, r)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	request, err := rewrite.DecodeBatchRequest(data)
	if err != nil {
		s.writeRewriteError(w, err)
		return
	}
	report, err := rewrite.RewriteBatch(request)
	if err != nil {
		s.writeRewriteError(w, err)
		return
	}
	s.writeJSONResponse(w, http.StatusOK, report)
}

func (s *Server) writeRewriteError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if rewrite.IsInputError(err) || validation.IsInputError(err) {
		code = http.StatusBadRequest
	}
	s.writeErrorResponse(w, code, err.Error())
}

// RewriteRequest is documentation-only. Runtime decoding is canonical and strict.
type RewriteRequest struct {
	SchemaVersion    int                    `json:"schema_version"`
	Mode             rewrite.Mode           `json:"mode,omitempty"`
	Document         analysis.QueryDocument `json:"document"`
	Rules            []RewriteRule          `json:"rules"`
	ValidationTarget json.RawMessage        `json:"validation_target,omitempty" swaggertype:"object"`
}

// RewriteBatchRequest is the documentation-only ordered batch shape.
type RewriteBatchRequest struct {
	SchemaVersion    int                      `json:"schema_version"`
	Mode             rewrite.Mode             `json:"mode,omitempty"`
	Documents        []analysis.QueryDocument `json:"documents"`
	Rules            []RewriteRule            `json:"rules"`
	ValidationTarget json.RawMessage          `json:"validation_target,omitempty" swaggertype:"object"`
}

// RewriteIdentity documents the exact atom-or-path union.
type RewriteIdentity struct {
	Name string   `json:"name,omitempty"`
	Path []string `json:"path,omitempty"`
}

// RewriteCondition documents recursive combinators and canonical facts.
type RewriteCondition struct {
	All      []RewriteCondition `json:"all,omitempty"`
	Any      []RewriteCondition `json:"any,omitempty"`
	Fact     string             `json:"fact,omitempty"`
	Kind     string             `json:"kind,omitempty"`
	Identity *RewriteIdentity   `json:"identity,omitempty"`
	Operator string             `json:"operator,omitempty"`
	Value    any                `json:"value,omitempty"`
}

// RewriteRule documents one explicit source-to-target identity rule.
type RewriteRule struct {
	ID     string            `json:"id"`
	Kind   string            `json:"kind"`
	Source RewriteIdentity   `json:"source"`
	Target RewriteIdentity   `json:"target"`
	When   *RewriteCondition `json:"when,omitempty"`
}
