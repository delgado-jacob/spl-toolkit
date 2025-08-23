package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/delgado-jacob/spl-toolkit/docs"
	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

// handleHealth returns the health status of the service
// @Summary Get service health status
// @Description Check if the SPL Toolkit API service is healthy and responsive
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Router /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "healthy",
		Version: s.version,
		Service: "spl-toolkit-api",
	}
	s.writeJSONResponse(w, http.StatusOK, response)
}

// handleMapQuery handles field mapping requests
// @Summary Map fields in an SPL query
// @Description Apply field mappings to transform field names in an SPL query. Supports both stateless operation (provide mappings/config in request) and stateful operation (use pre-loaded mappings). For production use, provide mappings or config in each request to avoid global state.
// @Tags query
// @Accept json
// @Produce json
// @Param request body MapQueryRequest true "Query mapping request"
// @Success 200 {object} MapQueryResponse "Successfully mapped the query"
// @Failure 400 {object} ValidationErrorResponse "Invalid request or validation errors"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /query/map [post]
func (s *Server) handleMapQuery(w http.ResponseWriter, r *http.Request) {
	var req MapQueryRequest
	if err := parseJSONRequest(w, r, &req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if validationErrors := validateMapQueryRequest(&req); len(validationErrors) > 0 {
		response := ValidationErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Code:    http.StatusBadRequest,
			Errors:  validationErrors,
		}
		s.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Determine which mapper to use: stateless (from request) or stateful (global)
	var currentMapper *mapper.Mapper

	if req.Config != nil || len(req.Mappings) > 0 {
		// Stateless mode: use mappings/config from request with caching
		mappingsHash := s.computeMappingsHash(req.Mappings, req.Config)
		currentMapper = s.getCachedMapper(mappingsHash, req.Mappings, req.Config)
	} else {
		// Stateful mode: use global mapper (fallback for backward compatibility)
		currentMapper = s.mapper.Load()
	}

	// Map the query
	var mappedQuery string
	var err error

	if req.Context != nil {
		mappedQuery, err = currentMapper.MapQueryWithContext(req.Query, req.Context)
	} else {
		mappedQuery, err = currentMapper.MapQuery(req.Query)
	}

	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to map query: "+err.Error())
		return
	}

	response := MapQueryResponse{
		OriginalQuery: req.Query,
		MappedQuery:   mappedQuery,
		Success:       true,
	}
	s.writeJSONResponse(w, http.StatusOK, response)
}

// handleDiscoverQuery handles query discovery requests
// @Summary Discover information about an SPL query
// @Description Analyze an SPL query to discover metadata like datamodels, lookups, macros, sources, and input fields
// @Tags query
// @Accept json
// @Produce json
// @Param request body DiscoverQueryRequest true "Query discovery request"
// @Success 200 {object} DiscoverQueryResponse "Successfully discovered query information"
// @Failure 400 {object} ValidationErrorResponse "Invalid request or validation errors"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /query/discover [post]
func (s *Server) handleDiscoverQuery(w http.ResponseWriter, r *http.Request) {
	var req DiscoverQueryRequest
	if err := parseJSONRequest(w, r, &req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if validationErrors := validateDiscoverQueryRequest(&req); len(validationErrors) > 0 {
		response := ValidationErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Code:    http.StatusBadRequest,
			Errors:  validationErrors,
		}
		s.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Discover query information
	m := s.mapper.Load()
	queryInfo, err := m.DiscoverQuery(req.Query)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to discover query: "+err.Error())
		return
	}

	response := DiscoverQueryResponse{
		Query:     req.Query,
		QueryInfo: queryInfo,
		Success:   true,
	}
	s.writeJSONResponse(w, http.StatusOK, response)
}

