package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

// mapperCacheEntry represents a cached temporary mapper with expiration
type mapperCacheEntry struct {
	mapper    *mapper.Mapper
	expiresAt time.Time
}

// Server represents the API server
type Server struct {
	mapper  atomic.Pointer[mapper.Mapper]
	mux     *http.ServeMux
	config  ServerConfig
	logger  *slog.Logger
	version string
	// Cache for temporary mappers with TTL
	mapperCache sync.Map // map[string]*mapperCacheEntry
}

// ContextKey type for context keys
type ContextKey string

const (
	RequestIDKey ContextKey = "request-id"
)

// ServerConfig holds configuration for the API server
type ServerConfig struct {
	CORSAllowedOrigins   []string
	CORSAllowedMethods   []string
	CORSAllowedHeaders   []string
	EnableAdminEndpoints bool
}

// NewServer creates a new API server instance with default configuration
func NewServer() *Server {
	config := getServerConfigFromEnv()
	return NewServerWithConfig(config)
}

// NewServerWithVersion creates a new API server instance with version
func NewServerWithVersion(version string) *Server {
	config := getServerConfigFromEnv()
	return NewServerWithConfigAndVersion(config, version)
}

// NewServerWithConfig creates a new API server instance with custom configuration
func NewServerWithConfig(config ServerConfig) *Server {
	return NewServerWithConfigAndVersion(config, "dev")
}

// NewServerWithConfigAndVersion creates a new API server instance with custom configuration and version
func NewServerWithConfigAndVersion(config ServerConfig, version string) *Server {
	// Set up structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	s := &Server{
		mux:     http.NewServeMux(),
		config:  config,
		logger:  logger,
		version: version,
	}
	m := mapper.New()
	s.mapper.Store(m)
	s.setupRoutes()
	return s
}

// getServerConfigFromEnv loads server configuration from environment variables
func getServerConfigFromEnv() ServerConfig {
	config := ServerConfig{
		// Default CORS settings for development
		CORSAllowedOrigins: []string{"*"},
		CORSAllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSAllowedHeaders: []string{"Content-Type", "Authorization"},
		// Admin endpoints disabled by default for production safety
		EnableAdminEndpoints: false,
	}

	// Override with environment variables if set
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		config.CORSAllowedOrigins = strings.Split(origins, ",")
		// Trim whitespace from origins
		for i, origin := range config.CORSAllowedOrigins {
			config.CORSAllowedOrigins[i] = strings.TrimSpace(origin)
		}
	}

	if methods := os.Getenv("CORS_ALLOWED_METHODS"); methods != "" {
		config.CORSAllowedMethods = strings.Split(methods, ",")
		for i, method := range config.CORSAllowedMethods {
			config.CORSAllowedMethods[i] = strings.TrimSpace(method)
		}
	}

	if headers := os.Getenv("CORS_ALLOWED_HEADERS"); headers != "" {
		config.CORSAllowedHeaders = strings.Split(headers, ",")
		for i, header := range config.CORSAllowedHeaders {
			config.CORSAllowedHeaders[i] = strings.TrimSpace(header)
		}
	}

	// Enable admin endpoints if explicitly set
	if adminFlag := os.Getenv("ENABLE_ADMIN_ENDPOINTS"); adminFlag == "true" {
		config.EnableAdminEndpoints = true
	}

	return config
}

