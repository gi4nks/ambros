package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

type TemplateCommand struct {
	*BaseCommand
}

func NewTemplateCommand(logger *zap.Logger, repo RepositoryInterface) *TemplateCommand {
	tc := &TemplateCommand{}

	cmd := &cobra.Command{
		Use:   "template <action> [args...]",
		Short: "Manage command templates",
		Long: `Create, manage and execute command templates for frequently used commands.

Actions:
  save <name> <command>    Save a command as a template
  run <name> [args...]     Execute a template with optional arguments
  list                     List all available templates
  delete <name>           Delete a template
  show <name>             Show template details

Examples:
  ambros template save "deploy" "kubectl apply -f deployment.yaml"
  ambros template run "deploy"
  ambros template save "build" "go build -o bin/app"
  ambros template run "build"
  ambros template list`,
		Args: cobra.MinimumNArgs(1),
		RunE: tc.runE,
	}

	tc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	tc.cmd = cmd
	return tc
}

func (tc *TemplateCommand) runE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand, "action is required", nil)
	}

	action := args[0]
	tc.logger.Debug("Template command invoked",
		zap.String("action", action),
		zap.Strings("args", args[1:]))

	switch action {
	case "save":
		return tc.saveTemplate(args[1:])
	case "run":
		return tc.runTemplate(args[1:])
	case "list":
		return tc.listTemplates()
	case "delete":
		return tc.deleteTemplate(args[1:])
	case "show":
		return tc.showTemplate(args[1:])
	default:
		return errors.NewError(errors.ErrInvalidCommand,
			fmt.Sprintf("unknown action: %s", action), nil)
	}
}

func (tc *TemplateCommand) saveTemplate(args []string) error {
	if len(args) < 2 {
		return errors.NewError(errors.ErrInvalidCommand,
			"usage: template save <name> <command>", nil)
	}

	name := args[0]
	command := strings.Join(args[1:], " ")

	// Parse command parts
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return errors.NewError(errors.ErrInvalidCommand, "empty command", nil)
	}

	// Create command as template storage
	templateCmd := &models.Command{
		Entity: models.Entity{
			ID:        fmt.Sprintf("TPL-%s-%d", name, time.Now().UnixNano()),
			CreatedAt: time.Now(),
		},
		Name:      cmdParts[0],
		Arguments: cmdParts[1:],
		Status:    true,
		Output:    fmt.Sprintf("Template: %s", command),
		Category:  "template",
		Tags:      []string{"template", name},
	}
	templateCmd.TerminatedAt = time.Now()

	// Store as a command with template tag
	if err := tc.repository.Put(context.Background(), *templateCmd); err != nil {
		tc.logger.Error("Failed to save template",
			zap.String("name", name),
			zap.String("command", command),
			zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite,
			"failed to save template", err)
	}

	color.Green("✅ Template '%s' saved successfully", name)
	fmt.Printf("Command: %s\n", command)

	tc.logger.Info("Template saved",
		zap.String("name", name),
		zap.String("command", command),
		zap.String("templateId", templateCmd.ID))

	return nil
}

func (tc *TemplateCommand) runTemplate(args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand,
			"usage: template run <name> [args...]", nil)
	}

	name := args[0]
	templateArgs := args[1:]

	// Find template by searching for commands with template tag and matching name
	commands, err := tc.repository.SearchByTag("template")
	if err != nil {
		tc.logger.Error("Failed to search templates", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to search templates", err)
	}

	var templateCmd *models.Command
	for _, cmd := range commands {
		// Check if this command has the template name tag
		for _, tag := range cmd.Tags {
			if tag == name {
				templateCmd = &cmd
				break
			}
		}
		if templateCmd != nil {
			break
		}
	}

	if templateCmd == nil {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("template not found: %s", name), nil)
	}

	// Combine template command with additional args
	fullArgs := append(templateCmd.Arguments, templateArgs...)

	color.Yellow("🚀 Running template: %s", name)
	fmt.Printf("Command: %s %s\n", templateCmd.Name, strings.Join(fullArgs, " "))

	// Create a run command instance and execute
	runCmd := NewRunCommand(tc.logger, tc.repository)
	runCmd.opts.store = true // Always store template executions
	runCmd.opts.tag = []string{"template", name}

	// Execute the template command by calling executeCommand and handling output
	output, errorMsg, success, err := runCmd.executeCommand(templateCmd.Name, fullArgs)
	if err != nil {
		return err
	}

	// Display output
	if output != "" {
		fmt.Print(output)
	}
	if errorMsg != "" {
		fmt.Fprint(os.Stderr, errorMsg)
	}

	// Store the execution if successful
	if runCmd.opts.store {
		newCmd := &models.Command{
			Entity: models.Entity{
				ID:        fmt.Sprintf("CMD-%d", time.Now().UnixNano()),
				CreatedAt: time.Now(),
			},
			Name:      templateCmd.Name,
			Arguments: fullArgs,
			Status:    success,
			Output:    output,
			Error:     errorMsg,
			Tags:      runCmd.opts.tag,
			Category:  runCmd.opts.category,
		}
		newCmd.TerminatedAt = time.Now()

		if err := tc.repository.Put(context.Background(), *newCmd); err != nil {
			tc.logger.Warn("Failed to store template execution", zap.Error(err))
		} else {
			duration := newCmd.TerminatedAt.Sub(newCmd.CreatedAt)
			if success {
				color.Green("[%s] ✅ Success (%v)", newCmd.ID, duration.Round(time.Millisecond))
			} else {
				color.Red("[%s] ❌ Failed (%v)", newCmd.ID, duration.Round(time.Millisecond))
			}
		}
	}

	return nil
}

