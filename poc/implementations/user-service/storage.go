package main

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrNameRequired      = errors.New("name is required")
)

// Storage provides in-memory storage for users
type Storage struct {
	mu    sync.RWMutex
	users map[string]*User       // id -> user
	emails map[string]string      // normalized email -> user id
}

// NewStorage creates a new in-memory storage
func NewStorage() *Storage {
	return &Storage{
		users:  make(map[string]*User),
		emails: make(map[string]string),
	}
}

// normalizeEmail converts email to lowercase
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) == 0 || len(email) > 255 {
		return false
	}
	// Basic check for @ symbol
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	// Check for dot in domain
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// CreateUser creates a new user
func (s *Storage) CreateUser(name, email string) (*User, error) {
	// Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameRequired
	}
	if len(name) > 100 {
		return nil, errors.New("name too long")
	}

	// Validate email
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	normalizedEmail := normalizeEmail(email)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if email already exists
	if _, exists := s.emails[normalizedEmail]; exists {
		return nil, ErrEmailAlreadyExists
	}

	// Create user with UUIDv4
	user := &User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     normalizedEmail,
		CreatedAt: time.Now().UTC(),
	}

	s.users[user.ID] = user
	s.emails[normalizedEmail] = user.ID

	return user, nil
}

// GetUser retrieves a user by ID
func (s *Storage) GetUser(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// ListUsers returns all users
func (s *Storage) ListUsers() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, *user)
	}

	return users
}

// ValidateUser checks if a user exists by ID
func (s *Storage) ValidateUser(userID string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, false
	}

	return user, true
}
