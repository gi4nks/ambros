package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/errors"
	"github.com/gi4nks/ambros/internal/models"
)

// SchedulerCommand represents the scheduler command
type SchedulerCommand struct {
	*BaseCommand
	enabled  bool
	interval string
	cronExpr string
}

// NewSchedulerCommand creates a new scheduler command
func NewSchedulerCommand(logger *zap.Logger, repo RepositoryInterface) *SchedulerCommand {
	sc := &SchedulerCommand{}

	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Manage scheduled command executions",
		Long: `Manage scheduled command executions with cron expressions or intervals.
Supports creating, listing, enabling, disabling, and removing scheduled commands.

Subcommands:
  list                      List all scheduled commands
  add <command-id> <cron>   Add a command to the schedule
  remove <command-id>       Remove a command from the schedule
  enable <command-id>       Enable a scheduled command
  disable <command-id>      Disable a scheduled command
  status                    Show scheduler status

Examples:
  ambros scheduler list
  ambros scheduler add CMD-123 "0 9 * * 1-5"  # Weekdays at 9 AM
  ambros scheduler add CMD-456 --interval 1h   # Every hour
  ambros scheduler enable CMD-123
  ambros scheduler disable CMD-456
  ambros scheduler remove CMD-123`,
		Args: cobra.MinimumNArgs(1),
		RunE: sc.runE,
	}

	sc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	sc.cmd = cmd
	sc.setupFlags(cmd)
	return sc
}

func (sc *SchedulerCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&sc.enabled, "enabled", true, "Enable the scheduled command")
	cmd.Flags().StringVar(&sc.interval, "interval", "", "Interval for execution (e.g., 1h, 30m)")
	cmd.Flags().StringVar(&sc.cronExpr, "cron", "", "Cron expression for scheduling")
}

func (sc *SchedulerCommand) runE(cmd *cobra.Command, args []string) error {
	sc.logger.Debug("Scheduler command invoked",
		zap.String("subcommand", args[0]),
		zap.Strings("args", args))

	switch args[0] {
	case "list":
		return sc.listScheduled()
	case "add":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "command ID required for add", nil)
		}
		return sc.addScheduled(args[1], args[2:])
	case "remove":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "command ID required for remove", nil)
		}
		return sc.removeScheduled(args[1])
	case "enable":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "command ID required for enable", nil)
		}
		return sc.enableScheduled(args[1])
	case "disable":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "command ID required for disable", nil)
		}
		return sc.disableScheduled(args[1])
	case "status":
		return sc.showStatus()
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unknown subcommand: "+args[0], nil)
	}
}

func (sc *SchedulerCommand) listScheduled() error {
	sc.logger.Debug("Listing scheduled commands")

	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		sc.logger.Error("Failed to retrieve commands", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead, "failed to retrieve commands", err)
	}

	scheduledCommands := make([]models.Command, 0)
	for _, cmd := range commands {
		if cmd.Schedule != nil {
			scheduledCommands = append(scheduledCommands, cmd)
		}
	}

	if len(scheduledCommands) == 0 {
		fmt.Println("No scheduled commands found")
		return nil
	}

	fmt.Printf("Scheduled commands (%d):\n\n", len(scheduledCommands))

	for i, cmd := range scheduledCommands {
		fmt.Printf("%d. %s %s\n", i+1, cmd.Name, strings.Join(cmd.Arguments, " "))
		fmt.Printf("   ID: %s\n", cmd.ID)
		fmt.Printf("   Cron: %s\n", cmd.Schedule.CronExpr)
		fmt.Printf("   Enabled: %t\n", cmd.Schedule.Enabled)
		if !cmd.Schedule.LastRun.IsZero() {
			fmt.Printf("   Last Run: %s\n", cmd.Schedule.LastRun.Format("2006-01-02 15:04:05"))
		}
		if !cmd.Schedule.NextRun.IsZero() {
			fmt.Printf("   Next Run: %s\n", cmd.Schedule.NextRun.Format("2006-01-02 15:04:05"))
		}

		if i < len(scheduledCommands)-1 {
			fmt.Println()
		}
	}

	sc.logger.Info("Listed scheduled commands",
		zap.Int("count", len(scheduledCommands)))

	return nil
}

func (sc *SchedulerCommand) addScheduled(commandId string, scheduleArgs []string) error {
	sc.logger.Debug("Adding scheduled command",
		zap.String("commandId", commandId),
		zap.Strings("scheduleArgs", scheduleArgs))

	// Get the command to schedule
	command, err := sc.repository.Get(commandId)
	if err != nil {
		sc.logger.Error("Command not found",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("command not found: %s", commandId), err)
	}

	// Determine schedule expression
	var cronExpr string
	if sc.cronExpr != "" {
		cronExpr = sc.cronExpr
	} else if sc.interval != "" {
		cronExpr = sc.intervalToCron(sc.interval)
	} else if len(scheduleArgs) > 0 {
		cronExpr = strings.Join(scheduleArgs, " ")
	} else {
		return errors.NewError(errors.ErrInvalidCommand, "schedule expression required", nil)
	}

	// Validate cron expression (basic validation)
	if err := sc.validateCronExpr(cronExpr); err != nil {
		return errors.NewError(errors.ErrInvalidCommand, "invalid cron expression", err)
	}

	// Create schedule
	schedule := &models.Schedule{
		CronExpr: cronExpr,
		Enabled:  sc.enabled,
		NextRun:  sc.calculateNextRun(cronExpr),
	}

	// Update command with schedule
	command.Schedule = schedule

	if err := sc.repository.Put(context.Background(), *command); err != nil {
		sc.logger.Error("Failed to update command with schedule",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite, "failed to schedule command", err)
	}

	fmt.Printf("Command scheduled successfully:\n")
	fmt.Printf("ID: %s\n", commandId)
	fmt.Printf("Cron: %s\n", cronExpr)
	fmt.Printf("Enabled: %t\n", sc.enabled)
	fmt.Printf("Next Run: %s\n", schedule.NextRun.Format("2006-01-02 15:04:05"))

	sc.logger.Info("Command scheduled successfully",
		zap.String("commandId", commandId),
		zap.String("cronExpr", cronExpr),
		zap.Bool("enabled", sc.enabled))

	return nil
}