// getCachedMapper retrieves a cached temporary mapper or creates one if not found/expired
func (s *Server) getCachedMapper(mappingsHash string, mappings []mapper.FieldMapping, config *mapper.MappingConfig) *mapper.Mapper {
	if entry, ok := s.mapperCache.Load(mappingsHash); ok {
		cacheEntry := entry.(*mapperCacheEntry)
		if time.Now().Before(cacheEntry.expiresAt) {
			return cacheEntry.mapper
		}
		// Expired, remove from cache
		s.mapperCache.Delete(mappingsHash)
	}

	// Create new temporary mapper
	var tempMapper *mapper.Mapper
	if config != nil {
		tempMapper = mapper.NewWithConfig(config)
	} else {
		tempMapper = mapper.New()
		if len(mappings) > 0 {
			mappingsJSON, _ := json.Marshal(mappings)
			tempMapper.LoadMappings(mappingsJSON)
		}
	}

	// Cache with 5-minute TTL
	cacheEntry := &mapperCacheEntry{
		mapper:    tempMapper,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	s.mapperCache.Store(mappingsHash, cacheEntry)

	return tempMapper
}

// computeMappingsHash creates a hash of mappings/config for cache key
func (s *Server) computeMappingsHash(mappings []mapper.FieldMapping, config *mapper.MappingConfig) string {
	h := sha256.New()
	if config != nil {
		configJSON, _ := json.Marshal(config)
		h.Write(configJSON)
	} else if len(mappings) > 0 {
		mappingsJSON, _ := json.Marshal(mappings)
		h.Write(mappingsJSON)
	}
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars for cache key
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	return s.withMiddleware(s.mux)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// Query endpoints
	s.mux.HandleFunc("POST /api/v1/query/map", s.handleMapQuery)
	s.mux.HandleFunc("POST /api/v1/query/discover", s.handleDiscoverQuery)
	s.mux.HandleFunc("POST /api/v1/query/validate", s.handleValidateQuery)

	// Mapping configuration endpoints
	s.mux.HandleFunc("POST /api/v1/mappings", s.handleLoadMappings)

	// Documentation endpoint (will serve static swagger UI)
	s.mux.HandleFunc("GET /api/v1/docs", s.handleDocs)
	s.mux.HandleFunc("GET /api/v1/openapi.json", s.handleOpenAPISpec)
}

// withMiddleware wraps the handler with common middleware
func (s *Server) withMiddleware(handler http.Handler) http.Handler {
	return s.recoveryMiddleware(s.loggingMiddleware(s.requestIDMiddleware(s.securityHeadersMiddleware(s.corsMiddleware(s.contentTypeMiddleware(handler))))))
}

// requestIDMiddleware adds request ID to context and response headers
func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing X-Request-ID header, generate if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = s.generateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// recoveryMiddleware recovers from panics and returns a JSON error response
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := s.getRequestID(r.Context())
				s.logger.Error("panic recovered",
					"error", rec,
					"stack", string(debug.Stack()),
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
				)
				s.writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests with structured logging
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := s.getRequestID(r.Context())

		// Create a response writer wrapper to capture status code
		wrw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrw, r)

		duration := time.Since(start)

		// Structure log entry
		s.logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
			"remote_addr", s.getClientIP(r),
			"user_agent", r.Header.Get("User-Agent"),
			"content_length", r.ContentLength,
		)
	})
}

// corsMiddleware adds configurable CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if s.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(s.config.CORSAllowedOrigins) == 1 && s.config.CORSAllowedOrigins[0] == "*" {
			// Allow all origins only if explicitly configured
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", strings.Join(s.config.CORSAllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(s.config.CORSAllowedHeaders, ", "))

		// Allow credentials if not using wildcard origin
		if origin != "" && s.isOriginAllowed(origin) && !(len(s.config.CORSAllowedOrigins) == 1 && s.config.CORSAllowedOrigins[0] == "*") {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if the given origin is in the allowed list
func (s *Server) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range s.config.CORSAllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}

// securityHeadersMiddleware adds security-related HTTP headers
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME-type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Control referrer information
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Prevent XSS attacks
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Basic CSP for API endpoints (more restrictive for docs endpoint)
		if r.URL.Path == "/api/v1/docs" {
			// Allow Swagger UI CDN for docs endpoint
			csp := "default-src 'self'; " +
				"script-src 'self' https://unpkg.com 'unsafe-inline'; " +
				"style-src 'self' https://unpkg.com 'unsafe-inline'; " +
				"img-src 'self' data:; " +
				"font-src 'self' data:; " +
				"connect-src 'self'"
			w.Header().Set("Content-Security-Policy", csp)
		} else {
			// Strict CSP for API endpoints
			w.Header().Set("Content-Security-Policy", "default-src 'none'")
		}

		next.ServeHTTP(w, r)
	})
}

// contentTypeMiddleware sets default content type for API responses
func (s *Server) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/docs" { // Don't set JSON content type for docs endpoint
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture status codes
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// writeJSONResponse writes a JSON response with the specified status code
func (s *Server) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
}

// writeErrorResponse writes an error response in JSON format
func (s *Server) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := ErrorResponse{
		Error:   true,
		Message: message,
		Code:    statusCode,
	}
	s.writeJSONResponse(w, statusCode, response)
}

// generateRequestID creates a new request ID
func (s *Server) generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getRequestID retrieves the request ID from context
func (s *Server) getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// getClientIP extracts the client IP address from the request
func (s *Server) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (common in load balancers/proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (used by some proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
