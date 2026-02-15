package main

import (
	"time"
)

// Todo represents a TODO item in the system (internal representation with UserID)
type Todo struct {
	ID          string    `json:"id"`
	UserID      string    `json:"-"` // Internal field, not exposed in API
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateTodoRequest represents the request to create a TODO
type CreateTodoRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

// UpdateTodoStatusRequest represents the request to update TODO status
type UpdateTodoStatusRequest struct {
	Status string `json:"status"`
}

// ListTodosResponse represents the response for listing TODOs
type ListTodosResponse struct {
	Todos []Todo `json:"todos"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
