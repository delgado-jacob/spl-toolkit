package api

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

// Request/Response models for API endpoints

// MapQueryRequest represents a request to map fields in a query
// @Description Request to map fields in an SPL query. For stateless operation, provide either 'mappings' for simple field mappings or 'config' for advanced mapping configuration with conditional rules.
type MapQueryRequest struct {
	Query    string                 `json:"query" validate:"required" example:"search src_ip=192.168.1.1" extensions:"x-order=1"` // SPL query to map
	Context  map[string]interface{} `json:"context,omitempty" extensions:"x-order=2"`                                             // Optional context for mapping
	Mappings []mapper.FieldMapping  `json:"mappings,omitempty" extensions:"x-order=3"`                                            // Simple field mappings (for stateless operation)
	Config   *mapper.MappingConfig  `json:"config,omitempty" extensions:"x-order=4"`                                              // Advanced mapping configuration with conditional rules (for stateless operation)
}

// MapQueryResponse represents the response from mapping a query
// @Description Response from mapping an SPL query
type MapQueryResponse struct {
	OriginalQuery string `json:"original_query" example:"search src_ip=192.168.1.1" extensions:"x-order=1"`  // Original input query
	MappedQuery   string `json:"mapped_query" example:"search source_ip=192.168.1.1" extensions:"x-order=2"` // Query with mapped field names
	Success       bool   `json:"success" example:"true" extensions:"x-order=3"`                              // Whether the mapping was successful
}

// DiscoverQueryRequest represents a request to discover query information
// @Description Request to discover information about an SPL query
type DiscoverQueryRequest struct {
	Query string `json:"query" validate:"required" example:"search sourcetype=access_combined | stats count by src_ip"` // SPL query to analyze
}

// DiscoverQueryResponse represents the response from query discovery
// @Description Response from discovering SPL query information
type DiscoverQueryResponse struct {
	Query     string            `json:"query" example:"search sourcetype=access_combined | stats count by src_ip" extensions:"x-order=1"` // Original query
	QueryInfo *mapper.QueryInfo `json:"query_info" extensions:"x-order=2"`                                                                // Discovered query information
	Success   bool              `json:"success" example:"true" extensions:"x-order=3"`                                                    // Whether the discovery was successful
}

// ValidateQueryRequest represents a request to validate a query
// @Description Request to validate an SPL query
type ValidateQueryRequest struct {
	Query string `json:"query" validate:"required" example:"search index=web | stats count"` // SPL query to validate
}

// ValidateQueryResponse represents the response from query validation
// @Description Response from validating an SPL query
type ValidateQueryResponse struct {
	Query   string `json:"query" example:"search index=web | stats count" extensions:"x-order=1"`        // Original query
	Valid   bool   `json:"valid" example:"true" extensions:"x-order=2"`                                  // Whether the query is valid
	Success bool   `json:"success" example:"true" extensions:"x-order=3"`                                // Whether the validation was successful
	Error   string `json:"error,omitempty" example:"syntax error at position 10" extensions:"x-order=4"` // Error message if validation failed
}

// LoadMappingsRequest represents a request to load field mappings
// @Description Request to load field mappings into the server
type LoadMappingsRequest struct {
	Mappings []mapper.FieldMapping `json:"mappings,omitempty" extensions:"x-order=1"` // Simple field mappings
	Config   *mapper.MappingConfig `json:"config,omitempty" extensions:"x-order=2"`   // Complete mapping configuration with rules
}

// LoadMappingsResponse represents the response from loading mappings
// @Description Response from loading field mappings
type LoadMappingsResponse struct {
	Success       bool `json:"success" example:"true" extensions:"x-order=1"`            // Whether the mappings were loaded successfully
	MappingsCount int  `json:"mappings_count" example:"5" extensions:"x-order=2"`        // Number of field mappings loaded
	RulesCount    int  `json:"rules_count,omitempty" example:"2" extensions:"x-order=3"` // Number of conditional rules loaded
}

// HealthResponse represents the health check response
// @Description Health check response
type HealthResponse struct {
	Status  string `json:"status" example:"healthy" extensions:"x-order=1"`          // Service health status
	Version string `json:"version" example:"1.0.0" extensions:"x-order=2"`           // Service version
	Service string `json:"service" example:"spl-toolkit-api" extensions:"x-order=3"` // Service name
}

