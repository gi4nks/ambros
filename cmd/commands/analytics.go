package commands

import (
	"fmt"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

type AnalyticsCommand struct {
	*BaseCommand
	opts AnalyticsOptions
}

type AnalyticsOptions struct {
	period string
	format string
	detail bool
}

func NewAnalyticsCommand(logger *zap.Logger, repo RepositoryInterface) *AnalyticsCommand {
	ac := &AnalyticsCommand{}

	cmd := &cobra.Command{
		Use:   "analytics [action]",
		Short: "View analytics and insights about command usage",
		Long: `Get insights into your command usage patterns, performance, and statistics.

Actions:
  summary      Show general usage summary (default)
  most-used    Show most frequently used commands
  slowest      Show slowest commands by execution time
  failures     Show commands that fail most often

Examples:
  ambros analytics
  ambros analytics most-used
  ambros analytics slowest
  ambros analytics failures`,
		RunE: ac.runE,
	}

	ac.BaseCommand = NewBaseCommand(cmd, logger, repo)
	ac.cmd = cmd
	ac.setupFlags(cmd)
	return ac
}

func (ac *AnalyticsCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&ac.opts.period, "period", "p", "7d",
		"Analysis period (24h, 7d, 30d)")
	cmd.Flags().StringVarP(&ac.opts.format, "format", "f", "text",
		"Output format (text/json/yaml)")
	cmd.Flags().BoolVarP(&ac.opts.detail, "detail", "d", false,
		"Show detailed analytics")
}

func (ac *AnalyticsCommand) runE(cmd *cobra.Command, args []string) error {
	action := "summary"
	if len(args) > 0 {
		action = args[0]
	}

	ac.logger.Debug("Analytics command invoked",
		zap.String("action", action))

	switch action {
	case "summary":
		return ac.showSummary()
	case "most-used":
		return ac.showMostUsed()
	case "slowest":
		return ac.showSlowest()
	case "failures":
		return ac.showFailures()
	default:
		return errors.NewError(errors.ErrInvalidCommand,
			fmt.Sprintf("unknown action: %s", action), nil)
	}
}

func (ac *AnalyticsCommand) showSummary() error {
	commands, err := ac.repository.GetAllCommands()
	if err != nil {
		ac.logger.Error("Failed to retrieve commands", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to retrieve commands", err)
	}

	if len(commands) == 0 {
		fmt.Println("No commands found.")
		return nil
	}

	// Calculate basic statistics
	totalCommands := len(commands)
	successCount := 0
	totalDuration := time.Duration(0)
	templateCount := 0

	for _, cmd := range commands {
		if cmd.Status {
			successCount++
		}

		duration := cmd.TerminatedAt.Sub(cmd.CreatedAt)
		totalDuration += duration

		// Check if it's a template
		for _, tag := range cmd.Tags {
			if tag == "template" {
				templateCount++
				break
			}
		}
	}

	successRate := float64(successCount) / float64(totalCommands) * 100
	avgDuration := totalDuration / time.Duration(totalCommands)

	// Display summary
	color.Cyan("📊 Command Analytics Summary\n")

	fmt.Printf("Total Commands: ")
	color.Yellow("%d", totalCommands)
	fmt.Printf("\nSuccess Rate: ")
	if successRate >= 80 {
		color.Green("%.1f%%", successRate)
	} else if successRate >= 60 {
		color.Yellow("%.1f%%", successRate)
	} else {
		color.Red("%.1f%%", successRate)
	}
	fmt.Printf("\nAverage Duration: ")
	color.Yellow("%v", avgDuration.Round(time.Millisecond))
	fmt.Printf("\nTemplate Runs: ")
	color.Cyan("%d", templateCount)
	fmt.Println()

	return nil
}

func (ac *AnalyticsCommand) showMostUsed() error {
	commands, err := ac.repository.GetAllCommands()
	if err != nil {
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to retrieve commands", err)
	}

	// Count command frequency
	cmdCount := make(map[string]int)
	for _, cmd := range commands {
		cmdCount[cmd.Name]++
	}

	// Sort by frequency
	type cmdFreq struct {
		name  string
		count int
	}

	var sorted []cmdFreq
	for name, count := range cmdCount {
		sorted = append(sorted, cmdFreq{name, count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	color.Cyan("🔥 Most Used Commands:\n")

	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}

	for i := 0; i < limit; i++ {
		item := sorted[i]
		fmt.Printf("%d. ", i+1)
		color.Green("%s", item.name)
		fmt.Printf(" - ")
		color.Yellow("%d times", item.count)
		fmt.Printf("\n")
	}

	return nil
}

func (ac *AnalyticsCommand) showSlowest() error {
	commands, err := ac.repository.GetAllCommands()
	if err != nil {
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to retrieve commands", err)
	}

	// Calculate durations and sort
	type cmdDuration struct {
		cmd      models.Command
		duration time.Duration
	}

	var withDuration []cmdDuration
	for _, cmd := range commands {
		duration := cmd.TerminatedAt.Sub(cmd.CreatedAt)
		if duration > time.Millisecond*10 { // Only meaningful durations
			withDuration = append(withDuration, cmdDuration{cmd, duration})
		}
	}

	sort.Slice(withDuration, func(i, j int) bool {
		return withDuration[i].duration > withDuration[j].duration
	})

	color.Cyan("🐌 Slowest Commands:\n")

	limit := 10
	if len(withDuration) < limit {
		limit = len(withDuration)
	}

	for i := 0; i < limit; i++ {
		item := withDuration[i]
		fmt.Printf("%d. ", i+1)
		color.Green("%s", item.cmd.Name)
		fmt.Printf(" - ")
		color.Yellow("%v", item.duration.Round(time.Millisecond))
		fmt.Printf("\n")
	}

	return nil
}

func (ac *AnalyticsCommand) showFailures() error {
	commands, err := ac.repository.GetAllCommands()
	if err != nil {
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to retrieve commands", err)
	}

	// Count failures
	failCount := make(map[string]int)
	for _, cmd := range commands {
		if !cmd.Status {
			failCount[cmd.Name]++
		}
	}

	if len(failCount) == 0 {
		color.Green("🎉 No command failures found!")
		return nil
	}

	// Sort by failure count
	type failureStats struct {
		name     string
		failures int
	}

	var failures []failureStats
	for name, count := range failCount {
		failures = append(failures, failureStats{name, count})
	}

	sort.Slice(failures, func(i, j int) bool {
		return failures[i].failures > failures[j].failures
	})

	color.Cyan("💥 Commands with Failures:\n")

	for i, item := range failures {
		if i >= 10 { // Limit to top 10
			break
		}

		fmt.Printf("%d. ", i+1)
		color.Red("%s", item.name)
		fmt.Printf(" - ")
		color.Yellow("%d failures", item.failures)
		fmt.Printf("\n")
	}

	return nil
}

func (ac *AnalyticsCommand) Command() *cobra.Command {
	return ac.cmd
}
