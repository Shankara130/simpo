package repositories

import "fmt"

// Common repository errors

// ErrNotFound is returned when a record is not found
var ErrNotFound = NewRepositoryError("record not found")

// ErrDuplicate is returned when a unique constraint violation occurs
var ErrDuplicate = NewRepositoryError("duplicate record")

// ErrInvalidInput is returned when input validation fails
var ErrInvalidInput = NewRepositoryError("invalid input")

// RepositoryError represents a repository-level error
type RepositoryError struct {
	Message string
}

func (e *RepositoryError) Error() string {
	return e.Message
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(message string) *RepositoryError {
	return &RepositoryError{Message: message}
}

// Errorf creates a new repository error with formatted message
func Errorf(format string, args ...interface{}) *RepositoryError {
	return &RepositoryError{Message: fmt.Sprintf(format, args...)}
}
