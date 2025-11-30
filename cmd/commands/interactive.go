package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// splitShellArgs splits a shell command string into arguments, honoring simple quotes.
// It is a lightweight helper and does not aim to be a full shell parser.
func splitShellArgs(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	esc := false

	for _, r := range s {
		if esc {
			cur.WriteRune(r)
			esc = false
			continue
		}
		switch r {
		case '\\':
			esc = true
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
			cur.WriteRune(r)
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
			cur.WriteRune(r)
		case ' ', '\t', '\n':
			if inSingle || inDouble {
				cur.WriteRune(r)
			} else {
				if cur.Len() > 0 {
					args = append(args, cur.String())
					cur.Reset()
				}
			}
		default:
			cur.WriteRune(r)
		}
	}
	if esc || inSingle || inDouble {
		return nil, errors.NewError(errors.ErrInvalidCommand, "unterminated quote or escape in command", nil)
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args, nil
}

// InteractiveCommand represents the interactive command
type InteractiveCommand struct {
	*BaseCommand
	mode     string
	executor *Executor // New field for shared execution logic
}

// NewInteractiveCommand creates a new interactive command
func NewInteractiveCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *InteractiveCommand {
	ic := &InteractiveCommand{
		executor: NewExecutor(logger), // Initialize the executor
	}

	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Interactive command management",
		Long: `Interactive command management with menus and selection.
Provides user-friendly interfaces for command operations.

Modes:
  search    Interactive search with filters
  select    Select and execute from command history
  cleanup   Interactive cleanup of old commands
  manage    Manage templates and environments interactively

Examples:
  ambros interactive search
  ambros interactive select
  ambros interactive cleanup
  ambros interactive manage`,
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
		return ic.showMainMenu()
	}

	ic.mode = args[0]
	switch ic.mode {
	case "search":
		return ic.interactiveSearch()
	case "select":
		return ic.interactiveSelect()
	case "cleanup":
		return ic.interactiveCleanup()
	case "manage":
		return ic.interactiveManage()
	default:
		return errors.NewError(errors.ErrInvalidCommand,
			fmt.Sprintf("unknown mode: %s", ic.mode), nil)
	}
}

func (ic *InteractiveCommand) showMainMenu() error {
	color.Cyan("🎯 Ambros Interactive Mode")
	color.White("Select an option:")
	fmt.Println()

	options := []string{
		"🔍 Interactive Search",
		"📋 Select & Execute Commands",
		"🧹 Cleanup Old Commands",
		"⚙️  Manage Templates & Environments",
		"❌ Exit",
	}

	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}

	fmt.Print("\nEnter your choice (1-5): ")
	choice, err := ic.readUserInput()
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		return ic.interactiveSearch()
	case "2":
		return ic.interactiveSelect()
	case "3":
		return ic.interactiveCleanup()
	case "4":
		return ic.interactiveManage()
	case "5":
		color.Green("👋 Goodbye!")
		return nil
	default:
		color.Red("❌ Invalid choice. Please try again.")
		return ic.showMainMenu()
	}
}

