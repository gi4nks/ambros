package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
)

type OutputCommand struct {
	*BaseCommand
	commandId string
}

func NewOutputCommand(logger *zap.Logger, repo RepositoryInterface) *OutputCommand {
	oc := &OutputCommand{}

	cmd := &cobra.Command{
		Use:   "output",
		Short: "Show the output of a stored command",
		Long:  `Display the output of a previously stored command by its ID.`,
		RunE:  oc.runE,
	}

	oc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	oc.cmd = cmd
	oc.setupFlags(cmd)
	return oc
}

func (oc *OutputCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&oc.commandId, "id", "i", "", "ID of the command to show output for")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		oc.logger.Error("Failed to mark flag as required", zap.Error(err))
	}
}

func (oc *OutputCommand) runE(cmd *cobra.Command, args []string) error {
	if oc.commandId == "" {
		return errors.NewError(errors.ErrInvalidCommand, "command ID is required", nil)
	}

	command, err := oc.repository.Get(oc.commandId)
	if err != nil {
		oc.logger.Error("Failed to retrieve command",
			zap.String("commandId", oc.commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("command not found: %s", oc.commandId), err)
	}

	// Display the output to user (keep fmt for user output)
	if command.Output != "" {
		fmt.Printf("Output for command %s:\n%s\n", oc.commandId, command.Output)
	} else {
		fmt.Printf("No output available for command %s\n", oc.commandId)
	}

	// Display error if present
	if command.Error != "" {
		fmt.Printf("Error for command %s:\n%s\n", oc.commandId, command.Error)
	}

	oc.logger.Debug("Command output retrieved successfully",
		zap.String("commandId", oc.commandId),
		zap.String("commandName", command.Name),
		zap.Bool("hasOutput", command.Output != ""),
		zap.Bool("hasError", command.Error != ""))

	return nil
}

func (oc *OutputCommand) Command() *cobra.Command {
	return oc.cmd
}
