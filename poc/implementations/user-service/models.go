package main

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ValidateUserRequest represents the request to validate a user
type ValidateUserRequest struct {
	UserID string `json:"user_id"`
}

// ValidateUserResponse represents the response for user validation
type ValidateUserResponse struct {
	Valid bool  `json:"valid"`
	User  *User `json:"user,omitempty"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users []User `json:"users"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
