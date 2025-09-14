package commands

import (
	"os"
	"testing"

	"github.com/creack/pty"
)

// Test that executeCommandAuto can run a command that expects a TTY.
// This test will be skipped if the test runner does not have a TTY available.
func TestExecuteCommandAuto_PTY(t *testing.T) {
	// Try to open a pty for the test; if it fails, skip the test
	master, slave, err := pty.Open()
	if err != nil {
		t.Skip("No PTY available for PTY test; skipping")
	}
	defer master.Close()
	defer slave.Close()

	// Replace os.Stdin with the pty master so the code under test sees a TTY.
	oldStdin := os.Stdin
	os.Stdin = master
	defer func() { os.Stdin = oldStdin }()

	rc := &RunCommand{}
	exitCode, err := rc.executeCommandAuto("bash", []string{"-lc", "stty size"})
	if err != nil {
		t.Fatalf("executeCommandAuto returned error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

// Test non-PTY path: run a simple non-interactive command
func TestExecuteCommandAuto_NonPTY(t *testing.T) {
	rc := &RunCommand{}
	exitCode, err := rc.executeCommandAuto("true", []string{})
	if err != nil {
		t.Fatalf("executeCommandAuto returned error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func isTerminalForTests() bool {
	// Keep backward compatibility, but prefer pty.Open
	_, _, err := pty.Open()
	return err == nil
}
