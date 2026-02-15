package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize storage
	storage := NewStorage()

	// Initialize handler
	handler := NewHandler(storage)

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", handler.Health)

	// API routes
	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path and method
		if r.URL.Path == "/api/v1/users" {
			if r.Method == http.MethodPost {
				handler.CreateUser(w, r)
			} else if r.Method == http.MethodGet {
				handler.ListUsers(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc("/api/v1/users/validate", handler.ValidateUser)

	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		// Handle /api/v1/users/{id}
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")

		// If path is "validate", don't handle here
		if path == "validate" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		handler.GetUser(w, r)
	})

	// Start server
	addr := ":" + port
	log.Printf("Starting user-service on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
