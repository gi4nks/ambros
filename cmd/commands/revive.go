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

type ReviveCommand struct {
	*BaseCommand
	store    bool
	dryRun   bool
	exitFunc func(int) // For testable exit behavior
}

func NewReviveCommand(logger *zap.Logger, repo RepositoryInterface) *ReviveCommand {
	rc := &ReviveCommand{
		exitFunc: os.Exit, // Default to os.Exit
	}

	cmd := &cobra.Command{
		Use:   "revive <command-id>",
		Short: "Re-execute a previously stored command",
		Long: `Re-execute a command from your history by its ID.
The command will be executed with the same arguments as originally stored.
Optionally store the new execution results.

Examples:
  ambros revive CMD-123           # Re-execute command CMD-123
  ambros revive --store CMD-456   # Re-execute and store new results
  ambros revive --dry-run CMD-789 # Show what would be executed`,
		Args: cobra.ExactArgs(1),
		RunE: rc.runE,
	}

	rc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	rc.cmd = cmd
	rc.setupFlags(cmd)
	return rc
}

func (rc *ReviveCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&rc.store, "store", "s", false, "Store the new execution results")
	cmd.Flags().BoolVar(&rc.dryRun, "dry-run", false, "Show what would be executed without running")
}

func (rc *ReviveCommand) runE(cmd *cobra.Command, args []string) error {
	commandId := args[0]

	rc.logger.Debug("Revive command invoked",
		zap.String("commandId", commandId),
		zap.Bool("store", rc.store),
		zap.Bool("dryRun", rc.dryRun))

	// Find the stored command
	storedCommand, err := rc.repository.Get(commandId)
	if err != nil {
		rc.logger.Error("Command not found",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("command not found: %s", commandId), err)
	}

	rc.logger.Info("Found stored command",
		zap.String("commandId", commandId),
		zap.String("name", storedCommand.Name),
		zap.Strings("arguments", storedCommand.Arguments))

	if rc.dryRun {
		// User output (keep fmt for dry run display)
		fmt.Printf("Would re-execute: %s %s\n",
			storedCommand.Name, strings.Join(storedCommand.Arguments, " "))
		fmt.Printf("Original execution: %s\n",
			storedCommand.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Original status: %s\n",
			rc.formatStatus(storedCommand.Status))

		rc.logger.Info("Dry run completed", zap.String("commandId", commandId))
		return nil
	}

	// Execute the command
	output, errorMsg, success, err := rc.executeCommand(storedCommand.Name, storedCommand.Arguments)
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

	// Store new execution if requested
	if rc.store {
		newCommand := &models.Command{
			Entity: models.Entity{
				ID:        rc.generateCommandID(),
				CreatedAt: time.Now(),
			},
			Name:      storedCommand.Name,
			Arguments: storedCommand.Arguments,
			Tags:      append(storedCommand.Tags, "revived"),
			Category:  storedCommand.Category,
			Status:    success,
			Output:    output,
			Error:     errorMsg,
		}
		newCommand.TerminatedAt = time.Now()

		if err := rc.repository.Put(context.Background(), *newCommand); err != nil {
			rc.logger.Warn("Failed to store revived command execution",
				zap.Error(err),
				zap.String("originalCommandId", commandId),
				zap.String("newCommandId", newCommand.ID))
		} else {
			rc.logger.Debug("Revived command execution stored",
				zap.String("originalCommandId", commandId),
				zap.String("newCommandId", newCommand.ID),
				zap.Bool("success", success))
		}
	}

	rc.logger.Info("Command revive completed",
		zap.String("originalCommandId", commandId),
		zap.String("command", storedCommand.Name),
		zap.Strings("args", storedCommand.Arguments),
		zap.Bool("success", success),
		zap.Bool("stored", rc.store))

	// Exit with the same code as the executed command if it failed
	if !success {
		rc.logger.Debug("Revived command failed, exiting with code 1",
			zap.String("commandId", commandId))
		rc.exitFunc(1)
	}

	return nil
}

func (rc *ReviveCommand) executeCommand(name string, args []string) (string, string, bool, error) {
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

func (rc *ReviveCommand) generateCommandID() string {
	return fmt.Sprintf("CMD-%d", time.Now().UnixNano())
}

func (rc *ReviveCommand) formatStatus(status bool) string {
	if status {
		return "Success"
	}
	return "Failed"
}

func (rc *ReviveCommand) Command() *cobra.Command {
	return rc.cmd
}