// handleValidateQuery handles query validation requests
// @Summary Validate an SPL query
// @Description Check if an SPL query has valid syntax and can be parsed correctly
// @Tags query
// @Accept json
// @Produce json
// @Param request body ValidateQueryRequest true "Query validation request"
// @Success 200 {object} ValidateQueryResponse "Query is valid"
// @Failure 400 {object} ValidationErrorResponse "Invalid request structure"
// @Failure 422 {object} ValidateQueryResponse "Query has invalid syntax"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /query/validate [post]
func (s *Server) handleValidateQuery(w http.ResponseWriter, r *http.Request) {
	var req ValidateQueryRequest
	if err := parseJSONRequest(w, r, &req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request structure
	if validationErrors := validateValidateQueryRequest(&req); len(validationErrors) > 0 {
		response := ValidationErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Code:    http.StatusBadRequest,
			Errors:  validationErrors,
		}
		s.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Validate the SPL query
	m := s.mapper.Load()
	err := m.ValidateQuery(req.Query)

	if err != nil {
		// Return 422 for invalid queries with structured error response
		response := ValidateQueryResponse{
			Query:   req.Query,
			Valid:   false,
			Success: false,
			Error:   err.Error(),
		}
		s.writeJSONResponse(w, http.StatusUnprocessableEntity, response)
		return
	}

	// Return 200 for valid queries
	response := ValidateQueryResponse{
		Query:   req.Query,
		Valid:   true,
		Success: true,
	}
	s.writeJSONResponse(w, http.StatusOK, response)
}

// handleLoadMappings handles loading field mappings (admin-only endpoint)
// @Summary Load field mappings into the server (ADMIN ONLY - DEV USE)
// @Description **WARNING: This is an ephemeral, process-global, development-only endpoint.** Loads field mappings or mapping configuration globally for all subsequent requests. Not suitable for production multi-user environments. Use the mappings/config parameter in /query/map instead.
// @Tags mappings
// @Accept json
// @Produce json
// @Param request body LoadMappingsRequest true "Mappings loading request"
// @Success 200 {object} LoadMappingsResponse "Successfully loaded mappings"
// @Failure 400 {object} ValidationErrorResponse "Invalid request or validation errors"
// @Failure 404 {object} ErrorResponse "Admin endpoints disabled"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /mappings [post]
func (s *Server) handleLoadMappings(w http.ResponseWriter, r *http.Request) {
	// Check if admin endpoints are enabled
	if !s.config.EnableAdminEndpoints {
		s.writeErrorResponse(w, http.StatusNotFound, "endpoint not found")
		return
	}
	var req LoadMappingsRequest
	if err := parseJSONRequest(w, r, &req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if validationErrors := validateLoadMappingsRequest(&req); len(validationErrors) > 0 {
		response := ValidationErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Code:    http.StatusBadRequest,
			Errors:  validationErrors,
		}
		s.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	var mappingsCount, rulesCount int

	// Handle config-based mappings
	if req.Config != nil {
		newMapper := mapper.NewWithConfig(req.Config)
		s.mapper.Store(newMapper)
		mappingsCount = len(req.Config.Mappings)
		rulesCount = len(req.Config.Rules)
		log.Printf("Loaded mapping config with %d basic mappings and %d conditional rules", mappingsCount, rulesCount)
	} else {
		// Handle simple mappings array
		mappingsJSON, err := json.Marshal(req.Mappings)
		if err != nil {
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to process mappings")
			return
		}

		newMapper := mapper.New()
		if err := newMapper.LoadMappings(mappingsJSON); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Failed to load mappings: "+err.Error())
			return
		}
		s.mapper.Store(newMapper)
		mappingsCount = len(req.Mappings)
		log.Printf("Loaded %d field mappings", mappingsCount)
	}

	response := LoadMappingsResponse{
		Success:       true,
		MappingsCount: mappingsCount,
		RulesCount:    rulesCount,
	}
	s.writeJSONResponse(w, http.StatusOK, response)
}

// handleDocs serves the Swagger UI documentation with SRI protection
// @Summary Get API documentation
// @Description Serve interactive Swagger UI documentation for the SPL Toolkit API
// @Tags docs
// @Produce text/html
// @Success 200 {string} string "HTML page with Swagger UI"
// @Router /docs [get]
func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	// Swagger UI without SRI for now (can cause loading issues)
	html := `<!DOCTYPE html>
<html>
<head>
    <title>SPL Toolkit API Documentation</title>
    <link rel="stylesheet" type="text/css" 
          href="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui.css" 
          crossorigin="anonymous" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #1976d2;
        }
        .swagger-ui .topbar .download-url-wrapper {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-bundle.js" 
            crossorigin="anonymous"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-standalone-preset.js" 
            crossorigin="anonymous"></script>
    <script>
        window.onload = function() {
            try {
                const ui = SwaggerUIBundle({
                    url: '/api/v1/openapi.json',
                    dom_id: '#swagger-ui',
                    deepLinking: true,
                    presets: [
                        SwaggerUIBundle.presets.apis,
                        SwaggerUIStandalonePreset
                    ],
                    plugins: [
                        SwaggerUIBundle.plugins.DownloadUrl
                    ],
                    layout: "StandaloneLayout",
                    docExpansion: "none",
                    operationsSorter: "alpha"
                });
            } catch (error) {
                console.error('Failed to initialize Swagger UI:', error);
                document.getElementById('swagger-ui').innerHTML = 
                    '<div style="padding: 20px; color: red;">Failed to load API documentation. Please check console for errors.</div>';
            }
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// handleOpenAPISpec serves the OpenAPI JSON specification
// @Summary Get OpenAPI specification
// @Description Get the OpenAPI 3.1 specification for the SPL Toolkit API in JSON format
// @Tags docs
// @Produce json
// @Success 200 {object} map[string]interface{} "OpenAPI 3.1 specification"
// @Router /openapi.json [get]
func (s *Server) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	// Get the generated spec from swag
	spec := docs.SwaggerInfo.ReadDoc()

	// Parse JSON to set dynamic values
	var specMap map[string]interface{}
	if err := json.Unmarshal([]byte(spec), &specMap); err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to parse OpenAPI spec")
		return
	}

	// Update version if available
	if info, ok := specMap["info"].(map[string]interface{}); ok {
		info["version"] = s.version
	}

	s.writeJSONResponse(w, http.StatusOK, specMap)
}
