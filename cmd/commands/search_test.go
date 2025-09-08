package commands

import (
	"testing"
	"time"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSearchCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Sample commands for testing
	sampleCommands := []models.Command{
		{
			Entity: models.Entity{
				ID:        "cmd1",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"hello", "world"},
			Tags:      []string{"test", "output"},
			Category:  "utility",
			Status:    true,
		},
		{
			Entity: models.Entity{
				ID:        "cmd2",
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
			Name:      "ls",
			Arguments: []string{"-la", "/tmp"},
			Tags:      []string{"file", "list"},
			Category:  "file",
			Status:    false,
		},
		{
			Entity: models.Entity{
				ID:        "cmd3",
				CreatedAt: time.Now().Add(-25 * time.Hour), // Over 24h ago
			},
			Name:      "grep",
			Arguments: []string{"pattern", "file.txt"},
			Tags:      []string{"search", "text"},
			Category:  "utility",
			Status:    true,
		},
	}

	t.Run("search all commands", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with text query", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{"echo"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with tag filter", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.tag = []string{"test"}
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with category filter", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.category = "utility"
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with status filter", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.status = "success"
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with since filter", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.since = "24h"
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with limit", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.limit = 1
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with JSON output", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.format = "json"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with YAML output", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.format = "yaml"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search with invalid format", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(sampleCommands, nil)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.format = "invalid"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("GetAllCommands").Return(nil, assert.AnError)

		searchCmd := NewSearchCommand(logger, mockRepo)
		searchCmd.opts.format = "text"

		err := searchCmd.runE(searchCmd.cmd, []string{})
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		searchCmd := NewSearchCommand(logger, mockRepo)

		assert.Equal(t, "search [query]", searchCmd.cmd.Use)
		assert.Equal(t, "Search through command history", searchCmd.cmd.Short)
		assert.NotNil(t, searchCmd.Command())
		assert.Equal(t, searchCmd.cmd, searchCmd.Command())

		// Test flags
		assert.NotNil(t, searchCmd.cmd.Flags().Lookup("tag"))
		assert.NotNil(t, searchCmd.cmd.Flags().Lookup("category"))
		assert.NotNil(t, searchCmd.cmd.Flags().Lookup("since"))
		assert.NotNil(t, searchCmd.cmd.Flags().Lookup("status"))
		assert.NotNil(t, searchCmd.cmd.Flags().Lookup("limit"))
		assert.NotNil(t, searchCmd.cmd.Flags().Lookup("format"))
	})
}

func TestSearchCommand_FilterCommands(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	commands := []models.Command{
		{
			Entity: models.Entity{
				ID:        "cmd1",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"hello"},
			Tags:      []string{"test"},
			Category:  "utility",
			Status:    true,
		},
		{
			Entity: models.Entity{
				ID:        "cmd2",
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
			Name:      "ls",
			Arguments: []string{"-la"},
			Tags:      []string{"file"},
			Category:  "file",
			Status:    false,
		},
	}

	tests := []struct {
		name     string
		query    string
		opts     SearchOptions
		expected int
	}{
		{
			name:     "no filters",
			query:    "",
			opts:     SearchOptions{},
			expected: 2,
		},
		{
			name:     "text query match",
			query:    "echo",
			opts:     SearchOptions{},
			expected: 1,
		},
		{
			name:     "tag filter match",
			query:    "",
			opts:     SearchOptions{tag: []string{"test"}},
			expected: 1,
		},
		{
			name:     "category filter match",
			query:    "",
			opts:     SearchOptions{category: "utility"},
			expected: 1,
		},
		{
			name:     "status filter success",
			query:    "",
			opts:     SearchOptions{status: "success"},
			expected: 1,
		},
		{
			name:     "status filter failure",
			query:    "",
			opts:     SearchOptions{status: "failure"},
			expected: 1,
		},
		{
			name:     "since filter",
			query:    "",
			opts:     SearchOptions{since: "90m"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchCmd := NewSearchCommand(logger, mockRepo)
			searchCmd.opts = tt.opts

			filtered := searchCmd.filterCommands(commands, tt.query)
			assert.Equal(t, tt.expected, len(filtered))
		})
	}
}
