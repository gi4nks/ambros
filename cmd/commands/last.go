package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
)

type LastCommand struct {
	*BaseCommand
	limit      int
	failedOnly bool
}

func NewLastCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *LastCommand {
	lc := &LastCommand{}

	cmd := &cobra.Command{
		Use:   "last",
		Short: "Show last executed commands",
		Long:  `Display the most recently executed commands with optional filtering.`,
		RunE:  lc.runE,
	}

	lc.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	lc.cmd = cmd
	lc.setupFlags(cmd)
	return lc
}

func (lc *LastCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&lc.limit, "limit", "n", 10, "Number of commands to show")
	cmd.Flags().BoolVarP(&lc.failedOnly, "failed", "f", false, "Show only failed commands")
}

func (lc *LastCommand) runE(cmd *cobra.Command, args []string) error {
	lc.logger.Debug("Last command invoked",
		zap.Int("limit", lc.limit),
		zap.Bool("failedOnly", lc.failedOnly))

	commands, err := lc.repository.GetAllCommands()
	if err != nil {
		lc.logger.Error("Failed to retrieve commands", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead, "failed to retrieve commands", err)
	}

	// Sort commands by creation time (most recent first)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].CreatedAt.After(commands[j].CreatedAt)
	})

	// Apply limit
	if len(commands) > lc.limit {
		commands = commands[:lc.limit]
	}

	// Filter for failed commands if requested
	if lc.failedOnly {
		filtered := make([]models.Command, 0)
		for _, command := range commands {
			if !command.Status {
				filtered = append(filtered, command)
			}
		}
		commands = filtered
	}

	if len(commands) == 0 {
		if lc.failedOnly {
			fmt.Println("No failed commands found")
		} else {
			fmt.Println("No commands found")
		}
		return nil
	}

	// Display commands to user
	fmt.Printf("Last %d command(s):\n\n", len(commands))

	for i, command := range commands {
		fmt.Printf("%d. %s %s\n", i+1, command.Name, strings.Join(command.Arguments, " "))

		// Display ID with color based on status
		if command.Status {
			color.Green("   ID: %s", command.ID)
		} else {
			color.Red("   ID: %s", command.ID)
		}

		fmt.Printf("   Created: %s\n", command.CreatedAt.Format("2006-01-02 15:04:05"))

		// Display status with color
		fmt.Print("   Status: ")
		if command.Status {
			color.Green("Success")
		} else {
			color.Red("Failed")
		}

		if len(command.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(command.Tags, ", "))
		}

		if command.Category != "" {
			fmt.Printf("   Category: %s\n", command.Category)
		}

		if i < len(commands)-1 {
			fmt.Println()
		}
	}

	lc.logger.Info("Last command completed",
		zap.Int("commandsRetrieved", len(commands)),
		zap.Bool("failedOnly", lc.failedOnly))

	return nil
}

func (lc *LastCommand) formatStatus(status bool) string {
	if status {
		return "Success"
	}
	return "Failed"
}

func (lc *LastCommand) Command() *cobra.Command {
	return lc.cmd
}