func (ic *InteractiveCommand) interactiveSearch() error {
	color.Cyan("🔍 Interactive Search")
	color.White("Build your search criteria:")
	fmt.Println()

	var filters []string

	// Text search
	fmt.Print("🔤 Search text (press Enter to skip): ")
	text, _ := ic.readUserInput()
	if text != "" {
		filters = append(filters, fmt.Sprintf("text:%s", text))
	}

	// Status filter
	fmt.Print("✅ Filter by status (success/failed/all) [all]: ")
	status, _ := ic.readUserInput()
	if status == "" {
		status = "all"
	}
	if status != "all" {
		filters = append(filters, fmt.Sprintf("status:%s", status))
	}

	// Tag filter
	fmt.Print("🏷️  Filter by tag (press Enter to skip): ")
	tag, _ := ic.readUserInput()
	if tag != "" {
		filters = append(filters, fmt.Sprintf("tag:%s", tag))
	}

	// Date filter (From only maps to --since)
	fmt.Print("📅 From date (YYYY-MM-DD, press Enter to skip): ")
	fromDateStr, _ := ic.readUserInput()
	var fromDate time.Time
	if fromDateStr != "" {
		t, err := time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			color.Red("Invalid date format, expected YYYY-MM-DD")
			return nil
		}
		fromDate = t
		filters = append(filters, fmt.Sprintf("from:%s", fromDateStr))
	}

	fmt.Print("📅 To date (YYYY-MM-DD, press Enter to skip): ")
	toDateStr, _ := ic.readUserInput()
	if toDateStr != "" {
		if _, err := time.Parse("2006-01-02", toDateStr); err != nil {
			color.Red("Invalid date format, expected YYYY-MM-DD")
			return nil
		}
		filters = append(filters, fmt.Sprintf("to:%s", toDateStr))
	}

	// Execute search by invoking the search command
	color.Yellow("\n🔍 Executing search with filters: %s", strings.Join(filters, ", "))

	// Build args for search command
	searchArgs := []string{}
	if text != "" {
		searchArgs = append(searchArgs, text)
	}
	// Use tag/status/from/to as flags where applicable
	if tag != "" {
		searchArgs = append(searchArgs, "--tag", tag)
	}
	if status != "" && status != "all" {
		// normalize status: accept 'failed' -> 'failure'
		if strings.EqualFold(status, "failed") {
			status = "failure"
		}
		if !strings.EqualFold(status, "success") && !strings.EqualFold(status, "failure") {
			color.Red("Invalid status filter. Use 'success', 'failure', or leave empty for all")
			return nil
		}
		// search command expects status via --status
		searchArgs = append(searchArgs, "--status", status)
	}
	// Note: date filters in interactive UI map to --since (simple heuristic)
	if !fromDate.IsZero() {
		// convert fromDate to duration like '720h'
		dur := time.Since(fromDate)
		hours := int(dur.Hours())
		if hours < 1 {
			hours = 1
		}
		searchArgs = append(searchArgs, "--since", fmt.Sprintf("%dh", hours))
	}

	sc := NewSearchCommand(ic.logger, ic.repository, ic.api)
	// Execute SearchCommand.runE with constructed args
	if err := sc.runE(sc.cmd, searchArgs); err != nil {
		ic.logger.Error("interactive search failed", zap.Error(err))
		return err
	}

	color.Green("✅ Search completed")

	return nil
}

