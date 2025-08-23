// @title SPL Toolkit API
// @version 1.0.0
// @description REST API for the SPL Toolkit library that provides field mapping, query discovery, and validation capabilities for Splunk SPL queries.
// @description This API enables programmatic analysis and manipulation of Splunk SPL queries in a robust, language-aware fashion.
//
// @contact.name SPL Toolkit Support
// @contact.email support@example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @servers.url http://localhost:8080/api/v1
// @servers.description Development server
//
// @tag.name health
// @tag.description Health check operations
//
// @tag.name query
// @tag.description SPL query operations including mapping, discovery, and validation
//
// @tag.name mappings
// @tag.description Field mapping management operations
//
// @tag.name docs
// @tag.description API documentation endpoints

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/delgado-jacob/spl-toolkit/pkg/api"
)

// Version will be set at build time via ldflags
var Version = "dev"

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create API server with version
	server := api.NewServerWithVersion(Version)

	// Create HTTP server with hardened settings
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.Handler(),
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes, // 1MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("SPL Toolkit API Server %s starting on port %s", Version, port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
