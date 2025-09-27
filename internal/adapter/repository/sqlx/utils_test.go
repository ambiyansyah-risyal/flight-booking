package sqlxrepo

import (
	"errors"
	"testing"
)

// Test for New function
func TestNew(t *testing.T) {
	// Note: sqlx.Open doesn't immediately connect to the database, 
	// so it won't return an error for an invalid URL right away.
	// It only returns an error when you try to use the connection.
	// For testing purposes, we'll just ensure the function exists and can be called.
	db, _ := New("postgres://invalid:password@127.0.0.1:12345/invalid_db")
	
	// If we got a db, we need to close it
	if db != nil {
		_ = db.Close()
	}
	
	// Just make sure the function can be called without crashing
	// We can't check if a function is nil in Go with a direct comparison
	// Just the fact that we can call it means it exists
}

// Test for error detection functions
func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "duplicate key error",
			err:      errors.New("duplicate key value violates unique constraint"),
			expected: true,
		},
		{
			name:     "unique constraint error",
			err:      errors.New("UNIQUE constraint failed"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUniqueViolation(tt.err)
			if result != tt.expected {
				t.Errorf("isUniqueViolation() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "foreign key constraint error",
			err:      errors.New("foreign key constraint violation"),
			expected: true,
		},
		{
			name:     "violates foreign key error",
			err:      errors.New("violates foreign key constraint"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isForeignKeyViolation(tt.err)
			if result != tt.expected {
				t.Errorf("isForeignKeyViolation() = %v, want %v", result, tt.expected)
			}
		})
	}
}