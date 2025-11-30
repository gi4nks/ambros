package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
)

// SchedulerCommand represents the scheduler command
type SchedulerCommand struct {
	*BaseCommand
	enabled  bool
	interval string
	cronExpr string
}

// NewSchedulerCommand creates a new scheduler command
func NewSchedulerCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *SchedulerCommand {
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

	sc.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
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
	// Get all commands and filter for scheduled ones
	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		sc.logger.Error("Failed to get commands", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead, "failed to get commands", err)
	}

	var scheduledCommands []models.Command
	for _, cmd := range commands {
		if cmd.Schedule != nil {
			scheduledCommands = append(scheduledCommands, cmd)
		}
	}

	if len(scheduledCommands) == 0 {
		color.Yellow("📅 No scheduled commands found")
		color.Cyan("\nCreate a scheduled command:")
		color.White("  ambros scheduler add <command-id> \"0 9 * * 1-5\"")
		return nil
	}

	color.Cyan("📅 Scheduled Commands (%d):", len(scheduledCommands))

	for i, cmd := range scheduledCommands {
		status := "🟢 Enabled"
		if !cmd.Schedule.Enabled {
			status = "🔴 Disabled"
		}

		fmt.Printf("\n%d. %s\n", i+1, color.WhiteString(cmd.Command))
		fmt.Printf("   ID: %s\n", color.CyanString(cmd.ID))
		fmt.Printf("   Cron: %s\n", color.YellowString(cmd.Schedule.CronExpr))
		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Next Run: %s\n", cmd.Schedule.NextRun.Format("2006-01-02 15:04:05"))

		if !cmd.Schedule.LastRun.IsZero() {
			fmt.Printf("   Last Run: %s\n", cmd.Schedule.LastRun.Format("2006-01-02 15:04:05"))
		}
	}

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

	color.Green("✅ Command scheduled successfully:")
	fmt.Printf("ID: %s\n", color.CyanString(commandId))
	fmt.Printf("Command: %s\n", color.WhiteString(command.Command))
	fmt.Printf("Cron: %s\n", color.YellowString(cronExpr))
	fmt.Printf("Enabled: %s\n", color.GreenString("%t", sc.enabled))
	fmt.Printf("Next Run: %s\n", color.CyanString(schedule.NextRun.Format("2006-01-02 15:04:05")))

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
	// Use proper cron parser for validation
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronExpr)
	if err != nil {
		return errors.NewError(errors.ErrScheduleInvalid, "invalid cron expression", err)
	}
	return nil
}

func (sc *SchedulerCommand) calculateNextRun(cronExpr string) time.Time {
	// Use proper cron parser to calculate next run
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		// Fallback to simple calculation
		return time.Now().Add(time.Hour)
	}
	return schedule.Next(time.Now())
}

func (sc *SchedulerCommand) Command() *cobra.Command {
	return sc.cmd
}
