package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func TestHealthEndpoint(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Service != "spl-toolkit-api" {
		t.Errorf("Expected service 'spl-toolkit-api', got '%s'", response.Service)
	}
}

func TestMapQueryEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		request        MapQueryRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid query with mappings",
			request: MapQueryRequest{
				Query: "search src_ip=192.168.1.1",
				Mappings: []mapper.FieldMapping{
					{Source: "src_ip", Target: "source_ip"},
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Valid query without mappings",
			request: MapQueryRequest{
				Query: "search index=web | stats count",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Empty query",
			request: MapQueryRequest{
				Query: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Query with invalid mappings",
			request: MapQueryRequest{
				Query: "search test=1",
				Mappings: []mapper.FieldMapping{
					{Source: "", Target: "target"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			req, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				t.Errorf("response body: %s", rr.Body.String())
			}

			if tt.expectError {
				var errorResp ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err == nil {
					if !errorResp.Error {
						t.Errorf("Expected error response")
					}
				} else {
					// Try validation error response
					var validationResp ValidationErrorResponse
					if err := json.Unmarshal(rr.Body.Bytes(), &validationResp); err != nil {
						t.Errorf("Failed to parse error response: %v", err)
					}
					if !validationResp.Error {
						t.Errorf("Expected error response")
					}
				}
			} else {
				var response MapQueryResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if !response.Success {
					t.Errorf("Expected successful response")
				}
			}
		})
	}
}

func TestDiscoverQueryEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		request        DiscoverQueryRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid query",
			request: DiscoverQueryRequest{
				Query: "search sourcetype=access_combined | stats count by src_ip",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Query with datamodel",
			request: DiscoverQueryRequest{
				Query: "| datamodel Network_Traffic All_Traffic search | stats count",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Empty query",
			request: DiscoverQueryRequest{
				Query: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			req, err := http.NewRequest("POST", "/api/v1/query/discover", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				t.Errorf("response body: %s", rr.Body.String())
			}

			if tt.expectError {
				var errorResp ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err == nil {
					if !errorResp.Error {
						t.Errorf("Expected error response")
					}
				} else {
					// Try validation error response
					var validationResp ValidationErrorResponse
					if err := json.Unmarshal(rr.Body.Bytes(), &validationResp); err != nil {
						t.Errorf("Failed to parse error response: %v", err)
					}
				}
			} else {
				var response DiscoverQueryResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if !response.Success {
					t.Errorf("Expected successful response")
				}
				if response.QueryInfo == nil {
					t.Errorf("Expected query info in response")
				}
			}
		})
	}
}

func TestValidateQueryEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		request        ValidateQueryRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid query",
			request: ValidateQueryRequest{
				Query: "search index=web | stats count",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Empty query",
			request: ValidateQueryRequest{
				Query: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			req, err := http.NewRequest("POST", "/api/v1/query/validate", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				t.Errorf("response body: %s", rr.Body.String())
			}

			if tt.expectError {
				var validationResp ValidationErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &validationResp); err != nil {
					t.Errorf("Failed to parse validation error response: %v", err)
				}
				if !validationResp.Error {
					t.Errorf("Expected error response")
				}
			} else {
				var response ValidateQueryResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if !response.Success {
					t.Errorf("Expected successful response")
				}
			}
		})
	}
}

