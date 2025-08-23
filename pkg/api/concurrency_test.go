package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func TestConcurrentMapperAccess(t *testing.T) {
	// Enable admin endpoints for this test
	config := ServerConfig{
		CORSAllowedOrigins:   []string{"*"},
		CORSAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSAllowedHeaders:   []string{"Content-Type", "Authorization"},
		EnableAdminEndpoints: true,
	}
	server := NewServerWithConfig(config)

	// Test concurrent reads while updating mappings
	const numWorkers = 10
	const numRequests = 20

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers*numRequests)

	// Worker that continuously makes map requests
	mapWorker := func() {
		defer wg.Done()
		for i := 0; i < numRequests; i++ {
			req := MapQueryRequest{
				Query: "search src_ip=192.168.1.1",
			}
			jsonData, _ := json.Marshal(req)

			httpReq, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer(jsonData))
			if err != nil {
				errors <- err
				continue
			}
			httpReq.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, httpReq)

			if rr.Code != http.StatusOK {
				errors <- fmt.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
			}
		}
	}

	// Worker that updates mappings
	mappingWorker := func() {
		defer wg.Done()
		for i := 0; i < numRequests; i++ {
			req := LoadMappingsRequest{
				Mappings: []mapper.FieldMapping{
					{Source: "src_ip", Target: "source_ip"},
					{Source: "dst_ip", Target: "dest_ip"},
				},
			}
			jsonData, _ := json.Marshal(req)

			httpReq, err := http.NewRequest("POST", "/api/v1/mappings", bytes.NewBuffer(jsonData))
			if err != nil {
				errors <- err
				continue
			}
			httpReq.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, httpReq)

			if rr.Code != http.StatusOK {
				errors <- fmt.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
			}
		}
	}

	// Start workers
	for i := 0; i < numWorkers/2; i++ {
		wg.Add(1)
		go mapWorker()
	}
	for i := 0; i < numWorkers/2; i++ {
		wg.Add(1)
		go mappingWorker()
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestContentTypeWithCharset(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name        string
		contentType string
		expectError bool
	}{
		{
			name:        "Valid JSON content type",
			contentType: "application/json",
			expectError: false,
		},
		{
			name:        "JSON with charset",
			contentType: "application/json; charset=utf-8",
			expectError: false,
		},
		{
			name:        "JSON with boundary",
			contentType: "application/json; boundary=something",
			expectError: false,
		},
		{
			name:        "Invalid content type",
			contentType: "text/plain",
			expectError: true,
		},
		{
			name:        "XML content type",
			contentType: "application/xml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := MapQueryRequest{
				Query: "search index=test",
			}
			jsonData, _ := json.Marshal(req)

			httpReq, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			httpReq.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, httpReq)

			if tt.expectError && rr.Code == http.StatusOK {
				t.Errorf("Expected error for content type %s, but got 200", tt.contentType)
			}
			if !tt.expectError && rr.Code != http.StatusOK {
				t.Errorf("Expected success for content type %s, but got %d: %s", tt.contentType, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestUnknownFields(t *testing.T) {
	server := NewServer()

	// Test request with unknown fields
	jsonData := `{
		"query": "search index=test",
		"unknown_field": "should_be_rejected",
		"another_unknown": 123
	}`

	req, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for unknown fields, got %d: %s", rr.Code, rr.Body.String())
	}

	// Check that the error message mentions unknown fields
	if !strings.Contains(rr.Body.String(), "unknown field") && !strings.Contains(rr.Body.String(), "invalid JSON") {
		t.Errorf("Error message should mention unknown fields, got: %s", rr.Body.String())
	}
}

func TestRequestSizeLimits(t *testing.T) {
	server := NewServer()

	// Test with large query (should be rejected due to validation)
	largeQuery := strings.Repeat("search field=value ", 10000) // Creates a very large query
	req := MapQueryRequest{
		Query: largeQuery,
	}
	jsonData, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", "/api/v1/query/map", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, httpReq)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for oversized query, got %d", rr.Code)
	}
}

func TestCORSPreflightRequests(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("OPTIONS", "/api/v1/query/map", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for OPTIONS request, got %d", rr.Code)
	}

	// Check CORS headers are present
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header")
	}
	if rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Expected Access-Control-Allow-Headers header")
	}
}

func TestRequestIDGeneration(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	requestID := rr.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set")
	}

	// Test that existing request ID is preserved
	req2, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req2.Header.Set("X-Request-ID", "existing-id")

	rr2 := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr2, req2)

	if rr2.Header().Get("X-Request-ID") != "existing-id" {
		t.Errorf("Expected existing request ID to be preserved, got %s", rr2.Header().Get("X-Request-ID"))
	}
}

func TestSecurityHeaders(t *testing.T) {
	server := NewServer()

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	expectedHeaders := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"X-XSS-Protection",
		"Content-Security-Policy",
	}

	for _, header := range expectedHeaders {
		if rr.Header().Get(header) == "" {
			t.Errorf("Expected security header %s to be set", header)
		}
	}
}

func TestValidationEndpointStatusCodes(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name         string
		query        string
		expectedCode int
	}{
		{
			name:         "Valid query returns 200",
			query:        "search index=test",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid query returns 422",
			query:        "search | | | invalid",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Empty query returns 400",
			query:        "",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := ValidateQueryRequest{
				Query: tt.query,
			}
			jsonData, _ := json.Marshal(req)

			httpReq, err := http.NewRequest("POST", "/api/v1/query/validate", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			httpReq.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, httpReq)

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d: %s", tt.expectedCode, rr.Code, rr.Body.String())
			}
		})
	}
}
