package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

type StoreCommand struct {
	*BaseCommand
	name        string
	description string
	tags        []string
	category    string
	force       bool
}

func NewStoreCommand(logger *zap.Logger, repo RepositoryInterface) *StoreCommand {
	sc := &StoreCommand{}

	cmd := &cobra.Command{
		Use:   "store [command...]",
		Short: "Store a command for future use",
		Long: `Store a command with optional metadata for easy retrieval later.
The command can be tagged, categorized, and given a description.

Examples:
  ambros store echo "hello world"                 # Store a simple command
  ambros store -n backup -t daily rsync -av src/ dest/  # Store with name and tag
  ambros store -c "file ops" -d "List files" ls -la     # Store with category and description
  ambros store --force existing-name new-command        # Overwrite existing stored command`,
		Args: cobra.MinimumNArgs(1),
		RunE: sc.runE,
	}

	sc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	sc.cmd = cmd
	sc.setupFlags(cmd)
	return sc
}

func (sc *StoreCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&sc.name, "name", "n", "", "Name for the stored command")
	cmd.Flags().StringVarP(&sc.description, "description", "d", "", "Description of the command")
	cmd.Flags().StringSliceVarP(&sc.tags, "tag", "t", nil, "Tags for the command")
	cmd.Flags().StringVarP(&sc.category, "category", "c", "", "Category for the command")
	cmd.Flags().BoolVar(&sc.force, "force", false, "Overwrite existing command with same name")
}

func (sc *StoreCommand) runE(cmd *cobra.Command, args []string) error {
	sc.logger.Debug("Store command invoked",
		zap.Strings("args", args),
		zap.String("name", sc.name),
		zap.String("description", sc.description),
		zap.Strings("tags", sc.tags),
		zap.String("category", sc.category),
		zap.Bool("force", sc.force))

	commandName := args[0]
	commandArgs := args[1:]

	// Generate command ID or use provided name
	var commandId string
	if sc.name != "" {
		commandId = sc.name
		// Check if command with this name already exists
		if !sc.force {
			if sc.commandExists(commandId) {
				return errors.NewError(errors.ErrInvalidCommand,
					fmt.Sprintf("command with name '%s' already exists. Use --force to overwrite", commandId), nil)
			}
		}
	} else {
		commandId = sc.generateCommandID()
	}

	// Create command
	command := &models.Command{
		Entity: models.Entity{
			ID:           commandId,
			CreatedAt:    time.Now(),
			TerminatedAt: time.Now(),
		},
		Name:      commandName,
		Arguments: commandArgs,
		Status:    true, // Stored commands are marked as successful by default
		Tags:      append(sc.tags, "stored"),
		Category:  sc.category,
	}

	// Set description if provided
	if sc.description != "" {
		command.Output = sc.description // Using Output field to store description
	}

	// Store the command
	if err := sc.repository.Put(context.Background(), *command); err != nil {
		sc.logger.Error("Failed to store command",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite, "failed to store command", err)
	}

	// Display confirmation
	fmt.Printf("Command stored successfully:\n")
	fmt.Printf("ID: %s\n", commandId)
	fmt.Printf("Command: %s %s\n", commandName, strings.Join(commandArgs, " "))
	if sc.description != "" {
		fmt.Printf("Description: %s\n", sc.description)
	}
	if len(sc.tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(sc.tags, ", "))
	}
	if sc.category != "" {
		fmt.Printf("Category: %s\n", sc.category)
	}

	sc.logger.Info("Command stored successfully",
		zap.String("commandId", commandId),
		zap.String("name", commandName),
		zap.Strings("args", commandArgs),
		zap.Strings("tags", sc.tags),
		zap.String("category", sc.category))

	return nil
}

func (sc *StoreCommand) commandExists(id string) bool {
	_, err := sc.repository.Get(id)
	return err == nil
}

func (sc *StoreCommand) generateCommandID() string {
	return fmt.Sprintf("STORED-%d", time.Now().UnixNano())
}

func (sc *StoreCommand) Command() *cobra.Command {
	return sc.cmd
}
