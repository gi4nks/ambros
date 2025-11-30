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

func TestRerunCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("successful rerun without storage", func(t *testing.T) {
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

		rerunCmd := NewRerunCommand(logger, mockRepo, nil)
		rerunCmd.store = false

		err := rerunCmd.runE(rerunCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful rerun with storage", func(t *testing.T) {
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
			return len(cmd.Tags) >= 1
		})).Return(nil)

		rerunCmd := NewRerunCommand(logger, mockRepo, nil)
		rerunCmd.store = true

		err := rerunCmd.runE(rerunCmd.cmd, []string{"test-id"})
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

		rerunCmd := NewRerunCommand(logger, mockRepo, nil)
		rerunCmd.dryRun = true

		err := rerunCmd.runE(rerunCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Put")
	})

	t.Run("command not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "nonexistent-id").Return(nil, assert.AnError)

		rerunCmd := NewRerunCommand(logger, mockRepo, nil)

		err := rerunCmd.runE(rerunCmd.cmd, []string{"nonexistent-id"})
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
			return !cmd.Status && len(cmd.Tags) >= 1
		})).Return(nil)

		rerunCmd := NewRerunCommand(logger, mockRepo, nil)
		rerunCmd.store = true

		var exitCode int
		rerunCmd.exitFunc = func(code int) { exitCode = code }

		err := rerunCmd.runE(rerunCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		assert.Equal(t, 1, exitCode)
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

		rerunCmd := NewRerunCommand(logger, mockRepo, nil)
		rerunCmd.store = true

		err := rerunCmd.runE(rerunCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("rerun with history flag", func(t *testing.T) {
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

		rerunCmd := NewRerunCommand(logger, mockRepo, nil)
		rerunCmd.history = true

		err := rerunCmd.runE(rerunCmd.cmd, []string{"test-id"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		rerunCmd := NewRerunCommand(logger, mockRepo, nil)

		assert.Equal(t, "rerun <command-id>", rerunCmd.cmd.Use)
		assert.NotNil(t, rerunCmd.Command())
		assert.Equal(t, rerunCmd.cmd, rerunCmd.Command())

		// Test flags
		assert.NotNil(t, rerunCmd.cmd.Flags().Lookup("history"))
		assert.NotNil(t, rerunCmd.cmd.Flags().Lookup("store"))
		assert.NotNil(t, rerunCmd.cmd.Flags().Lookup("dry-run"))

		// Test flag defaults
		historyFlag := rerunCmd.cmd.Flags().Lookup("history")
		assert.Equal(t, "false", historyFlag.DefValue)

		storeFlag := rerunCmd.cmd.Flags().Lookup("store")
		assert.Equal(t, "false", storeFlag.DefValue)

		// Test short flags
		yFlag := rerunCmd.cmd.Flags().ShorthandLookup("y")
		assert.NotNil(t, yFlag)
		assert.Equal(t, yFlag, historyFlag)

		sFlag := rerunCmd.cmd.Flags().ShorthandLookup("s")
		assert.NotNil(t, sFlag)
		assert.Equal(t, sFlag, storeFlag)
	})
}

func TestRerunCommand_ExecuteCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	rerunCmd := NewRerunCommand(logger, mockRepo, nil)

	t.Run("successful command", func(t *testing.T) {
		output, errorMsg, success, err := rerunCmd.executeCommand("echo", []string{"hello"})

		assert.NoError(t, err)
		assert.True(t, success)
		assert.Contains(t, output, "hello")
		assert.Empty(t, errorMsg)
	})

	t.Run("failing command", func(t *testing.T) {
		_, errorMsg, success, err := rerunCmd.executeCommand("false", []string{})

		assert.NoError(t, err)
		assert.False(t, success)
		assert.NotEmpty(t, errorMsg)
	})
}

func TestRerunCommand_GenerateCommandID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	rerunCmd := NewRerunCommand(logger, mockRepo, nil)

	id1 := rerunCmd.generateCommandID()
	time.Sleep(1 * time.Millisecond)
	id2 := rerunCmd.generateCommandID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "CMD-")
	assert.Contains(t, id2, "CMD-")
}

func TestRerunCommand_FormatStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	rerunCmd := NewRerunCommand(logger, mockRepo, nil)

	tests := []struct {
		name     string
		status   bool
		expected string
	}{
		{"success status", true, "Success"},
		{"failed status", false, "Failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rerunCmd.formatStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
