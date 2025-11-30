package commands

import (
	"testing"
	"time"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLastCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("successful last commands retrieval with default limit", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Now().Add(-1 * time.Hour),
				},
				Name:   "ls",
				Status: true,
			},
			{
				Entity: models.Entity{
					ID:        "cmd-2",
					CreatedAt: time.Now().Add(-2 * time.Hour),
				},
				Name:   "pwd",
				Status: false,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		lastCmd := NewLastCommand(logger, mockRepo, nil)
		err := lastCmd.runE(lastCmd.cmd, []string{})

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful last commands retrieval with custom limit", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Now(),
				},
				Name:   "echo",
				Status: true,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		lastCmd := NewLastCommand(logger, mockRepo, nil)
		lastCmd.limit = 5

		err := lastCmd.runE(lastCmd.cmd, []string{})

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failed commands only filter", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Now(),
				},
				Name:   "success-cmd",
				Status: true,
			},
			{
				Entity: models.Entity{
					ID:        "cmd-2",
					CreatedAt: time.Now(),
				},
				Name:   "failed-cmd",
				Status: false,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		lastCmd := NewLastCommand(logger, mockRepo, nil)
		lastCmd.failedOnly = true

		err := lastCmd.runE(lastCmd.cmd, []string{})

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error handling", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(nil, assert.AnError)

		lastCmd := NewLastCommand(logger, mockRepo, nil)
		err := lastCmd.runE(lastCmd.cmd, []string{})

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty commands list", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return([]models.Command{}, nil)

		lastCmd := NewLastCommand(logger, mockRepo, nil)
		err := lastCmd.runE(lastCmd.cmd, []string{})

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("test command structure", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		lastCmd := NewLastCommand(logger, mockRepo, nil)

		assert.Equal(t, "last", lastCmd.cmd.Use)
		assert.Equal(t, "Show last executed commands", lastCmd.cmd.Short)
		assert.NotNil(t, lastCmd.Command())
		assert.Equal(t, lastCmd.cmd, lastCmd.Command())
	})

	t.Run("test flags setup", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		lastCmd := NewLastCommand(logger, mockRepo, nil)

		// Test that flags are properly set up
		limitFlag := lastCmd.cmd.Flags().Lookup("limit")
		assert.NotNil(t, limitFlag)
		assert.Equal(t, "10", limitFlag.DefValue)

		failedFlag := lastCmd.cmd.Flags().Lookup("failed")
		assert.NotNil(t, failedFlag)
		assert.Equal(t, "false", failedFlag.DefValue)

		// Test short flags
		nFlag := lastCmd.cmd.Flags().ShorthandLookup("n")
		assert.NotNil(t, nFlag)
		assert.Equal(t, nFlag, limitFlag)

		fFlag := lastCmd.cmd.Flags().ShorthandLookup("f")
		assert.NotNil(t, fFlag)
		assert.Equal(t, fFlag, failedFlag)
	})
}

func TestLastCommand_FormatStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	lastCmd := NewLastCommand(logger, mockRepo, nil)

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
			result := lastCmd.formatStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
