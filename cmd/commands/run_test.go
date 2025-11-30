package commands

import (
	"testing"
	"time"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestRunCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("successful command execution with storage", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.store = true

		// Use a command that won't cause the test to exit
		err := runCmd.runE(runCmd.cmd, []string{"true"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command execution without storage", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.store = false

		err := runCmd.runE(runCmd.cmd, []string{"true"})
		assert.NoError(t, err)
		// Should not call Put when store is false
		mockRepo.AssertNotCalled(t, "Put")
	})

	t.Run("dry run mode", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.dryRun = true
		runCmd.opts.tag = []string{"test"}
		runCmd.opts.category = "testing"

		err := runCmd.runE(runCmd.cmd, []string{"rm", "-rf", "/"})
		assert.NoError(t, err)
		// Should not execute or store anything in dry run
		mockRepo.AssertNotCalled(t, "Put")
	})

	t.Run("command with tags and category", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return len(cmd.Tags) == 2 &&
				cmd.Tags[0] == "test" &&
				cmd.Tags[1] == "demo" &&
				cmd.Category == "utility"
		})).Return(nil)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.store = true
		runCmd.opts.tag = []string{"test", "demo"}
		runCmd.opts.category = "utility"

		err := runCmd.runE(runCmd.cmd, []string{"true"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command with template", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		template := &models.Template{
			Entity:  models.Entity{ID: "tmpl1"},
			Name:    "echo-template",
			Pattern: "echo {0} world",
		}
		mockRepo.On("GetTemplate", "echo-template").Return(template, nil)
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.template = "echo-template"
		runCmd.opts.store = true

		err := runCmd.runE(runCmd.cmd, []string{"hello"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("template not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetTemplate", "nonexistent").Return(nil, assert.AnError)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.template = "nonexistent"

		err := runCmd.runE(runCmd.cmd, []string{"echo", "hello"})
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no command specified", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		runCmd := NewRunCommand(logger, mockRepo, nil)

		err := runCmd.runE(runCmd.cmd, []string{})
		assert.Error(t, err)
	})

	t.Run("storage failure should not stop execution", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(assert.AnError)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.store = true

		// Should still succeed even if storage fails
		err := runCmd.runE(runCmd.cmd, []string{"true"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failed command execution", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return !cmd.Status // Command should be marked as failed
		})).Return(nil)

		runCmd := NewRunCommand(logger, mockRepo, nil)
		runCmd.opts.store = true

		// Override the os.Exit behavior for testing
		originalExit := runCmd.exitFunc
		var exitCode int
		runCmd.exitFunc = func(code int) {
			exitCode = code
		}
		defer func() {
			runCmd.exitFunc = originalExit
		}()

		err := runCmd.runE(runCmd.cmd, []string{"false"})
		assert.NoError(t, err)       // The runE itself shouldn't error
		assert.Equal(t, 1, exitCode) // But it should call exit with code 1
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		runCmd := NewRunCommand(logger, mockRepo, nil)

		assert.Equal(t, "run [flags] [--] <command> [args...]", runCmd.cmd.Use)
		assert.Equal(t, "Run a command and optionally store its execution details", runCmd.cmd.Short)
		assert.NotNil(t, runCmd.Command())
		assert.Equal(t, runCmd.cmd, runCmd.Command())

		// Test flags
		assert.NotNil(t, runCmd.cmd.Flags().Lookup("store"))
		assert.NotNil(t, runCmd.cmd.Flags().Lookup("tag"))
		assert.NotNil(t, runCmd.cmd.Flags().Lookup("category"))
		assert.NotNil(t, runCmd.cmd.Flags().Lookup("template"))
		assert.NotNil(t, runCmd.cmd.Flags().Lookup("dry-run"))

		// Test flag defaults
		storeFlag := runCmd.cmd.Flags().Lookup("store")
		assert.Equal(t, "true", storeFlag.DefValue)
	})
}

func TestRunCommand_ExecuteCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	runCmd := NewRunCommand(logger, mockRepo, nil)

	t.Run("successful command", func(t *testing.T) {
		output, errorMsg, success, err := runCmd.executor.ExecuteCommand("echo", []string{"hello"})

		assert.NoError(t, err)
		assert.True(t, success)
		assert.Contains(t, output, "hello")
		assert.Empty(t, errorMsg)
	})

	t.Run("successful command with true", func(t *testing.T) {
		_, errorMsg, success, err := runCmd.executor.ExecuteCommand("true", []string{})

		assert.NoError(t, err)
		assert.True(t, success)
		assert.Empty(t, errorMsg)
	})

	t.Run("failing command", func(t *testing.T) {
		_, errorMsg, success, err := runCmd.executor.ExecuteCommand("false", []string{})

		assert.NoError(t, err) // executeCommand doesn't return error for command failure
		assert.False(t, success)
		assert.NotEmpty(t, errorMsg)
	})

	t.Run("command with error output", func(t *testing.T) {
		// Use a command that writes to stderr
		output, errorMsg, success, err := runCmd.executor.ExecuteCommand("sh", []string{"-c", "echo 'error' >&2; exit 1"})

		assert.NoError(t, err)
		assert.False(t, success)
		assert.Contains(t, output, "error") // CombinedOutput captures both stdout and stderr
		assert.NotEmpty(t, errorMsg)
	})

	t.Run("nonexistent command", func(t *testing.T) {
		_, errorMsg, success, err := runCmd.executor.ExecuteCommand("nonexistentcommand123", []string{})

		assert.NoError(t, err) // executeCommand doesn't return error for command not found
		assert.False(t, success)
		assert.NotEmpty(t, errorMsg)
	})
}

func TestRunCommand_GenerateCommandID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	runCmd := NewRunCommand(logger, mockRepo, nil)

	id1 := runCmd.generateCommandID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := runCmd.generateCommandID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // IDs should be unique
	assert.Contains(t, id1, "CMD-")
	assert.Contains(t, id2, "CMD-")
}
