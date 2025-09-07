package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/errors"
	"github.com/gi4nks/ambros/internal/models"
)

type RunCommand struct {
	*BaseCommand
	opts     RunOptions
	wrapper  *CommandWrapper
	exitFunc func(int) // For testable exit behavior
}

type RunOptions struct {
	store    bool
	tag      []string
	category string
	template string
	dryRun   bool
}

func NewRunCommand(logger *zap.Logger, repo RepositoryInterface) *RunCommand {
	rc := &RunCommand{
		wrapper:  NewCommandWrapper(logger, repo),
		exitFunc: os.Exit, // Default to os.Exit
	}

	cmd := &cobra.Command{
		Use:   "run [flags] [--] <command> [args...]",
		Short: "Run a command and optionally store its execution details",
		Long: `Execute a command and capture its output, error, and execution details.
Optionally store the execution for future reference.

The -- separator is optional for simple commands without flags, but required when the 
command you're running has flags that might conflict with ambros flags.

Examples:
  ambros run echo "hello"                  # Simple command, no -- needed
  ambros run -- ls -la                    # Required: ls flags would conflict
  ambros run --store -- echo "hello"      # Store the command execution
  ambros run -t dev,test -- make build    # Run with tags
  ambros run --dry-run -- rm -rf /        # Show what would be executed

Note: If you get "unknown flag" errors, add -- before your command.`,
		RunE: rc.runE,
		Args: cobra.MinimumNArgs(1),
	}

	rc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	rc.cmd = cmd
	rc.setupFlags(cmd)
	return rc
}

func (rc *RunCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&rc.opts.store, "store", "s", true,
		"Store the command execution details")
	cmd.Flags().StringSliceVarP(&rc.opts.tag, "tag", "t", nil,
		"Add tags to the command")
	cmd.Flags().StringVarP(&rc.opts.category, "category", "c", "",
		"Assign a category to the command")
	cmd.Flags().StringVarP(&rc.opts.template, "template", "p", "",
		"Use a command template")
	cmd.Flags().BoolVar(&rc.opts.dryRun, "dry-run", false,
		"Show what would be executed without running")
}

func (rc *RunCommand) runE(cmd *cobra.Command, args []string) error {
	rc.logger.Debug("Run command invoked",
		zap.Strings("args", args),
		zap.String("template", rc.opts.template),
		zap.Bool("dryRun", rc.opts.dryRun),
		zap.Bool("store", rc.opts.store))

	// Check if command was provided
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand,
			"No command specified. Use: ambros run [flags] -- <command> [args...]", nil)
	}

	// Handle template if specified
	if rc.opts.template != "" {
		template, err := rc.repository.GetTemplate(rc.opts.template)
		if err != nil {
			rc.logger.Error("Template not found",
				zap.String("template", rc.opts.template),
				zap.Error(err))
			return errors.NewError(errors.ErrCommandNotFound,
				fmt.Sprintf("template not found: %s", rc.opts.template), err)
		}
		args = template.BuildCommand(args)
		rc.logger.Debug("Template applied",
			zap.String("template", rc.opts.template),
			zap.Strings("resultArgs", args))
	}

	if len(args) == 0 {
		rc.logger.Error("No command specified")
		return errors.NewError(errors.ErrInvalidCommand, "no command specified", nil)
	}

	commandName := args[0]
	commandArgs := args[1:]

	// Create command record
	command := &models.Command{
		Entity: models.Entity{
			ID:        rc.generateCommandID(),
			CreatedAt: time.Now(),
		},
		Name:      commandName,
		Arguments: commandArgs,
		Tags:      rc.opts.tag,
		Category:  rc.opts.category,
	}

	if rc.opts.dryRun {
		// User output (keep fmt)
		fmt.Printf("Would execute: %s %s\n", commandName, strings.Join(commandArgs, " "))
		if len(rc.opts.tag) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(rc.opts.tag, ", "))
		}
		if rc.opts.category != "" {
			fmt.Printf("Category: %s\n", rc.opts.category)
		}

		rc.logger.Info("Dry run completed",
			zap.String("command", commandName),
			zap.Strings("args", commandArgs))
		return nil
	}

	// Execute the command
	output, errorMsg, success, err := rc.executeCommand(commandName, commandArgs)
	if err != nil {
		rc.logger.Error("Command execution failed",
			zap.String("command", commandName),
			zap.Strings("args", commandArgs),
			zap.Error(err))
		return errors.NewError(errors.ErrExecutionFailed, "failed to execute command", err)
	}

	// Update command with results
	command.TerminatedAt = time.Now()
	command.Status = success
	command.Output = output
	command.Error = errorMsg

	// Display output to user (keep fmt for user output)
	if output != "" {
		fmt.Print(output)
	}
	if errorMsg != "" {
		fmt.Fprint(os.Stderr, errorMsg)
	}

	// Store if requested
	if rc.opts.store {
		if err := rc.wrapper.FinalizeCommand(command); err != nil {
			rc.logger.Warn("Failed to store command execution",
				zap.Error(err),
				zap.String("commandId", command.ID))
		} else {
			// Show user-friendly confirmation message with color
			if success {
				color.Green("[%s]", command.ID)
			} else {
				color.Red("[%s]", command.ID)
			}
		}
	}

	// Log execution summary (only in debug mode)
	rc.logger.Debug("Command execution completed",
		zap.String("commandId", command.ID),
		zap.String("command", commandName),
		zap.Strings("args", commandArgs),
		zap.Bool("success", success),
		zap.Duration("duration", command.TerminatedAt.Sub(command.CreatedAt)),
		zap.Bool("stored", rc.opts.store))

	// Exit with the same code as the executed command if it failed
	if !success {
		rc.logger.Debug("Command failed, exiting with code 1",
			zap.String("commandId", command.ID))
		rc.exitFunc(1)
	}

	return nil
}

func (rc *RunCommand) executeCommand(name string, args []string) (string, string, bool, error) {
	cmd := exec.Command(name, args...)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	success := err == nil
	var errorMsg string

	if err != nil {
		errorMsg = err.Error()
	}

	return string(output), errorMsg, success, nil
}

func (rc *RunCommand) generateCommandID() string {
	// Simple ID generation - could be enhanced
	return fmt.Sprintf("CMD-%d", time.Now().UnixNano())
}

func (rc *RunCommand) Command() *cobra.Command {
	return rc.cmd
}
