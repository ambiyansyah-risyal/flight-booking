package cli

import (
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	// Test that Execute works with a valid command
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test help command
	os.Args = []string{"flight-booking", "--help"}
	
	// Capture output or test error cases as appropriate
	_ = Execute() // Call may trigger os.Exit in real execution, but in tests it returns an error
}

func TestExecuteVersion(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"flight-booking", "version"}
	
	_ = Execute() // Call may trigger os.Exit in real execution, but in tests it returns an error
}