func (sc *SchedulerCommand) removeScheduled(commandId string) error {
	sc.logger.Debug("Removing scheduled command",
		zap.String("commandId", commandId))

	command, err := sc.repository.Get(commandId)
	if err != nil {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("command not found: %s", commandId), err)
	}

	if command.Schedule == nil {
		return errors.NewError(errors.ErrInvalidCommand, "command is not scheduled", nil)
	}

	// Remove schedule
	command.Schedule = nil

	if err := sc.repository.Put(context.Background(), *command); err != nil {
		sc.logger.Error("Failed to remove schedule",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite, "failed to remove schedule", err)
	}

	fmt.Printf("Schedule removed for command: %s\n", commandId)

	sc.logger.Info("Schedule removed",
		zap.String("commandId", commandId))

	return nil
}

func (sc *SchedulerCommand) enableScheduled(commandId string) error {
	return sc.toggleScheduled(commandId, true)
}

func (sc *SchedulerCommand) disableScheduled(commandId string) error {
	return sc.toggleScheduled(commandId, false)
}

func (sc *SchedulerCommand) toggleScheduled(commandId string, enabled bool) error {
	action := "Enabling"
	if !enabled {
		action = "Disabling"
	}

	sc.logger.Debug(action+" scheduled command",
		zap.String("commandId", commandId))

	command, err := sc.repository.Get(commandId)
	if err != nil {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("command not found: %s", commandId), err)
	}

	if command.Schedule == nil {
		return errors.NewError(errors.ErrInvalidCommand, "command is not scheduled", nil)
	}

	command.Schedule.Enabled = enabled

	if err := sc.repository.Put(context.Background(), *command); err != nil {
		sc.logger.Error("Failed to update schedule",
			zap.String("commandId", commandId),
			zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite, "failed to update schedule", err)
	}

	status := "enabled"
	if !enabled {
		status = "disabled"
	}

	fmt.Printf("Schedule %s for command: %s\n", status, commandId)

	sc.logger.Info("Schedule updated",
		zap.String("commandId", commandId),
		zap.Bool("enabled", enabled))

	return nil
}

func (sc *SchedulerCommand) showStatus() error {
	sc.logger.Debug("Showing scheduler status")

	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		return errors.NewError(errors.ErrRepositoryRead, "failed to retrieve commands", err)
	}

	var total, enabled, disabled int
	for _, cmd := range commands {
		if cmd.Schedule != nil {
			total++
			if cmd.Schedule.Enabled {
				enabled++
			} else {
				disabled++
			}
		}
	}

	fmt.Printf("Scheduler Status:\n")
	fmt.Printf("Total scheduled commands: %d\n", total)
	fmt.Printf("Enabled: %d\n", enabled)
	fmt.Printf("Disabled: %d\n", disabled)

	sc.logger.Info("Scheduler status displayed",
		zap.Int("total", total),
		zap.Int("enabled", enabled),
		zap.Int("disabled", disabled))

	return nil
}

// Helper functions
func (sc *SchedulerCommand) intervalToCron(interval string) string {
	// Simple interval to cron conversion
	// This is a basic implementation - could be enhanced
	switch interval {
	case "1m":
		return "* * * * *"
	case "5m":
		return "*/5 * * * *"
	case "10m":
		return "*/10 * * * *"
	case "30m":
		return "*/30 * * * *"
	case "1h":
		return "0 * * * *"
	case "2h":
		return "0 */2 * * *"
	case "6h":
		return "0 */6 * * *"
	case "12h":
		return "0 */12 * * *"
	case "1d":
		return "0 0 * * *"
	default:
		return interval // Assume it's already a cron expression
	}
}

func (sc *SchedulerCommand) validateCronExpr(cronExpr string) error {
	// Basic validation - could be enhanced with proper cron parsing
	parts := strings.Fields(cronExpr)
	if len(parts) != 5 {
		return fmt.Errorf("cron expression must have 5 fields")
	}
	return nil
}

func (sc *SchedulerCommand) calculateNextRun(cronExpr string) time.Time {
	// Simple calculation - in a real implementation, use a proper cron parser
	return time.Now().Add(time.Hour)
}

func (sc *SchedulerCommand) Command() *cobra.Command {
	return sc.cmd
}
