package commands

import (
	"testing"
	"time"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogsCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("list all logs", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		commands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "test-1",
					CreatedAt: time.Now(),
				},
				Name:   "echo",
				Status: true,
			},
			{
				Entity: models.Entity{
					ID:        "test-2",
					CreatedAt: time.Now(),
				},
				Name:   "ls",
				Status: false,
			},
		}
		mockRepo.On("GetAllCommands").Return(commands, nil)

		logsCmd := NewLogsCommand(logger, mockRepo, nil)

		err := logsCmd.runE(logsCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list failed logs only", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		commands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "test-1",
					CreatedAt: time.Now(),
				},
				Name:   "test",
				Status: false,
			},
			{
				Entity: models.Entity{
					ID:        "test-2",
					CreatedAt: time.Now(),
				},
				Name:   "success",
				Status: true,
			},
		}
		mockRepo.On("GetAllCommands").Return(commands, nil)

		logsCmd := NewLogsCommand(logger, mockRepo, nil)
		logsCmd.failedOnly = true

		err := logsCmd.runE(logsCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list logs since date", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)

		commands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "test-1",
					CreatedAt: now,
				},
				Name:   "recent",
				Status: true,
			},
			{
				Entity: models.Entity{
					ID:        "test-2",
					CreatedAt: yesterday,
				},
				Name:   "old",
				Status: true,
			},
		}
		mockRepo.On("GetAllCommands").Return(commands, nil)

		logsCmd := NewLogsCommand(logger, mockRepo, nil)
		logsCmd.since = now.Add(-12 * time.Hour).Format("2006-01-02")

		err := logsCmd.runE(logsCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid date format", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return([]models.Command{}, nil)

		logsCmd := NewLogsCommand(logger, mockRepo, nil)
		logsCmd.since = "invalid-date"

		err := logsCmd.runE(logsCmd.cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date format")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(nil, assert.AnError)

		logsCmd := NewLogsCommand(logger, mockRepo, nil)

		err := logsCmd.runE(logsCmd.cmd, []string{})
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		logsCmd := NewLogsCommand(logger, mockRepo, nil)

		assert.Equal(t, "logs", logsCmd.cmd.Use)
		assert.Equal(t, "Show command execution logs", logsCmd.cmd.Short)
		assert.NotNil(t, logsCmd.Command())
		assert.Equal(t, logsCmd.cmd, logsCmd.Command())

		// Test flags
		sinceFlag := logsCmd.cmd.Flags().Lookup("since")
		assert.NotNil(t, sinceFlag)

		failedFlag := logsCmd.cmd.Flags().Lookup("failed")
		assert.NotNil(t, failedFlag)
		assert.Equal(t, "false", failedFlag.DefValue)

		// Test short flags
		sFlag := logsCmd.cmd.Flags().ShorthandLookup("s")
		assert.NotNil(t, sFlag)
		assert.Equal(t, sFlag, sinceFlag)

		fFlag := logsCmd.cmd.Flags().ShorthandLookup("f")
		assert.NotNil(t, fFlag)
		assert.Equal(t, fFlag, failedFlag)
	})

	t.Run("empty logs", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return([]models.Command{}, nil)

		logsCmd := NewLogsCommand(logger, mockRepo, nil)

		err := logsCmd.runE(logsCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   bool
		expected string
	}{
		{
			name:     "success status",
			status:   true,
			expected: "SUCCESS",
		},
		{
			name:     "failed status",
			status:   false,
			expected: "FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
