package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/errors"
)

type LogsCommand struct {
	*BaseCommand
	since      string
	failedOnly bool
}

func NewLogsCommand(logger *zap.Logger, repo RepositoryInterface) *LogsCommand {
	lc := &LogsCommand{}

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show command execution logs",
		Long:  `Display the execution logs of stored commands, with optional filtering.`,
		RunE:  lc.runE,
	}

	lc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	lc.cmd = cmd
	lc.setupFlags(cmd)
	return lc
}

func (lc *LogsCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&lc.since, "since", "s", "", "Show logs since timestamp (e.g. 2024-01-01)")
	cmd.Flags().BoolVarP(&lc.failedOnly, "failed", "f", false, "Show only failed commands")
}

func (lc *LogsCommand) runE(cmd *cobra.Command, args []string) error {
	lc.logger.Debug("Logs command invoked",
		zap.String("since", lc.since),
		zap.Bool("failedOnly", lc.failedOnly))

	commands, err := lc.repository.GetAllCommands()
	if err != nil {
		lc.logger.Error("Failed to retrieve commands", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead, "failed to retrieve commands", err)
	}

	var sinceTime time.Time
	if lc.since != "" {
		sinceTime, err = time.Parse("2006-01-02", lc.since)
		if err != nil {
			lc.logger.Error("Invalid date format",
				zap.String("since", lc.since),
				zap.Error(err))
			return errors.NewError(errors.ErrInvalidCommand, "invalid date format", err)
		}
	}

	displayedCount := 0
	for _, command := range commands {
		if lc.since != "" && command.CreatedAt.Before(sinceTime) {
			continue
		}

		if lc.failedOnly && command.Status {
			continue
		}

		// Display to user (keep fmt for user output)
		fmt.Printf("[%s] %s %s - Status: %s\n",
			command.CreatedAt.Format("2006-01-02 15:04:05"),
			command.Name,
			fmt.Sprintf("(ID: %s)", command.ID),
			formatStatus(command.Status))

		displayedCount++
	}

	lc.logger.Info("Logs command completed",
		zap.Int("totalCommands", len(commands)),
		zap.Int("displayedCommands", displayedCount),
		zap.String("since", lc.since),
		zap.Bool("failedOnly", lc.failedOnly))

	return nil
}

func formatStatus(status bool) string {
	if status {
		return "SUCCESS"
	}
	return "FAILED"
}

func (lc *LogsCommand) Command() *cobra.Command {
	return lc.cmd
}