// ErrorResponse represents an error response
// @Description Standard error response
type ErrorResponse struct {
	Error   bool   `json:"error" example:"true" extensions:"x-order=1"`              // Always true for error responses
	Message string `json:"message" example:"Invalid request" extensions:"x-order=2"` // Human-readable error message
	Code    int    `json:"code" example:"400" extensions:"x-order=3"`                // HTTP status code
}

// ValidationError represents a validation error for specific fields
// @Description Validation error for a specific field
type ValidationError struct {
	Field   string `json:"field" example:"query" extensions:"x-order=1"`               // Field that failed validation
	Message string `json:"message" example:"field is required" extensions:"x-order=2"` // Validation error message
}

// ValidationErrorResponse represents a response with validation errors
// @Description Response with detailed validation errors
type ValidationErrorResponse struct {
	Error   bool              `json:"error" example:"true" extensions:"x-order=1"`                // Always true for error responses
	Message string            `json:"message" example:"Validation failed" extensions:"x-order=2"` // General error message
	Code    int               `json:"code" example:"400" extensions:"x-order=3"`                  // HTTP status code
	Errors  []ValidationError `json:"validation_errors" extensions:"x-order=4"`                   // List of specific validation errors
}

// parseJSONRequest parses and validates a JSON request with proper content-type handling and size limits
func parseJSONRequest(w http.ResponseWriter, r *http.Request, target interface{}) error {
	// Parse and validate Content-Type
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "application/json" {
		return fmt.Errorf("content-type must be application/json")
	}

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit

	// Parse JSON with strict validation
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// validateMapQueryRequest validates a MapQueryRequest
func validateMapQueryRequest(req *MapQueryRequest) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(req.Query) == "" {
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query is required and cannot be empty",
		})
	}

	// Limit query length to prevent excessive processing
	if len(req.Query) > 64*1024 { // 64KB limit
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query is too long (maximum 64KB)",
		})
	}

	// Validate mappings if provided
	for i, mapping := range req.Mappings {
		if strings.TrimSpace(mapping.Source) == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("mappings[%d].source", i),
				Message: "source field is required",
			})
		}
		if strings.TrimSpace(mapping.Target) == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("mappings[%d].target", i),
				Message: "target field is required",
			})
		}
	}

	// Validate config if provided
	if req.Config != nil {
		validationResult := req.Config.Validate()
		if !validationResult.Valid {
			for _, err := range validationResult.Errors {
				errors = append(errors, ValidationError{
					Field:   "config",
					Message: err,
				})
			}
		}
	}

	return errors
}

// validateDiscoverQueryRequest validates a DiscoverQueryRequest
func validateDiscoverQueryRequest(req *DiscoverQueryRequest) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(req.Query) == "" {
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query is required and cannot be empty",
		})
	}

	// Limit query length to prevent excessive processing
	if len(req.Query) > 64*1024 { // 64KB limit
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query is too long (maximum 64KB)",
		})
	}

	return errors
}

// validateValidateQueryRequest validates a ValidateQueryRequest
func validateValidateQueryRequest(req *ValidateQueryRequest) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(req.Query) == "" {
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query is required and cannot be empty",
		})
	}

	// Limit query length to prevent excessive processing
	if len(req.Query) > 64*1024 { // 64KB limit
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query is too long (maximum 64KB)",
		})
	}

	return errors
}

// validateLoadMappingsRequest validates a LoadMappingsRequest
func validateLoadMappingsRequest(req *LoadMappingsRequest) []ValidationError {
	var errors []ValidationError

	// Must have either mappings or config
	if len(req.Mappings) == 0 && req.Config == nil {
		errors = append(errors, ValidationError{
			Field:   "mappings",
			Message: "either mappings or config must be provided",
		})
		errors = append(errors, ValidationError{
			Field:   "config",
			Message: "either mappings or config must be provided",
		})
	}

	// Validate mappings if provided
	for i, mapping := range req.Mappings {
		if strings.TrimSpace(mapping.Source) == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("mappings[%d].source", i),
				Message: "source field is required",
			})
		}
		if strings.TrimSpace(mapping.Target) == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("mappings[%d].target", i),
				Message: "target field is required",
			})
		}
	}

	// Validate config if provided
	if req.Config != nil {
		validationResult := req.Config.Validate()
		if !validationResult.Valid {
			for _, err := range validationResult.Errors {
				errors = append(errors, ValidationError{
					Field:   "config",
					Message: err,
				})
			}
		}
	}

	return errors
}
