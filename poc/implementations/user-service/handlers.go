package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	storage *Storage
}

// NewHandler creates a new handler
func NewHandler(storage *Storage) *Handler {
	return &Handler{storage: storage}
}

// CreateUser handles POST /api/v1/users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.storage.CreateUser(req.Name, req.Email)
	if err != nil {
		switch err {
		case ErrNameRequired:
			respondError(w, "name is required", http.StatusBadRequest)
		case ErrInvalidEmail:
			respondError(w, "invalid email format", http.StatusBadRequest)
		case ErrEmailAlreadyExists:
			respondError(w, "email already exists", http.StatusConflict)
		default:
			respondError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	respondJSON(w, user, http.StatusCreated)
}

// GetUser handles GET /api/v1/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	id := strings.Split(path, "/")[0]

	if id == "" {
		respondError(w, "user id is required", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUser(id)
	if err != nil {
		if err == ErrUserNotFound {
			respondError(w, "user not found", http.StatusNotFound)
		} else {
			respondError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, user, http.StatusOK)
}

// ListUsers handles GET /api/v1/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users := h.storage.ListUsers()
	response := ListUsersResponse{Users: users}

	respondJSON(w, response, http.StatusOK)
}

// ValidateUser handles POST /api/v1/users/validate
func (h *Handler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, valid := h.storage.ValidateUser(req.UserID)

	if valid {
		response := ValidateUserResponse{
			Valid: true,
			User:  user,
		}
		respondJSON(w, response, http.StatusOK)
	} else {
		response := ValidateUserResponse{
			Valid: false,
		}
		respondJSON(w, response, http.StatusNotFound)
	}
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