func (tc *TemplateCommand) listTemplates() error {
	// Get all commands with template tag
	commands, err := tc.repository.SearchByTag("template")
	if err != nil {
		tc.logger.Error("Failed to retrieve templates", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to retrieve templates", err)
	}

	// Filter and group by template name
	templates := make(map[string]*models.Command)
	for _, cmd := range commands {
		if cmd.Category == "template" {
			// Extract template name from tags (skip first "template" tag)
			for i, tag := range cmd.Tags {
				if tag == "template" && i+1 < len(cmd.Tags) {
					templateName := cmd.Tags[i+1]
					if _, exists := templates[templateName]; !exists {
						templates[templateName] = &cmd
					}
					break
				}
			}
		}
	}

	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}

	color.Cyan("📋 Available Templates (%d):\n", len(templates))

	i := 1
	for name, template := range templates {
		fmt.Printf("%d. ", i)
		color.Green("%s", name)
		fmt.Printf("\n   Command: %s %s\n", template.Name, strings.Join(template.Arguments, " "))
		fmt.Printf("   Created: %s\n", template.CreatedAt.Format("2006-01-02 15:04:05"))
		if i < len(templates) {
			fmt.Println()
		}
		i++
	}

	return nil
}

func (tc *TemplateCommand) deleteTemplate(args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand,
			"usage: template delete <name>", nil)
	}

	name := args[0]

	// Find and delete the template command
	commands, err := tc.repository.SearchByTag("template")
	if err != nil {
		tc.logger.Error("Failed to search templates", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to search templates", err)
	}

	found := false
	for _, cmd := range commands {
		// Check if this is the template we want to delete
		for _, tag := range cmd.Tags {
			if tag == name && cmd.Category == "template" {
				// Delete the template from repository
				if err := tc.repository.Delete(cmd.ID); err != nil {
					return fmt.Errorf("failed to delete template: %w", err)
				}
				color.Green("🗑️  Template '%s' deleted successfully", name)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("template not found: %s", name), nil)
	}

	tc.logger.Info("Template deleted", zap.String("name", name))
	return nil
}

func (tc *TemplateCommand) showTemplate(args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand,
			"usage: template show <name>", nil)
	}

	name := args[0]

	// Find template by searching for commands with template tag
	commands, err := tc.repository.SearchByTag("template")
	if err != nil {
		tc.logger.Error("Failed to search templates", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to search templates", err)
	}

	var templateCmd *models.Command
	for _, cmd := range commands {
		// Check if this command has the template name tag
		for _, tag := range cmd.Tags {
			if tag == name && cmd.Category == "template" {
				templateCmd = &cmd
				break
			}
		}
		if templateCmd != nil {
			break
		}
	}

	if templateCmd == nil {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("template not found: %s", name), nil)
	}

	color.Cyan("📄 Template Details:")
	fmt.Printf("Name: ")
	color.Green("%s", name)
	fmt.Printf("\nID: %s\n", templateCmd.ID)
	fmt.Printf("Command: %s %s\n", templateCmd.Name, strings.Join(templateCmd.Arguments, " "))
	fmt.Printf("Created: %s\n", templateCmd.CreatedAt.Format("2006-01-02 15:04:05"))
	if len(templateCmd.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(templateCmd.Tags, ", "))
	}

	return nil
}

func (tc *TemplateCommand) Command() *cobra.Command {
	return tc.cmd
}
