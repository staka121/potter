package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	storage    *Storage
	userClient *UserClient
}

// NewHandler creates a new handler
func NewHandler(storage *Storage, userClient *UserClient) *Handler {
	return &Handler{
		storage:    storage,
		userClient: userClient,
	}
}

// CreateTodo handles POST /api/v1/todos
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from X-User-ID header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	// Validate user exists via user-service
	valid, err := h.userClient.ValidateUser(userID)
	if err != nil {
		respondError(w, "failed to validate user", http.StatusInternalServerError)
		return
	}
	if !valid {
		respondError(w, "invalid user", http.StatusUnauthorized)
		return
	}

	var req CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	todo, err := h.storage.CreateTodo(userID, req.Title, req.Description)
	if err != nil {
		switch err {
		case ErrTitleRequired:
			respondError(w, "title cannot be empty", http.StatusBadRequest)
		case ErrTitleTooLong:
			respondError(w, "title too long (max 200 characters)", http.StatusBadRequest)
		case ErrDescriptionTooLong:
			respondError(w, "description too long (max 2000 characters)", http.StatusBadRequest)
		default:
			respondError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	respondJSON(w, todo, http.StatusCreated)
}

// ListTodos handles GET /api/v1/todos
func (h *Handler) ListTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from X-User-ID header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	// Get status filter from query parameter
	statusParam := r.URL.Query().Get("status")
	var statusFilter *string
	if statusParam != "" {
		// Validate status value
		if statusParam != "pending" && statusParam != "completed" {
			respondError(w, "invalid status value", http.StatusBadRequest)
			return
		}
		statusFilter = &statusParam
	}

	todos := h.storage.ListTodos(userID, statusFilter)
	response := ListTodosResponse{Todos: todos}

	respondJSON(w, response, http.StatusOK)
}

// GetTodo handles GET /api/v1/todos/{id}
func (h *Handler) GetTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from X-User-ID header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/todos/")
	id := strings.Split(path, "/")[0]

	if id == "" {
		respondError(w, "todo id is required", http.StatusBadRequest)
		return
	}

	todo, err := h.storage.GetTodo(id)
	if err != nil {
		switch err {
		case ErrTodoNotFound:
			respondError(w, "todo not found", http.StatusNotFound)
		case ErrInvalidUUID:
			respondError(w, "invalid id format", http.StatusBadRequest)
		default:
			respondError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Verify the TODO belongs to the requesting user
	if todo.UserID != userID {
		respondError(w, "todo not found", http.StatusNotFound)
		return
	}

	respondJSON(w, todo, http.StatusOK)
}

// UpdateTodoStatus handles PATCH /api/v1/todos/{id}
func (h *Handler) UpdateTodoStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from X-User-ID header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/todos/")
	id := strings.Split(path, "/")[0]

	if id == "" {
		respondError(w, "todo id is required", http.StatusBadRequest)
		return
	}

	// First, verify the TODO exists and belongs to the user
	existingTodo, err := h.storage.GetTodo(id)
	if err != nil {
		switch err {
		case ErrTodoNotFound:
			respondError(w, "todo not found", http.StatusNotFound)
		case ErrInvalidUUID:
			respondError(w, "invalid id format", http.StatusBadRequest)
		default:
			respondError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if existingTodo.UserID != userID {
		respondError(w, "todo not found", http.StatusNotFound)
		return
	}

	var req UpdateTodoStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	todo, err := h.storage.UpdateTodoStatus(id, req.Status)
	if err != nil {
		switch err {
		case ErrTodoNotFound:
			respondError(w, "todo not found", http.StatusNotFound)
		case ErrInvalidStatus:
			respondError(w, "invalid status value", http.StatusBadRequest)
		case ErrInvalidUUID:
			respondError(w, "invalid id format", http.StatusBadRequest)
		default:
			respondError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, todo, http.StatusOK)
}

// DeleteTodo handles DELETE /api/v1/todos/{id}
func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from X-User-ID header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/todos/")
	id := strings.Split(path, "/")[0]

	if id == "" {
		respondError(w, "todo id is required", http.StatusBadRequest)
		return
	}

	// First, verify the TODO exists and belongs to the user
	existingTodo, err := h.storage.GetTodo(id)
	if err != nil {
		switch err {
		case ErrTodoNotFound:
			respondError(w, "todo not found", http.StatusNotFound)
		case ErrInvalidUUID:
			respondError(w, "invalid id format", http.StatusBadRequest)
		default:
			respondError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if existingTodo.UserID != userID {
		respondError(w, "todo not found", http.StatusNotFound)
		return
	}

	err = h.storage.DeleteTodo(id)
	if err != nil {
		switch err {
		case ErrTodoNotFound:
			respondError(w, "todo not found", http.StatusNotFound)
		case ErrInvalidUUID:
			respondError(w, "invalid id format", http.StatusBadRequest)
		default:
			respondError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
