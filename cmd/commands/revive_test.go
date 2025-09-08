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

func TestReviveCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("successful command revive without storage", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"hello"},
			Status:    true,
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)

		reviveCmd := NewReviveCommand(logger, mockRepo)
		reviveCmd.store = false

		err := reviveCmd.runE(reviveCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful command revive with storage", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storedCmd := &models.Command{
			Entity: models.Entity{
				ID:        "test-id",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"hello"},
			Status:    true,
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil)

		reviveCmd := NewReviveCommand(logger, mockRepo)
		reviveCmd.store = true

		err := reviveCmd.runE(reviveCmd.cmd, []string{"test-id"})
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
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)

		reviveCmd := NewReviveCommand(logger, mockRepo)
		reviveCmd.dryRun = true

		err := reviveCmd.runE(reviveCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		// Should not call Put in dry run mode
		mockRepo.AssertNotCalled(t, "Put")
	})

	t.Run("command not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "nonexistent-id").Return(nil, assert.AnError)

		reviveCmd := NewReviveCommand(logger, mockRepo)

		err := reviveCmd.runE(reviveCmd.cmd, []string{"nonexistent-id"})
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
		}
		mockRepo.On("Get", "test-id").Return(storedCmd, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return !cmd.Status // New execution should be marked as failed
		})).Return(nil)

		reviveCmd := NewReviveCommand(logger, mockRepo)
		reviveCmd.store = true

		// Override the os.Exit behavior for testing
		var exitCode int
		reviveCmd.exitFunc = func(code int) {
			exitCode = code
		}

		err := reviveCmd.runE(reviveCmd.cmd, []string{"test-id"})
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

		reviveCmd := NewReviveCommand(logger, mockRepo)
		reviveCmd.store = true

		err := reviveCmd.runE(reviveCmd.cmd, []string{"test-id"})
		assert.NoError(t, err) // Should still succeed even if storage fails
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		reviveCmd := NewReviveCommand(logger, mockRepo)

		assert.Equal(t, "revive <command-id>", reviveCmd.cmd.Use)
		assert.Equal(t, "Re-execute a previously stored command", reviveCmd.cmd.Short)
		assert.NotNil(t, reviveCmd.Command())
		assert.Equal(t, reviveCmd.cmd, reviveCmd.Command())

		// Test flags
		assert.NotNil(t, reviveCmd.cmd.Flags().Lookup("store"))
		assert.NotNil(t, reviveCmd.cmd.Flags().Lookup("dry-run"))

		// Test flag defaults
		storeFlag := reviveCmd.cmd.Flags().Lookup("store")
		assert.Equal(t, "false", storeFlag.DefValue)
	})
}

func TestReviveCommand_ExecuteCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	reviveCmd := NewReviveCommand(logger, mockRepo)

	t.Run("successful command", func(t *testing.T) {
		output, errorMsg, success, err := reviveCmd.executeCommand("echo", []string{"hello"})

		assert.NoError(t, err)
		assert.True(t, success)
		assert.Contains(t, output, "hello")
		assert.Empty(t, errorMsg)
	})

	t.Run("failing command", func(t *testing.T) {
		_, errorMsg, success, err := reviveCmd.executeCommand("false", []string{})

		assert.NoError(t, err) // executeCommand doesn't return error for command failure
		assert.False(t, success)
		assert.NotEmpty(t, errorMsg)
	})
}

func TestReviveCommand_GenerateCommandID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	reviveCmd := NewReviveCommand(logger, mockRepo)

	id1 := reviveCmd.generateCommandID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := reviveCmd.generateCommandID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "CMD-")
	assert.Contains(t, id2, "CMD-")
}

func TestReviveCommand_FormatStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	reviveCmd := NewReviveCommand(logger, mockRepo)

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
			result := reviveCmd.formatStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
