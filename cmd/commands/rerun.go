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

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

// RerunCommand unifies recall and revive behaviors
type RerunCommand struct {
	*BaseCommand
	history  bool
	store    bool
	dryRun   bool
	tag      string // tag to append when storing
	exitFunc func(int)
}

// NewRerunCommand creates a new rerun command that subsumes recall and revive
func NewRerunCommand(logger *zap.Logger, repo RepositoryInterface) *RerunCommand {
	rc := &RerunCommand{exitFunc: os.Exit}

	cmd := &cobra.Command{
		Use:   "rerun <command-id>",
		Short: "Re-execute a previously stored command (recall/revive unified)",
		Long: `Re-execute a command from your history by its ID.
Can optionally pull from history and store the new execution. This command
unifies the previous 'recall' and 'revive' behavior.

Examples:
  ambros rerun CMD-123           # Re-execute command CMD-123
  ambros rerun -y CMD-456        # Re-execute from history
  ambros rerun -s CMD-456        # Re-execute and store new results
  ambros rerun --dry-run CMD-789 # Show what would be executed`,
		Args: cobra.ExactArgs(1),
		RunE: rc.runE,
	}

	rc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	rc.cmd = cmd
	rc.setupFlags(cmd)
	return rc
}

func (rc *RerunCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&rc.history, "history", "y", false, "Recall command from history")
	cmd.Flags().BoolVarP(&rc.store, "store", "s", false, "Store the new execution results")
	cmd.Flags().BoolVar(&rc.dryRun, "dry-run", false, "Show what would be executed without running")
	cmd.Flags().StringVar(&rc.tag, "tag", "", "Tag to append when storing the rerun result")
}

func (rc *RerunCommand) runE(cmd *cobra.Command, args []string) error {
	commandId := args[0]

	rc.logger.Debug("Rerun command invoked",
		zap.String("commandId", commandId),
		zap.Bool("history", rc.history),
		zap.Bool("store", rc.store),
		zap.Bool("dryRun", rc.dryRun))

	// Find the stored command
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
		fmt.Printf("Would re-execute: %s %s\n", stored.Name, strings.Join(stored.Arguments, " "))
		fmt.Printf("Original execution: %s\n", stored.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Original status: %s\n", rc.formatStatus(stored.Status))
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

	if output != "" {
		fmt.Print(output)
	}
	if errorMsg != "" {
		fmt.Fprint(os.Stderr, errorMsg)
	}

	// Store results if requested
	if rc.store {
		// Determine tag to apply to stored rerun result
		tag := "rerun"
		if rc.tag != "" {
			tag = rc.tag
		} else if len(stored.Tags) > 0 {
			// preserve a sense of the original intent if present
			tag = stored.Tags[0]
		}

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
			Tags:      append(stored.Tags, tag),
			Category:  stored.Category,
		}
		newCommand.TerminatedAt = time.Now()

		if err := rc.repository.Put(context.Background(), *newCommand); err != nil {
			rc.logger.Warn("Failed to store rerun execution",
				zap.Error(err),
				zap.String("originalCommandId", commandId),
				zap.String("newCommandId", newCommand.ID))
		} else {
			rc.logger.Debug("Rerun execution stored",
				zap.String("originalCommandId", commandId),
				zap.String("newCommandId", newCommand.ID),
				zap.Bool("success", success))
		}
	}

	rc.logger.Info("Command rerun completed",
		zap.String("originalCommandId", commandId),
		zap.String("command", stored.Name),
		zap.Strings("args", stored.Arguments),
		zap.Bool("success", success),
		zap.Bool("stored", rc.store))

	if !success {
		rc.logger.Debug("Rerun failed, exiting with code 1",
			zap.String("commandId", commandId))
		rc.exitFunc(1)
	}

	return nil
}

func (rc *RerunCommand) executeCommand(name string, args []string) (string, string, bool, error) {
	if _, err := ResolveCommandPath(name); err != nil {
		return "", err.Error(), false, err
	}
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	success := err == nil
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	return string(output), errorMsg, success, nil
}

func (rc *RerunCommand) generateCommandID() string {
	return fmt.Sprintf("CMD-%d", time.Now().UnixNano())
}

func (rc *RerunCommand) formatStatus(status bool) string {
	if status {
		return "Success"
	}
	return "Failed"
}

func (rc *RerunCommand) Command() *cobra.Command { return rc.cmd }