func TestLoadMappingsEndpoint(t *testing.T) {
	err := os.Setenv("ENABLE_ADMIN_ENDPOINTS", "true")
	defer os.Setenv("ENABLE_ADMIN_ENDPOINTS", "false")
	server := NewServer()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name           string
		request        LoadMappingsRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Load simple mappings",
			request: LoadMappingsRequest{
				Mappings: []mapper.FieldMapping{
					{Source: "src_ip", Target: "source_ip"},
					{Source: "dst_ip", Target: "dest_ip"},
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Load mapping config",
			request: LoadMappingsRequest{
				Config: &mapper.MappingConfig{
					Version: "1.0",
					Name:    "Test Config",
					Mappings: []mapper.FieldMapping{
						{Source: "src_port", Target: "source_port"},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Empty request",
			request: LoadMappingsRequest{
				Mappings: []mapper.FieldMapping{},
				Config:   nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Invalid mappings",
			request: LoadMappingsRequest{
				Mappings: []mapper.FieldMapping{
					{Source: "", Target: "target"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			req, err := http.NewRequest("POST", "/api/v1/mappings", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				t.Errorf("response body: %s", rr.Body.String())
			}

			if tt.expectError {
				var validationResp ValidationErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &validationResp); err != nil {
					t.Errorf("Failed to parse validation error response: %v", err)
				}
				if !validationResp.Error {
					t.Errorf("Expected error response")
				}
			} else {
				var response LoadMappingsResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if !response.Success {
					t.Errorf("Expected successful response")
				}
			}
		})
	}
}

func TestInvalidContentType(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer([]byte("invalid")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
		t.Errorf("Failed to parse error response: %v", err)
	}
	if !errorResp.Error {
		t.Errorf("Expected error response")
	}
}

func TestInvalidJSON(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
		t.Errorf("Failed to parse error response: %v", err)
	}
	if !errorResp.Error {
		t.Errorf("Expected error response")
	}
}

func TestOpenAPISpecEndpoint(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("GET", "/api/v1/openapi.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Errorf("Failed to parse OpenAPI spec: %v", err)
	}

	if spec["openapi"] != "3.1.0" {
		t.Errorf("Expected OpenAPI version '3.1.0', got '%v'", spec["openapi"])
	}

	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected info object in OpenAPI spec")
	}
	if info["title"] != "SPL Toolkit API" {
		t.Errorf("Expected title 'SPL Toolkit API', got '%v'", info["title"])
	}
}

func TestDocsEndpoint(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("GET", "/api/v1/docs", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected content-type 'text/html; charset=utf-8', got '%s'", contentType)
	}

	if !bytes.Contains(rr.Body.Bytes(), []byte("swagger-ui")) {
		t.Errorf("Expected Swagger UI in response body")
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("/api/v1/openapi.json")) {
		t.Errorf("Expected OpenAPI spec URL in response body")
	}
}

func TestAdminEndpointsDisabledByDefault(t *testing.T) {
	server := NewServer() // Admin endpoints disabled by default

	req, err := http.NewRequest("POST", "/api/v1/mappings", bytes.NewBuffer([]byte(`{"mappings":[{"source":"test","target":"test2"}]}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected 404 when admin endpoints disabled, got %d", status)
	}
}

func TestAdminEndpointsWhenEnabled(t *testing.T) {
	config := ServerConfig{
		CORSAllowedOrigins:   []string{"*"},
		CORSAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSAllowedHeaders:   []string{"Content-Type", "Authorization"},
		EnableAdminEndpoints: true,
	}
	server := NewServerWithConfig(config)

	req, err := http.NewRequest("POST", "/api/v1/mappings", bytes.NewBuffer([]byte(`{"mappings":[{"source":"test","target":"test2"}]}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200 when admin endpoints enabled, got %d: %s", status, rr.Body.String())
	}
}

func TestStatelessMapping(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "Stateless with simple mappings",
			requestBody: `{
				"query": "search src_ip=192.168.1.1",
				"mappings": [{"source": "src_ip", "target": "source_ip"}]
			}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Stateless with config",
			requestBody: `{
				"query": "search src_ip=192.168.1.1",
				"config": {
					"version": "1.0",
					"name": "Test Config",
					"mappings": [{"source": "src_ip", "target": "source_ip"}]
				}
			}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Stateful fallback (no mappings/config)",
			requestBody: `{
				"query": "search src_ip=192.168.1.1"
			}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer([]byte(tt.requestBody)))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d: %s", tt.expectedStatus, status, rr.Body.String())
			}

			if tt.expectedStatus == http.StatusOK {
				var response MapQueryResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if !response.Success {
					t.Errorf("Expected successful response")
				}
			}
		})
	}
}

func TestMapperCaching(t *testing.T) {
	server := NewServer()

	// Make two identical requests to test caching
	requestBody := `{
		"query": "search src_ip=192.168.1.1",
		"mappings": [{"source": "src_ip", "target": "source_ip"}]
	}`

	// First request
	req1, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatal(err)
	}
	req1.Header.Set("Content-Type", "application/json")

	rr1 := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr1, req1)

	if status := rr1.Code; status != http.StatusOK {
		t.Errorf("First request failed with status %d: %s", status, rr1.Body.String())
	}

	// Second request (should use cached mapper)
	req2, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatal(err)
	}
	req2.Header.Set("Content-Type", "application/json")

	rr2 := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr2, req2)

	if status := rr2.Code; status != http.StatusOK {
		t.Errorf("Second request failed with status %d: %s", status, rr2.Body.String())
	}

	// Both responses should be identical
	if rr1.Body.String() != rr2.Body.String() {
		t.Errorf("Cached response differs from original")
	}
}