func (ic *InteractiveCommand) interactiveSelect() error {
	color.Cyan("📋 Select & Execute Commands")

	// Get all commands (we'll paginate locally)
	commands, err := ic.repository.GetAllCommands()
	if err != nil {
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to get commands", err)
	}

	if len(commands) == 0 {
		color.Yellow("📭 No commands found")
		return nil
	}

	// Paginate results locally
	pageSize := 20
	total := len(commands)
	if total == 0 {
		color.Yellow("📭 No commands found")
		return nil
	}
	page := 0
	for {
		start := page * pageSize
		if start >= total {
			start = 0
			page = 0
		}
		end := start + pageSize
		if end > total {
			end = total
		}

		color.White("Commands (page %d):", page+1)
		for i := start; i < end; i++ {
			cmd := commands[i]
			status := "✅"
			if !cmd.Status {
				status = "❌"
			}
			fmt.Printf("%d. %s %s %s\n", i+1, status, color.WhiteString(cmd.Command), color.CyanString("(%s)", cmd.CreatedAt.Format("15:04:05")))
		}

		fmt.Printf("\nEnter command number to execute (1-%d), n for next, p for prev, 0 to cancel: ", total)
		choice, err := ic.readUserInput()
		if err != nil {
			return err
		}
		if choice == "0" {
			color.Yellow("📫 Operation cancelled")
			return nil
		}
		if choice == "n" {
			if end == total {
				color.Yellow("Already at last page")
			} else {
				page++
			}
			continue
		}
		if choice == "p" {
			if page == 0 {
				color.Yellow("Already at first page")
			} else {
				page--
			}
			continue
		}

		index, err := strconv.Atoi(choice)
		if err != nil || index < 1 || index > total {
			color.Red("❌ Invalid selection")
			return nil
		}

		selectedCmd := commands[index-1]

		// Confirm execution
		fmt.Printf("\nExecute command: %s ? (y/N): ", selectedCmd.Command)
		confirm, err := ic.readUserInput()
		if err != nil {
			return err
		}
		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			color.Yellow("📫 Operation cancelled")
			return nil
		}

		// Use stored Arguments if available to avoid naive splitting
		parts := selectedCmd.Arguments
		if len(parts) == 0 {
			// use shell-like splitting to respect quotes
			ps, err := splitShellArgs(selectedCmd.Command)
			if err != nil {
				// fallback to simple split
				parts = strings.Fields(selectedCmd.Command)
			} else {
				parts = ps
			}
		}
		if len(parts) == 0 {
			color.Yellow("No executable command found")
			return nil
		}

		// Ask user whether to stream live output or capture and show after
		fmt.Print("Run mode - stream live (s) or capture and show (c) [s]: ")
		modeChoice, err := ic.readUserInput()
		if err != nil {
			return err
		}
		if modeChoice == "" {
			modeChoice = "s"
		}

		if strings.ToLower(modeChoice) == "c" || strings.ToLower(modeChoice) == "capture" {
			// Capture output and show after
			// Show the command being executed
			color.Cyan("\n🔁 Executing: %s", selectedCmd.Command)

			exitCode, out, err := ic.executor.ExecuteCapture(parts[0], parts[1:])
			if err != nil {
				ic.logger.Error("failed to execute selected command (capture)", zap.Error(err))
				color.Red("Execution failed: %v", err)
			} else {
				if out != "" {
					fmt.Println("\n--- Command output ---")
					fmt.Print(out)
					fmt.Println("--- End output ---")
				}
				if exitCode == 0 {
					color.Green("✅ Command exited with code 0")
				} else {
					color.Red("❌ Command exited with code %d", exitCode)
				}
			}
		} else {
			// Stream live to terminal
			// Show the command being executed
			color.Cyan("\n🔁 Executing: %s", selectedCmd.Command)

			exitCode, err := ic.executor.ExecuteCommandAuto(parts[0], parts[1:])
			if err != nil {
				ic.logger.Error("failed to execute selected command (stream)", zap.Error(err))
				color.Red("Execution failed: %v", err)
			} else {
				if exitCode == 0 {
					color.Green("✅ Command exited with code 0")
				} else {
					color.Red("❌ Command exited with code %d", exitCode)
				}
			}
		}

		// Stay on the same listing page (continue the loop)
		continue

	}
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
	dryRun := !(strings.ToLower(dryRunResp) == "y" || strings.ToLower(dryRunResp) == "yes")

	// List items
	fmt.Printf("\nCommands to be deleted: %d\n", len(toDeleteIDs))
	for id, reason := range toDeleteIDs {
		fmt.Printf(" - %s (reason: %s)\n", id, reason)
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

func (ic *InteractiveCommand) interactiveManage() error {
	color.Cyan("⚙️  Interactive Management")

	fmt.Println("Management options:")
	fmt.Println("1. 🎯 Manage Templates")
	fmt.Println("2. 🌍 Manage Environments")
	fmt.Println("3. 📊 View Analytics")
	fmt.Println("4. ⚙️  System Settings")
	fmt.Println("5. ❌ Back to main menu")

	fmt.Print("\nSelect option (1-5): ")
	choice, err := ic.readUserInput()
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		return ic.manageTemplates()
	case "2":
		return ic.manageEnvironments()
	case "3":
		return ic.viewAnalytics()
	case "4":
		return ic.systemSettings()
	case "5":
		return ic.showMainMenu()
	default:
		color.Red("❌ Invalid selection")
		return ic.interactiveManage()
	}
}

func (ic *InteractiveCommand) manageTemplates() error {
	color.Cyan("🎯 Template Management")

	// Get templates
	templates, err := ic.repository.SearchByTag("template")
	if err != nil {
		return err
	}

	if len(templates) == 0 {
		color.Yellow("📭 No templates found")
		color.Cyan("Create your first template:")
		color.White("  ambros template save mytemplate \"echo hello\"")
		return nil
	}

	color.White("Available templates:")
	for i, template := range templates {
		fmt.Printf("%d. %s\n", i+1, color.GreenString(template.Name))
	}

	fmt.Println("\nOptions:")
	fmt.Println("1. Delete a template")
	fmt.Println("2. Back")
	fmt.Print("Select option (1-2): ")
	choice, err := ic.readUserInput()
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		fmt.Print("Enter template number to delete: ")
		idxStr, err := ic.readUserInput()
		if err != nil {
			return err
		}
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 1 || idx > len(templates) {
			color.Red("Invalid selection")
			return nil
		}
		tpl := templates[idx-1]
		fmt.Printf("Dry run (show what would be deleted)? (y/N): ")
		dryRunResp, _ := ic.readUserInput()
		dryRun := !(strings.ToLower(dryRunResp) == "y" || strings.ToLower(dryRunResp) == "yes")

		fmt.Printf("\nTemplate to delete: %s\n", tpl.Name)
		if dryRun {
			color.Green("✅ Dry run - nothing deleted")
			return nil
		}
		fmt.Print("Proceed with deletion? (y/N): ")
		proceed, _ := ic.readUserInput()
		if strings.ToLower(proceed) != "y" && strings.ToLower(proceed) != "yes" {
			color.Yellow("Cancelled")
			return nil
		}

		// Templates are stored as commands with category 'template' — delete matched entries
		// Find template entries by name and delete
		deleted := 0
		for _, t := range templates {
			if t.Name == tpl.Name {
				if err := ic.repository.Delete(t.ID); err != nil {
					ic.logger.Error("failed to delete template", zap.String("id", t.ID), zap.Error(err))
					continue
				}
				deleted++
			}
		}
		color.Green("✅ Deleted %d template entries", deleted)
		return nil
	default:
		return nil
	}
}

