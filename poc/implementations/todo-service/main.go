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

	// Get user-service URL from environment or use default
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://user-service:8080"
	}

	// Initialize storage
	storage := NewStorage()

	// Initialize user client
	userClient := NewUserClient(userServiceURL)

	// Initialize handler
	handler := NewHandler(storage, userClient)

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", handler.Health)

	// API routes
	mux.HandleFunc("/api/v1/todos", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path and method
		if r.URL.Path == "/api/v1/todos" {
			if r.Method == http.MethodPost {
				handler.CreateTodo(w, r)
			} else if r.Method == http.MethodGet {
				handler.ListTodos(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc("/api/v1/todos/", func(w http.ResponseWriter, r *http.Request) {
		// Handle /api/v1/todos/{id}
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/todos/")

		if path == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodGet {
			handler.GetTodo(w, r)
		} else if r.Method == http.MethodPatch {
			handler.UpdateTodoStatus(w, r)
		} else if r.Method == http.MethodDelete {
			handler.DeleteTodo(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	addr := ":" + port
	log.Printf("Starting todo-service on %s", addr)
	log.Printf("User service URL: %s", userServiceURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
