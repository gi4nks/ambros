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

func TestTemplateCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("save template successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "echo" &&
				len(cmd.Arguments) == 1 && cmd.Arguments[0] == "hello" &&
				cmd.Category == "template" &&
				len(cmd.Tags) == 2 && cmd.Tags[0] == "template" && cmd.Tags[1] == "greeting" &&
				cmd.Status == true
		})).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "greeting", "echo", "hello"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("save template with complex command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "kubectl" &&
				len(cmd.Arguments) == 4 &&
				cmd.Arguments[0] == "apply" && cmd.Arguments[1] == "-f" &&
				cmd.Arguments[2] == "deployment.yaml" && cmd.Arguments[3] == "--namespace=prod" &&
				cmd.Category == "template" &&
				len(cmd.Tags) == 2 && cmd.Tags[0] == "template" && cmd.Tags[1] == "deploy"
		})).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "deploy", "kubectl", "apply", "-f", "deployment.yaml", "--namespace=prod"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("save template with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		// Missing command
		err := templateCmd.runE(templateCmd.cmd, []string{"save", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: template save <name> <command>")

		// Missing name and command
		err = templateCmd.runE(templateCmd.cmd, []string{"save"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: template save <name> <command>")

		mockRepo.AssertExpectations(t)
	})

	t.Run("save template with empty command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "test", ""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty command")

		mockRepo.AssertExpectations(t)
	})

	t.Run("save template with repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.Anything).Return(assert.AnError)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "test", "echo", "hello"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save template")
		mockRepo.AssertExpectations(t)
	})

	t.Run("list templates successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		templates := []models.Command{
			{
				Entity: models.Entity{
					ID:        "TPL-greeting-123",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Name:      "echo",
				Arguments: []string{"hello"},
				Category:  "template",
				Tags:      []string{"template", "greeting"},
			},
			{
				Entity: models.Entity{
					ID:        "TPL-deploy-456",
					CreatedAt: time.Date(2023, 1, 2, 14, 0, 0, 0, time.UTC),
				},
				Name:      "kubectl",
				Arguments: []string{"apply", "-f", "deployment.yaml"},
				Category:  "template",
				Tags:      []string{"template", "deploy"},
			},
		}

		mockRepo.On("SearchByTag", "template").Return(templates, nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list templates with no templates found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return([]models.Command{}, nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list templates with repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return(nil, assert.AnError)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve templates")
		mockRepo.AssertExpectations(t)
	})

	t.Run("show template successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		templates := []models.Command{
			{
				Entity: models.Entity{
					ID:        "TPL-greeting-123",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Name:      "echo",
				Arguments: []string{"hello", "world"},
				Category:  "template",
				Tags:      []string{"template", "greeting"},
				Output:    "Template: echo hello world",
			},
		}

		mockRepo.On("SearchByTag", "template").Return(templates, nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"show", "greeting"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("show template not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return([]models.Command{}, nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"show", "nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template not found: nonexistent")
		mockRepo.AssertExpectations(t)
	})

	t.Run("show template with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"show"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: template show <name>")

		mockRepo.AssertExpectations(t)
	})

	t.Run("show template with repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return(nil, assert.AnError)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"show", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to search templates")
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete template successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		templates := []models.Command{
			{
				Entity: models.Entity{
					ID: "TPL-greeting-123",
				},
				Name:     "echo",
				Category: "template",
				Tags:     []string{"template", "greeting"},
			},
		}

		mockRepo.On("SearchByTag", "template").Return(templates, nil)
		mockRepo.On("Delete", "TPL-greeting-123").Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"delete", "greeting"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete template not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return([]models.Command{}, nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"delete", "nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template not found: nonexistent")
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete template with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"delete"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: template delete <name>")

		mockRepo.AssertExpectations(t)
	})

	t.Run("delete template with repository search error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return(nil, assert.AnError)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"delete", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to search templates")
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete template with repository delete error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		templates := []models.Command{
			{
				Entity: models.Entity{
					ID: "TPL-greeting-123",
				},
				Name:     "echo",
				Category: "template",
				Tags:     []string{"template", "greeting"},
			},
		}

		mockRepo.On("SearchByTag", "template").Return(templates, nil)
		mockRepo.On("Delete", "TPL-greeting-123").Return(assert.AnError)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"delete", "greeting"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete template")
		mockRepo.AssertExpectations(t)
	})

	t.Run("run template successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		templates := []models.Command{
			{
				Entity: models.Entity{
					ID:        "TPL-greeting-123",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Name:      "echo",
				Arguments: []string{"hello"},
				Category:  "template",
				Tags:      []string{"template", "greeting"},
			},
		}

		mockRepo.On("SearchByTag", "template").Return(templates, nil)

		// Mock for storing the execution result
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "echo" &&
				len(cmd.Arguments) == 2 && cmd.Arguments[0] == "hello" && cmd.Arguments[1] == "world" &&
				len(cmd.Tags) >= 2 && cmd.Tags[0] == "template" && cmd.Tags[1] == "greeting"
		})).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"run", "greeting", "world"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("run template not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return([]models.Command{}, nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"run", "nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template not found: nonexistent")
		mockRepo.AssertExpectations(t)
	})

	t.Run("run template with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"run"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: template run <name> [args...]")

		mockRepo.AssertExpectations(t)
	})

	t.Run("run template with repository search error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("SearchByTag", "template").Return(nil, assert.AnError)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"run", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to search templates")
		mockRepo.AssertExpectations(t)
	})

	t.Run("unknown action", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown action: unknown")

		mockRepo.AssertExpectations(t)
	})

	t.Run("no action provided", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "action is required")

		mockRepo.AssertExpectations(t)
	})
}

