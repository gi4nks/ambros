package commands

import (
	"testing"

	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Test that --auto preserves exit code and streams output for a simple non-interactive command
func TestRunCommand_AutoMode_NonInteractive(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	runCmd := NewRunCommand(logger, mockRepo, nil)
	runCmd.opts.auto = true
	runCmd.opts.store = false

	// Override exitFunc so test can observe exit code
	var capturedExit int
	runCmd.exitFunc = func(code int) {
		capturedExit = code
	}

	// Use a simple command that prints and exits with code 5
	// sh -c 'echo hello; exit 5'
	err := runCmd.runE(runCmd.cmd, []string{"sh", "-c", "echo hello; exit 5"})
	assert.NoError(t, err)
	// In auto mode, runE should return nil but exitFunc should be called with child's exit code
	assert.Equal(t, 5, capturedExit)
}

// Test that --auto returns 0 for successful command
func TestRunCommand_AutoMode_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	runCmd := NewRunCommand(logger, mockRepo, nil)
	runCmd.opts.auto = true
	runCmd.opts.store = false

	var capturedExit int = -1
	runCmd.exitFunc = func(code int) {
		capturedExit = code
	}

	err := runCmd.runE(runCmd.cmd, []string{"true"})
	assert.NoError(t, err)
	// No exit call on success (exitFunc shouldn't be invoked), capturedExit stays -1
	assert.Equal(t, -1, capturedExit)
}
