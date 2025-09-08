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

func TestRecallCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("successful command recall without storage", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"hello"},
			Status:    true,
			Tags:      []string{"original"},
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)

		recallCmd := NewRecallCommand(logger, mockRepo)
		recallCmd.store = false

		err := recallCmd.runE(recallCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful command recall with storage", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"hello"},
			Status:    true,
			Tags:      []string{"original"},
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return len(cmd.Tags) == 2 && cmd.Tags[1] == "recalled"
		})).Return(nil)

		recallCmd := NewRecallCommand(logger, mockRepo)
		recallCmd.store = true

		err := recallCmd.runE(recallCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("dry run mode", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "rm",
			Arguments: []string{"-rf", "/"},
			Status:    false,
			Tags:      []string{"dangerous"},
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)

		recallCmd := NewRecallCommand(logger, mockRepo)
		recallCmd.dryRun = true

		err := recallCmd.runE(recallCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		// Should not call Put in dry run mode
		mockRepo.AssertNotCalled(t, "Put")
	})

	t.Run("command not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "nonexistent-id").Return(nil, assert.AnError)

		recallCmd := NewRecallCommand(logger, mockRepo)

		err := recallCmd.runE(recallCmd.cmd, []string{"nonexistent-id"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("failed command execution with exit", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "false",
			Arguments: []string{},
			Status:    true, // Original was successful
			Tags:      []string{"test"},
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return !cmd.Status && len(cmd.Tags) == 2 && cmd.Tags[1] == "recalled"
		})).Return(nil)

		recallCmd := NewRecallCommand(logger, mockRepo)
		recallCmd.store = true

		// Override the os.Exit behavior for testing
		var exitCode int
		recallCmd.exitFunc = func(code int) {
			exitCode = code
		}

		err := recallCmd.runE(recallCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)       // The runE itself shouldn't error
		assert.Equal(t, 1, exitCode) // But it should call exit with code 1
		mockRepo.AssertExpectations(t)
	})

	t.Run("storage failure should not stop execution", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"test"},
			Status:    true,
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(assert.AnError)

		recallCmd := NewRecallCommand(logger, mockRepo)
		recallCmd.store = true

		err := recallCmd.runE(recallCmd.cmd, []string{"test-id"})
		assert.NoError(t, err) // Should still succeed even if storage fails
		mockRepo.AssertExpectations(t)
	})

	t.Run("recall with history flag", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"historical"},
			Status:    true,
			Tags:      []string{"old"},
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)

		recallCmd := NewRecallCommand(logger, mockRepo)
		recallCmd.history = true

		err := recallCmd.runE(recallCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		recallCmd := NewRecallCommand(logger, mockRepo)

		assert.Equal(t, "recall <command-id>", recallCmd.cmd.Use)
		assert.Equal(t, "Recall and execute a stored command", recallCmd.cmd.Short)
		assert.NotNil(t, recallCmd.Command())
		assert.Equal(t, recallCmd.cmd, recallCmd.Command())

		// Test flags
		assert.NotNil(t, recallCmd.cmd.Flags().Lookup("history"))
		assert.NotNil(t, recallCmd.cmd.Flags().Lookup("store"))
		assert.NotNil(t, recallCmd.cmd.Flags().Lookup("dry-run"))

		// Test flag defaults
		historyFlag := recallCmd.cmd.Flags().Lookup("history")
		assert.Equal(t, "false", historyFlag.DefValue)

		storeFlag := recallCmd.cmd.Flags().Lookup("store")
		assert.Equal(t, "false", storeFlag.DefValue)

		// Test short flags
		yFlag := recallCmd.cmd.Flags().ShorthandLookup("y")
		assert.NotNil(t, yFlag)
		assert.Equal(t, yFlag, historyFlag)

		sFlag := recallCmd.cmd.Flags().ShorthandLookup("s")
		assert.NotNil(t, sFlag)
		assert.Equal(t, sFlag, storeFlag)
	})
}

func TestRecallCommand_ExecuteCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	recallCmd := NewRecallCommand(logger, mockRepo)

	t.Run("successful command", func(t *testing.T) {
		output, errorMsg, success, err := recallCmd.executeCommand("echo", []string{"hello"})

		assert.NoError(t, err)
		assert.True(t, success)
		assert.Contains(t, output, "hello")
		assert.Empty(t, errorMsg)
	})

	t.Run("failing command", func(t *testing.T) {
		_, errorMsg, success, err := recallCmd.executeCommand("false", []string{})

		assert.NoError(t, err) // executeCommand doesn't return error for command failure
		assert.False(t, success)
		assert.NotEmpty(t, errorMsg)
	})
}

func TestRecallCommand_GenerateCommandID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	recallCmd := NewRecallCommand(logger, mockRepo)

	id1 := recallCmd.generateCommandID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := recallCmd.generateCommandID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "CMD-")
	assert.Contains(t, id2, "CMD-")
}

func TestRecallCommand_FormatStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	recallCmd := NewRecallCommand(logger, mockRepo)

	tests := []struct {
		name     string
		status   bool
		expected string
	}{
		{
			name:     "success status",
			status:   true,
			expected: "Success",
		},
		{
			name:     "failed status",
			status:   false,
			expected: "Failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := recallCmd.formatStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
