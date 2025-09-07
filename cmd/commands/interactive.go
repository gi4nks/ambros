package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/errors"
)

// InteractiveCommand represents the interactive command
type InteractiveCommand struct {
	*BaseCommand
	mode string
}

// NewInteractiveCommand creates a new interactive command
func NewInteractiveCommand(logger *zap.Logger, repo RepositoryInterface) *InteractiveCommand {
	ic := &InteractiveCommand{}

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

	ic.BaseCommand = NewBaseCommand(cmd, logger, repo)
	ic.cmd = cmd
	return ic
}

func (ic *InteractiveCommand) runE(cmd *cobra.Command, args []string) error {
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

	// Date filter
	fmt.Print("📅 From date (YYYY-MM-DD, press Enter to skip): ")
	fromDate, _ := ic.readUserInput()
	if fromDate != "" {
		filters = append(filters, fmt.Sprintf("from:%s", fromDate))
	}

	fmt.Print("📅 To date (YYYY-MM-DD, press Enter to skip): ")
	toDate, _ := ic.readUserInput()
	if toDate != "" {
		filters = append(filters, fmt.Sprintf("to:%s", toDate))
	}

	// Execute search
	color.Yellow("\n🔍 Executing search with filters: %s", strings.Join(filters, ", "))

	// TODO: Integrate with actual search command
	color.Green("✅ Search completed! (Integration with search command coming soon)")

	return nil
}

func (ic *InteractiveCommand) interactiveSelect() error {
	color.Cyan("📋 Select & Execute Commands")

	// Get recent commands
	commands, err := ic.repository.GetLimitCommands(20)
	if err != nil {
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to get commands", err)
	}

	if len(commands) == 0 {
		color.Yellow("📭 No commands found")
		return nil
	}

	// Display commands
	color.White("Recent commands:")
	fmt.Println()

	for i, cmd := range commands {
		status := "✅"
		if !cmd.Status {
			status = "❌"
		}

		fmt.Printf("%d. %s %s %s\n",
			i+1,
			status,
			color.WhiteString(cmd.Command),
			color.CyanString("(%s)", cmd.CreatedAt.Format("15:04:05")))
	}

	fmt.Printf("\nEnter command number to execute (1-%d), or 0 to cancel: ", len(commands))
	choice, err := ic.readUserInput()
	if err != nil {
		return err
	}

	if choice == "0" {
		color.Yellow("📫 Operation cancelled")
		return nil
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(commands) {
		color.Red("❌ Invalid selection")
		return nil
	}

	selectedCmd := commands[index-1]
	color.Green("🚀 Would execute: %s", selectedCmd.Command)
	color.Yellow("⚠️  Command execution integration coming soon!")

	return nil
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
	failed := 0
	old := 0
	duplicates := 0

	for _, cmd := range commands {
		if !cmd.Status {
			failed++
		}
		// Simple duplicate detection by command text
		// TODO: Implement proper duplicate detection
	}

	color.White("📊 Cleanup Analysis:")
	fmt.Printf("Total commands: %s\n", color.YellowString("%d", len(commands)))
	fmt.Printf("Failed commands: %s\n", color.RedString("%d", failed))
	fmt.Printf("Commands older than 30 days: %s\n", color.CyanString("%d", old))
	fmt.Printf("Potential duplicates: %s\n", color.MagentaString("%d", duplicates))

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

	switch choice {
	case "1", "2", "3", "4":
		color.Yellow("⚠️  Cleanup functionality implementation coming soon!")
		color.Green("✅ Selected: Option %s", choice)
	case "5":
		color.Yellow("📫 Cleanup cancelled")
	default:
		color.Red("❌ Invalid selection")
	}

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

	color.Yellow("\n⚠️  Template management integration coming soon!")
	return nil
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
	envMap := make(map[string]int)
	for _, env := range environments {
		if env.Category == "environment" {
			envName := ic.extractEnvName(env.Name)
			envMap[envName]++
		}
	}

	i := 1
	for name, count := range envMap {
		fmt.Printf("%d. %s (%d variables)\n", i, color.GreenString(name), count-1)
		i++
	}

	color.Yellow("\n⚠️  Environment management integration coming soon!")
	return nil
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
