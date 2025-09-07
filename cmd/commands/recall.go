package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/errors"
	"github.com/gi4nks/ambros/internal/models"
)

// RecallCommand represents the recall command
type RecallCommand struct {
	*BaseCommand
	history  bool
	store    bool
	dryRun   bool
	exitFunc func(int) // For testable exit behavior
}

// NewRecallCommand creates a new recall command
func NewRecallCommand(logger *zap.Logger, repo RepositoryInterface) *RecallCommand {
	rc := &RecallCommand{
		exitFunc: os.Exit, // Default to os.Exit
	}

	cmd := &cobra.Command{
		Use:   "recall <command-id>",
		Short: "Recall and execute a stored command",
		Long: `Recall and execute a previously stored command by its ID.
Can retrieve commands from both recent and historical storage.
Optionally store the new execution results.

Examples:
  ambros recall CMD-123          # Recall and execute command CMD-123
  ambros recall -y CMD-456       # Recall from history
  ambros recall -s CMD-789       # Store the new execution results
  ambros recall --dry-run CMD-123 # Show what would be executed`,
		Args: cobra.ExactArgs(1),
		RunE: rc.runE,
	}

	rc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	rc.cmd = cmd
	rc.setupFlags(cmd)
	return rc
}

func (rc *RecallCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&rc.history, "history", "y", false, "Recall command from history")
	cmd.Flags().BoolVarP(&rc.store, "store", "s", false, "Store the execution results")
	cmd.Flags().BoolVar(&rc.dryRun, "dry-run", false, "Show what would be executed without running")
}

func (rc *RecallCommand) runE(cmd *cobra.Command, args []string) error {
	commandId := args[0]

	rc.logger.Debug("Recall command invoked",
		zap.String("commandId", commandId),
		zap.Bool("history", rc.history),
		zap.Bool("store", rc.store),
		zap.Bool("dryRun", rc.dryRun))

	// Find the command
	stored, err := rc.repository.Get(commandId)
	if err != nil {
		rc.logger.Error("Command not found",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrCommandNotFound, "command not found: "+commandId, err)
	}

	rc.logger.Info("Found stored command",
		zap.String("commandId", commandId),
		zap.String("name", stored.Name),
		zap.Strings("arguments", stored.Arguments))

	if rc.dryRun {
		// User output (keep fmt for dry run display)
		fmt.Printf("Would recall and execute: %s %s\n",
			stored.Name, strings.Join(stored.Arguments, " "))
		fmt.Printf("Original execution: %s\n",
			stored.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Original status: %s\n",
			rc.formatStatus(stored.Status))
		if len(stored.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(stored.Tags, ", "))
		}

		rc.logger.Info("Dry run completed", zap.String("commandId", commandId))
		return nil
	}

	// Execute the stored command
	output, errorMsg, success, err := rc.executeCommand(stored.Name, stored.Arguments)
	if err != nil {
		rc.logger.Error("Command execution failed",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrExecutionFailed, "failed to execute command", err)
	}

	// Display output to user
	if output != "" {
		fmt.Print(output)
	}
	if errorMsg != "" {
		fmt.Fprint(os.Stderr, errorMsg)
	}

	// Store results if requested
	if rc.store {
		newCommand := &models.Command{
			Entity: models.Entity{
				ID:        rc.generateCommandID(),
				CreatedAt: time.Now(),
			},
			Name:      stored.Name,
			Arguments: stored.Arguments,
			Status:    success,
			Output:    output,
			Error:     errorMsg,
			Tags:      append(stored.Tags, "recalled"),
			Category:  stored.Category,
		}
		newCommand.TerminatedAt = time.Now()

		if err := rc.repository.Put(context.Background(), *newCommand); err != nil {
			rc.logger.Warn("Failed to store recalled command execution",
				zap.Error(err),
				zap.String("originalCommandId", commandId),
				zap.String("newCommandId", newCommand.ID))
		} else {
			rc.logger.Debug("Recalled command execution stored",
				zap.String("originalCommandId", commandId),
				zap.String("newCommandId", newCommand.ID),
				zap.Bool("success", success))
		}
	}

	rc.logger.Info("Command recall completed",
		zap.String("originalCommandId", commandId),
		zap.String("command", stored.Name),
		zap.Strings("args", stored.Arguments),
		zap.Bool("success", success),
		zap.Bool("stored", rc.store))

	// Exit with the same code as the executed command if it failed
	if !success {
		rc.logger.Debug("Recalled command failed, exiting with code 1",
			zap.String("commandId", commandId))
		rc.exitFunc(1)
	}

	return nil
}

func (rc *RecallCommand) executeCommand(name string, args []string) (string, string, bool, error) {
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

func (rc *RecallCommand) generateCommandID() string {
	return fmt.Sprintf("CMD-%d", time.Now().UnixNano())
}

func (rc *RecallCommand) formatStatus(status bool) string {
	if status {
		return "Success"
	}
	return "Failed"
}

func (rc *RecallCommand) Command() *cobra.Command {
	return rc.cmd
}
