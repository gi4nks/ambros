package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/plugins"
	"github.com/gi4nks/ambros/v3/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// InteractiveCommand represents the interactive command
type InteractiveCommand struct {
	*BaseCommand
	executor *Executor
}

// NewInteractiveCommand creates a new interactive command
func NewInteractiveCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *InteractiveCommand {
	ic := &InteractiveCommand{
		executor: NewExecutor(logger),
	}

	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Interactive command management",
		Long: `Interactive command management with TUI menus.
Provides user-friendly interface for browsing history and management.

Modes:
  dashboard (default)  Interactive dashboard to browse and manage commands
  cleanup              Interactive cleanup of old/failed/duplicate commands

Examples:
  ambros interactive
  ambros interactive cleanup`,
		Args: cobra.MaximumNArgs(1),
		RunE: ic.runE,
	}

	ic.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	ic.cmd = cmd
	return ic
}

func (ic *InteractiveCommand) runE(cmd *cobra.Command, args []string) error {
	// Require a TTY for interactive mode
	if !ic.executor.IsTerminal() {
		return errors.NewError(errors.ErrInvalidCommand, "interactive mode requires a TTY", nil)
	}

	if len(args) == 0 {
		return ic.launchDashboard()
	}

	mode := args[0]
	switch mode {
	case "dashboard":
		return ic.launchDashboard()
	case "cleanup":
		return ic.interactiveCleanup()
	default:
		return errors.NewError(errors.ErrInvalidCommand,
			fmt.Sprintf("unknown mode: %s. Available modes: dashboard, cleanup", mode), nil)
	}
}

func (ic *InteractiveCommand) launchDashboard() error {
	m := tui.NewModel(ic.repository)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return errors.NewError(errors.ErrInternalServer, "failed to run TUI", err)
	}

	selected := m.GetChoice()
	if selected != nil {
		color.Cyan("\n🚀 Rerunning selected command: %s %s", selected.Command, strings.Join(selected.Arguments, " "))
		
		// Wait a bit so the user can see what's happening
		time.Sleep(500 * time.Millisecond)

		// Create a temporary rerun command to use its execution logic
		// or just use the executor directly
		output, errorMsg, success, err := ic.executeCommand(selected.Command, selected.Arguments)
		if err != nil {
			return errors.NewError(errors.ErrExecutionFailed, "failed to execute command", err)
		}

		if output != "" {
			fmt.Print(output)
		}
		if errorMsg != "" {
			fmt.Fprint(os.Stderr, errorMsg)
		}

		if !success {
			os.Exit(1)
		}
	}

	return nil
}

