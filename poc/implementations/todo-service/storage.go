package main

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage errors
var (
	ErrTodoNotFound      = errors.New("todo not found")
	ErrTitleRequired     = errors.New("title is required")
	ErrTitleTooLong      = errors.New("title too long (max 200 characters)")
	ErrDescriptionTooLong = errors.New("description too long (max 2000 characters)")
	ErrInvalidStatus     = errors.New("invalid status value")
	ErrInvalidUUID       = errors.New("invalid id format")
)

// Storage handles in-memory storage of TODOs
type Storage struct {
	todos map[string]*Todo
	mu    sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{
		todos: make(map[string]*Todo),
	}
}

// CreateTodo creates a new TODO item
func (s *Storage) CreateTodo(userID, title string, description *string) (*Todo, error) {
	// Validate title
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, ErrTitleRequired
	}
	if len(trimmedTitle) > 200 {
		return nil, ErrTitleTooLong
	}

	// Validate description
	if description != nil {
		if len(*description) > 2000 {
			return nil, ErrDescriptionTooLong
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new TODO
	todo := &Todo{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       trimmedTitle,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now().UTC(),
	}

	s.todos[todo.ID] = todo
	return todo, nil
}

// GetTodo retrieves a TODO by ID
func (s *Storage) GetTodo(id string) (*Todo, error) {
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidUUID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, ErrTodoNotFound
	}

	return todo, nil
}

// ListTodos retrieves all TODOs, optionally filtered by status and user ID
func (s *Storage) ListTodos(userID string, status *string) []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]Todo, 0)

	for _, todo := range s.todos {
		// Filter by user ID
		if todo.UserID != userID {
			continue
		}

		// Filter by status if provided
		if status != nil && todo.Status != *status {
			continue
		}

		todos = append(todos, *todo)
	}

	// Sort by created_at descending (newest first)
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].CreatedAt.After(todos[j].CreatedAt)
	})

	return todos
}

// UpdateTodoStatus updates the status of a TODO
func (s *Storage) UpdateTodoStatus(id, status string) (*Todo, error) {
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidUUID
	}

	// Validate status
	if status != "pending" && status != "completed" {
		return nil, ErrInvalidStatus
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, ErrTodoNotFound
	}

	todo.Status = status
	return todo, nil
}

// DeleteTodo deletes a TODO by ID
func (s *Storage) DeleteTodo(id string) error {
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidUUID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.todos[id]; !exists {
		return ErrTodoNotFound
	}

	delete(s.todos, id)
	return nil
}
