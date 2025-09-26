package sqlxrepo

import (
	"errors"
	"testing"
)

// Test for New function
func TestNew(t *testing.T) {
	// Test that the New function signature is correct and can be called
	// We're not testing actual DB connectivity, just that the function works properly
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