func (ic *InteractiveCommand) manageEnvironments() error {
	color.Cyan("🌍 Environment Management")

	// Get environments
	environments, err := ic.repository.SearchByTag("environment")
	if err != nil {
		return err
	}

	if len(environments) == 0 {
		color.Yellow("📭 No environments found")
		color.Cyan("Create your first environment:")
		color.White("  ambros env create development")
		return nil
	}

	color.White("Available environments:")
	envMap := make(map[string][]models.Command)
	for _, env := range environments {
		if env.Category == "environment" {
			envName := ic.extractEnvName(env.Name)
			envMap[envName] = append(envMap[envName], env)
		}
	}

	names := make([]string, 0, len(envMap))
	for name := range envMap {
		names = append(names, name)
	}

	for i, name := range names {
		fmt.Printf("%d. %s (%d variables)\n", i+1, color.GreenString(name), len(envMap[name]))
	}

	fmt.Println("\nOptions:")
	fmt.Println("1. Delete an environment")
	fmt.Println("2. Back")
	fmt.Print("Select option (1-2): ")
	choice, err := ic.readUserInput()
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		fmt.Print("Enter environment number to delete: ")
		idxStr, err := ic.readUserInput()
		if err != nil {
			return err
		}
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 1 || idx > len(names) {
			color.Red("Invalid selection")
			return nil
		}
		name := names[idx-1]
		entries := envMap[name]

		fmt.Printf("Dry run (show what would be deleted)? (y/N): ")
		dryRunResp, _ := ic.readUserInput()
		dryRun := !(strings.ToLower(dryRunResp) == "y" || strings.ToLower(dryRunResp) == "yes")

		fmt.Printf("\nEnvironment to delete: %s (entries: %d)\n", name, len(entries))
		if dryRun {
			color.Green("✅ Dry run - nothing deleted")
			return nil
		}

		fmt.Print("Proceed with deletion? (y/N): ")
		proceed, _ := ic.readUserInput()
		if strings.ToLower(proceed) != "y" && strings.ToLower(proceed) != "yes" {
			color.Yellow("Cancelled")
			return nil
		}

		deleted := 0
		for _, e := range entries {
			if err := ic.repository.Delete(e.ID); err != nil {
				ic.logger.Error("failed to delete environment entry", zap.String("id", e.ID), zap.Error(err))
				continue
			}
			deleted++
		}

		color.Green("✅ Deleted %d environment entries for %s", deleted, name)
		return nil
	default:
		return nil
	}
}

func (ic *InteractiveCommand) viewAnalytics() error {
	color.Cyan("📊 Analytics Dashboard")
	color.Yellow("⚠️  Interactive analytics coming soon!")
	color.Cyan("For now, use: ambros analytics")
	return nil
}

func (ic *InteractiveCommand) systemSettings() error {
	color.Cyan("⚙️  System Settings")
	color.Yellow("⚠️  Interactive settings coming soon!")
	color.Cyan("Current configuration file: ~/.ambros.yaml")
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

func (ic *InteractiveCommand) extractEnvName(cmdName string) string {
	parts := strings.Split(cmdName, ":")
	if len(parts) >= 2 && parts[0] == "env" {
		return parts[1]
	}
	return cmdName
}

func (ic *InteractiveCommand) Command() *cobra.Command {
	return ic.cmd
}
