package commands

import (
	"testing"
	"time"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestSchedulerCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("list empty scheduled commands", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return([]models.Command{}, nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list scheduled commands", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		commands := []models.Command{
			{
				Entity: models.Entity{ID: "cmd1"},
				Name:   "echo",
				Schedule: &models.Schedule{
					CronExpr: "0 9 * * *",
					Enabled:  true,
					NextRun:  time.Now().Add(time.Hour),
				},
			},
			{
				Entity: models.Entity{ID: "cmd2"},
				Name:   "ls",
				// No schedule
			},
		}
		mockRepo.On("GetAllCommands").Return(commands, nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("add scheduled command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		command := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
		}
		mockRepo.On("Get", "cmd1").Return(command, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Schedule != nil && cmd.Schedule.CronExpr == "0 9 * * *"
		})).Return(nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"add", "cmd1", "0", "9", "*", "*", "*"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("add scheduled command with cron flag", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		command := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
		}
		mockRepo.On("Get", "cmd1").Return(command, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Schedule != nil && cmd.Schedule.CronExpr == "0 */2 * * *"
		})).Return(nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)
		schedulerCmd.cronExpr = "0 */2 * * *"

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"add", "cmd1"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("add scheduled command with interval", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		command := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
		}
		mockRepo.On("Get", "cmd1").Return(command, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Schedule != nil && cmd.Schedule.CronExpr == "0 * * * *"
		})).Return(nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)
		schedulerCmd.interval = "1h"

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"add", "cmd1"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("add scheduled command - not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "nonexistent").Return(nil, assert.AnError)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"add", "nonexistent", "0", "9", "*", "*", "*"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("remove scheduled command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		command := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
			Schedule: &models.Schedule{
				CronExpr: "0 9 * * *",
				Enabled:  true,
			},
		}
		mockRepo.On("Get", "cmd1").Return(command, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Schedule == nil
		})).Return(nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"remove", "cmd1"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("enable scheduled command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		command := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
			Schedule: &models.Schedule{
				CronExpr: "0 9 * * *",
				Enabled:  false,
			},
		}
		mockRepo.On("Get", "cmd1").Return(command, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Schedule != nil && cmd.Schedule.Enabled
		})).Return(nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"enable", "cmd1"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("disable scheduled command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		command := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
			Schedule: &models.Schedule{
				CronExpr: "0 9 * * *",
				Enabled:  true,
			},
		}
		mockRepo.On("Get", "cmd1").Return(command, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Schedule != nil && !cmd.Schedule.Enabled
		})).Return(nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"disable", "cmd1"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("show status", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		commands := []models.Command{
			{
				Entity:   models.Entity{ID: "cmd1"},
				Schedule: &models.Schedule{Enabled: true},
			},
			{
				Entity:   models.Entity{ID: "cmd2"},
				Schedule: &models.Schedule{Enabled: false},
			},
			{
				Entity: models.Entity{ID: "cmd3"},
				// No schedule
			},
		}
		mockRepo.On("GetAllCommands").Return(commands, nil)

		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"status"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unknown subcommand", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		err := schedulerCmd.runE(schedulerCmd.cmd, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand")
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		schedulerCmd := NewSchedulerCommand(logger, mockRepo)

		assert.Equal(t, "scheduler", schedulerCmd.cmd.Use)
		assert.Equal(t, "Manage scheduled command executions", schedulerCmd.cmd.Short)
		assert.NotNil(t, schedulerCmd.Command())
		assert.Equal(t, schedulerCmd.cmd, schedulerCmd.Command())

		// Test flags
		assert.NotNil(t, schedulerCmd.cmd.Flags().Lookup("enabled"))
		assert.NotNil(t, schedulerCmd.cmd.Flags().Lookup("interval"))
		assert.NotNil(t, schedulerCmd.cmd.Flags().Lookup("cron"))
	})
}

func TestSchedulerCommand_IntervalToCron(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	schedulerCmd := NewSchedulerCommand(logger, mockRepo)

	tests := []struct {
		interval string
		expected string
	}{
		{"1m", "* * * * *"},
		{"5m", "*/5 * * * *"},
		{"1h", "0 * * * *"},
		{"1d", "0 0 * * *"},
		{"custom", "custom"}, // Pass through
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			result := schedulerCmd.intervalToCron(tt.interval)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSchedulerCommand_ValidateCronExpr(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	schedulerCmd := NewSchedulerCommand(logger, mockRepo)

	tests := []struct {
		name     string
		cronExpr string
		wantErr  bool
	}{
		{"valid cron", "0 9 * * *", false},
		{"invalid cron - too few fields", "0 9 *", true},
		{"invalid cron - too many fields", "0 9 * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := schedulerCmd.validateCronExpr(tt.cronExpr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