func TestTemplateCommand_EdgeCases(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("save template with single word command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "pwd" &&
				len(cmd.Arguments) == 0 &&
				cmd.Category == "template" &&
				len(cmd.Tags) == 2 && cmd.Tags[0] == "template" && cmd.Tags[1] == "current-dir"
		})).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "current-dir", "pwd"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("save template with quoted arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "echo" &&
				len(cmd.Arguments) == 2 && cmd.Arguments[0] == "hello" && cmd.Arguments[1] == "world" &&
				cmd.Category == "template"
		})).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "quoted", "echo", "hello", "world"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list templates with mixed categories", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		commands := []models.Command{
			{
				Entity:   models.Entity{ID: "TPL-1"},
				Name:     "echo",
				Category: "template",
				Tags:     []string{"template", "greeting"},
			},
			{
				Entity:   models.Entity{ID: "CMD-1"},
				Name:     "ls",
				Category: "files",
				Tags:     []string{"template"}, // Has template tag but wrong category
			},
			{
				Entity:   models.Entity{ID: "TPL-2"},
				Name:     "kubectl",
				Category: "template",
				Tags:     []string{"template", "deploy"},
			},
		}

		mockRepo.On("SearchByTag", "template").Return(commands, nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("run template with no additional arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		templates := []models.Command{
			{
				Entity:    models.Entity{ID: "TPL-pwd-123"},
				Name:      "pwd",
				Arguments: []string{},
				Category:  "template",
				Tags:      []string{"template", "current-dir"},
			},
		}

		mockRepo.On("SearchByTag", "template").Return(templates, nil)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "pwd" && len(cmd.Arguments) == 0
		})).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"run", "current-dir"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("template with special characters in name", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Tags[1] == "my-template_v1.0" &&
				cmd.Category == "template"
		})).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "my-template_v1.0", "echo", "test"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTemplateCommand_CommandStructure(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		// Verify command is properly configured
		assert.Equal(t, "template", templateCmd.cmd.Use[:8])
		assert.NotEmpty(t, templateCmd.cmd.Short)
		assert.NotEmpty(t, templateCmd.cmd.Long)
		assert.NotNil(t, templateCmd.cmd.RunE)
	})

	t.Run("template ID generation format", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		var capturedCmd models.Command
		mockRepo.On("Put", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			capturedCmd = args.Get(1).(models.Command)
		}).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "test", "echo", "hello"})
		assert.NoError(t, err)

		// Verify ID format: TPL-{name}-{timestamp}
		assert.Contains(t, capturedCmd.ID, "TPL-test-")
		assert.True(t, len(capturedCmd.ID) > 10) // Should include timestamp

		mockRepo.AssertExpectations(t)
	})

	t.Run("template timing and metadata", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		var capturedCmd models.Command
		mockRepo.On("Put", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			capturedCmd = args.Get(1).(models.Command)
		}).Return(nil)

		templateCmd := NewTemplateCommand(logger, mockRepo, nil)
		startTime := time.Now()

		err := templateCmd.runE(templateCmd.cmd, []string{"save", "timing-test", "echo", "timing"})
		assert.NoError(t, err)

		endTime := time.Now()

		// Verify timing
		assert.True(t, capturedCmd.CreatedAt.After(startTime.Add(-1*time.Second)))
		assert.True(t, capturedCmd.CreatedAt.Before(endTime.Add(1*time.Second)))
		assert.True(t, capturedCmd.TerminatedAt.After(capturedCmd.CreatedAt) || capturedCmd.TerminatedAt.Equal(capturedCmd.CreatedAt))

		// Verify metadata
		assert.Equal(t, "echo", capturedCmd.Name)
		assert.Equal(t, []string{"timing"}, capturedCmd.Arguments)
		assert.Equal(t, "template", capturedCmd.Category)
		assert.Equal(t, []string{"template", "timing-test"}, capturedCmd.Tags)
		assert.True(t, capturedCmd.Status)
		assert.Contains(t, capturedCmd.Output, "Template: echo timing")

		mockRepo.AssertExpectations(t)
	})
}