func (ic *InteractiveCommand) executeCommand(name string, args []string) (string, string, bool, error) {
	if _, err := ic.executor.ResolveCommandPath(name); err != nil {
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

func (ic *InteractiveCommand) interactiveCleanup() error {
	color.Cyan("🧹 Interactive Cleanup")

	// Get all commands for analysis
	commands, err := ic.repository.GetAllCommands()
	if err != nil {
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to get commands", err)
	}

	if len(commands) == 0 {
		color.Yellow("📭 No commands to clean up")
		return nil
	}

	// Analyze commands
	failedCount := 0
	oldCount := 0
	dupCount := 0

	// Duplicate detection map
	seen := make(map[string]string)
	dupIDs := make(map[string]struct{})

	cutoff := time.Now().Add(-30 * 24 * time.Hour)

	for _, cmd := range commands {
		if !cmd.Status {
			failedCount++
		}
		if cmd.CreatedAt.Before(cutoff) {
			oldCount++
		}
		key := strings.TrimSpace(cmd.Command)
		if prev, ok := seen[key]; ok {
			// mark duplicate ID for deletion
			dupIDs[cmd.ID] = struct{}{}
			// ensure prev is not double counted
			dupCount++
			_ = prev
		} else {
			seen[key] = cmd.ID
		}
	}

	color.White("📊 Cleanup Analysis:")
	fmt.Printf("Total commands: %s\n", color.YellowString("%d", len(commands)))
	fmt.Printf("Failed commands: %s\n", color.RedString("%d", failedCount))
	fmt.Printf("Commands older than 30 days: %s\n", color.CyanString("%d", oldCount))
	fmt.Printf("Potential duplicates: %s\n", color.MagentaString("%d", dupCount))

	fmt.Println("\nCleanup options:")
	fmt.Println("1. 🗑️  Remove failed commands")
	fmt.Println("2. 📅 Remove commands older than 30 days")
	fmt.Println("3. 🔄 Remove duplicate commands")
	fmt.Println("4. 🧹 Full cleanup (all above)")
	fmt.Println("5. ❌ Cancel")

	fmt.Print("\nSelect cleanup option (1-5): ")
	choice, err := ic.readUserInput()
	if err != nil {
		return err
	}

	var toDeleteIDs = make(map[string]string) // id -> reason

	switch choice {
	case "1": // failed
		for _, cmd := range commands {
			if !cmd.Status {
				toDeleteIDs[cmd.ID] = "failed"
			}
		}
	case "2": // older than 30 days
		for _, cmd := range commands {
			if cmd.CreatedAt.Before(cutoff) {
				toDeleteIDs[cmd.ID] = "older_than_30d"
			}
		}
	case "3": // duplicates
		for id := range dupIDs {
			toDeleteIDs[id] = "duplicate"
		}
	case "4": // full
		for _, cmd := range commands {
			if !cmd.Status || cmd.CreatedAt.Before(cutoff) {
				toDeleteIDs[cmd.ID] = "cleanup_full"
			}
		}
		for id := range dupIDs {
			toDeleteIDs[id] = "duplicate"
		}
	case "5":
		color.Yellow("📫 Cleanup cancelled")
		return nil
	default:
		color.Red("❌ Invalid selection")
		return nil
	}

	if len(toDeleteIDs) == 0 {
		color.Yellow("No commands matched the selected cleanup criteria.")
		return nil
	}

	// Confirm dry-run
	fmt.Print("Dry run (show what would be deleted)? (y/N): ")
	dryRunResp, err := ic.readUserInput()
	if err != nil {
		return err
	}
	dryRun := strings.ToLower(dryRunResp) == "y" || strings.ToLower(dryRunResp) == "yes"

	// List items (limit display to first 20)
	fmt.Printf("\nCommands to be deleted: %d\n", len(toDeleteIDs))
	count := 0
	for id, reason := range toDeleteIDs {
		if count < 20 {
			fmt.Printf(" - %s (reason: %s)\n", id, reason)
		}
		count++
	}
	if count > 20 {
		fmt.Printf(" ... and %d more\n", count-20)
	}

	if dryRun {
		color.Green("✅ Dry run complete — no changes made")
		return nil
	}

	// Final confirmation
	fmt.Print("Proceed with deletion? This cannot be undone. (y/N): ")
	proceed, err := ic.readUserInput()
	if err != nil {
		return err
	}
	if strings.ToLower(proceed) != "y" && strings.ToLower(proceed) != "yes" {
		color.Yellow("📫 Cleanup cancelled")
		return nil
	}

	// Perform deletions
	deleted := 0
	failedDel := 0
	for id := range toDeleteIDs {
		if err := ic.repository.Delete(id); err != nil {
			ic.logger.Error("failed to delete command", zap.String("id", id), zap.Error(err))
			failedDel++
			continue
		}
		deleted++
	}

	color.Green("✅ Deleted %d commands, %d failures", deleted, failedDel)
	return nil
}

func (ic *InteractiveCommand) readUserInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// cleanupDays allows customization of the age threshold for cleanup
var cleanupDays = 30

// SetCleanupDays sets the number of days threshold for old command cleanup (for testing)
func SetCleanupDays(days int) {
	cleanupDays = days
}

func (ic *InteractiveCommand) Command() *cobra.Command {
	return ic.cmd
}
