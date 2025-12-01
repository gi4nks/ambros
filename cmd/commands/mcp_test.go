package commands

import (
	"context"
	"testing"
	"time"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMCPCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("test command structure", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mcpCmd := NewMCPCommand(logger, mockRepo, nil)

		assert.Equal(t, "mcp", mcpCmd.cmd.Use)
		assert.Equal(t, "Start MCP server exposing Ambros tools", mcpCmd.cmd.Short)
		assert.NotNil(t, mcpCmd.Command())
		assert.Equal(t, mcpCmd.cmd, mcpCmd.Command())
	})
}

// Helper function to create a CallToolRequest with arguments
func newCallToolRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestMCPCommand_HandleLast(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("returns recent commands with default limit", func(t *testing.T) {
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

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{})

		result, err := mcpCmd.handleLast(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns recent commands with custom limit", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Now(),
				},
				Name:   "ls",
				Status: true,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"limit": float64(5),
		})

		result, err := mcpCmd.handleLast(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("filters failed only commands", func(t *testing.T) {
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

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"failed_only": true,
		})

		result, err := mcpCmd.handleLast(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(nil, assert.AnError)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{})

		result, err := mcpCmd.handleLast(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})
}

func TestMCPCommand_HandleSearch(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("searches commands by query", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Now(),
				},
				Name:   "docker",
				Status: true,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"query": "docker",
		})

		result, err := mcpCmd.handleSearch(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("searches commands by tag", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Now(),
				},
				Name:   "kubectl",
				Tags:   []string{"deployment"},
				Status: true,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"tag": "deployment",
		})

		result, err := mcpCmd.handleSearch(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("searches commands by status", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Now(),
				},
				Name:   "test",
				Status: false,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"status": "failed",
		})

		result, err := mcpCmd.handleSearch(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})
}

func TestMCPCommand_HandleAnalytics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("returns summary analytics", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		now := time.Now()
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{
					ID:           "cmd-1",
					CreatedAt:    now.Add(-1 * time.Hour),
					TerminatedAt: now.Add(-1*time.Hour + 5*time.Second),
				},
				Name:   "ls",
				Status: true,
			},
			{
				Entity: models.Entity{
					ID:           "cmd-2",
					CreatedAt:    now.Add(-2 * time.Hour),
					TerminatedAt: now.Add(-2*time.Hour + 3*time.Second),
				},
				Name:   "pwd",
				Status: false,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"action": "summary",
		})

		result, err := mcpCmd.handleAnalytics(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns most-used analytics", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{ID: "cmd-1", CreatedAt: time.Now()},
				Name:   "git",
				Status: true,
			},
			{
				Entity: models.Entity{ID: "cmd-2", CreatedAt: time.Now()},
				Name:   "git",
				Status: true,
			},
			{
				Entity: models.Entity{ID: "cmd-3", CreatedAt: time.Now()},
				Name:   "ls",
				Status: true,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"action": "most-used",
		})

		result, err := mcpCmd.handleAnalytics(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns failures analytics", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{
			{
				Entity: models.Entity{ID: "cmd-1", CreatedAt: time.Now()},
				Name:   "test",
				Status: false,
			},
		}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"action": "failures",
		})

		result, err := mcpCmd.handleAnalytics(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error for unknown action", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommands := []models.Command{}

		mockRepo.On("GetAllCommands").Return(expectedCommands, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"action": "invalid",
		})

		result, err := mcpCmd.handleAnalytics(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})
}

func TestMCPCommand_HandleOutput(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("returns command output", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommand := &models.Command{
			Entity: models.Entity{
				ID:        "cmd-1",
				CreatedAt: time.Now(),
			},
			Name:   "ls",
			Output: "file1.txt\nfile2.txt",
			Status: true,
		}

		mockRepo.On("Get", "cmd-1").Return(expectedCommand, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"id": "cmd-1",
		})

		result, err := mcpCmd.handleOutput(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{})

		result, err := mcpCmd.handleOutput(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("returns error when command not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "nonexistent").Return(nil, assert.AnError)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"id": "nonexistent",
		})

		result, err := mcpCmd.handleOutput(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})
}

func TestMCPCommand_HandleCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("returns command details as JSON", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		expectedCommand := &models.Command{
			Entity: models.Entity{
				ID:        "cmd-1",
				CreatedAt: time.Now(),
			},
			Name:      "docker",
			Arguments: []string{"ps", "-a"},
			Tags:      []string{"containers"},
			Status:    true,
		}

		mockRepo.On("Get", "cmd-1").Return(expectedCommand, nil)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{
			"id": "cmd-1",
		})

		result, err := mcpCmd.handleCommand(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		mcpCmd := NewMCPCommand(logger, mockRepo, nil)
		request := newCallToolRequest(map[string]any{})

		result, err := mcpCmd.handleCommand(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

func TestMCPCommand_FormatHelpers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	mcpCmd := NewMCPCommand(logger, mockRepo, nil)

	t.Run("formatStatus returns correct strings", func(t *testing.T) {
		assert.Equal(t, "Success", mcpCmd.formatStatus(true))
		assert.Equal(t, "Failed", mcpCmd.formatStatus(false))
	})

	t.Run("formatCommandsResponse handles empty list", func(t *testing.T) {
		result := mcpCmd.formatCommandsResponse([]models.Command{})
		assert.Equal(t, "No commands found.", result)
	})

	t.Run("formatCommandsResponse formats commands", func(t *testing.T) {
		commands := []models.Command{
			{
				Entity: models.Entity{
					ID:        "cmd-1",
					CreatedAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Name:      "ls",
				Arguments: []string{"-la"},
				Tags:      []string{"test"},
				Category:  "filesystem",
				Status:    true,
			},
		}

		result := mcpCmd.formatCommandsResponse(commands)
		assert.Contains(t, result, "Found 1 command(s)")
		assert.Contains(t, result, "ls -la")
		assert.Contains(t, result, "cmd-1")
		assert.Contains(t, result, "Success")
		assert.Contains(t, result, "test")
		assert.Contains(t, result, "filesystem")
	})

	t.Run("formatSummaryAnalytics handles empty list", func(t *testing.T) {
		result := mcpCmd.formatSummaryAnalytics([]models.Command{})
		assert.Equal(t, "No commands found in history.", result)
	})

	t.Run("formatFailuresAnalytics handles no failures", func(t *testing.T) {
		commands := []models.Command{
			{
				Entity: models.Entity{ID: "cmd-1"},
				Name:   "ls",
				Status: true,
			},
		}

		result := mcpCmd.formatFailuresAnalytics(commands)
		assert.Contains(t, result, "No command failures found")
	})
}
