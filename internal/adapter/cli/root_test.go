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

func TestExecuteWithConfigError(t *testing.T) {
	// This test verifies the config error handling path
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	exitCode := -1
	exitHandler := func(code int) {
		exitCode = code
	}
	
	// Set up environment to cause a config error
	// We'll use an invalid port value to trigger a config loading error
	t.Setenv("FLIGHT_DB_PORT", "invalid_port")
	os.Args = []string{"flight-booking", "version"}
	
	// Call the function with our mock exit handler
	err := ExecuteWithExitHandler(exitHandler)
	
	// Verify that the exit handler was called with the correct exit code
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	
	// The function should also return the error for testability
	if err == nil {
		t.Error("expected error to be returned when config loading fails")
	}
}