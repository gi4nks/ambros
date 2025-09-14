package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

// EnvCommand represents the environment management command
type EnvCommand struct {
	*BaseCommand
	action      string
	envName     string
	global      bool
	interactive bool
}

// NewEnvCommand creates a new environment command
func NewEnvCommand(logger *zap.Logger, repo RepositoryInterface) *EnvCommand {
	ec := &EnvCommand{}

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment variables and contexts",
		Long: `Manage environment variables and execution contexts.
Supports creating named environments, setting variables, and applying them to commands.

Subcommands:
  list                         List all environments
  create <name>                Create a new environment
  delete <name>                Delete an environment
  set <name> <key> <value>     Set a variable in an environment
  unset <name> <key>           Remove a variable from an environment
  show <name>                  Show environment details
  apply <name> <command>       Run command with environment variables

Examples:
  ambros env list
  ambros env create production
  ambros env set production API_URL https://api.prod.com
  ambros env set production DEBUG false
  ambros env show production
  ambros env apply production "curl $API_URL/health"
  ambros env delete production`,
		Args: cobra.MinimumNArgs(1),
		RunE: ec.runE,
	}

	ec.BaseCommand = NewBaseCommand(cmd, logger, repo)
	ec.cmd = cmd
	ec.setupFlags(cmd)
	return ec
}

func (ec *EnvCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&ec.global, "global", "g", false, "Apply to global environment")
	cmd.Flags().BoolVarP(&ec.interactive, "interactive", "i", false, "Interactive environment setup")
}

func (ec *EnvCommand) runE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand,
			"env command requires an action", nil)
	}

	ec.action = args[0]

	switch ec.action {
	case "list":
		return ec.listEnvironments()
	case "create":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand,
				"usage: env create <name>", nil)
		}
		ec.envName = args[1]
		return ec.createEnvironment()
	case "delete":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand,
				"usage: env delete <name>", nil)
		}
		ec.envName = args[1]
		return ec.deleteEnvironment()
	case "set":
		if len(args) < 4 {
			return errors.NewError(errors.ErrInvalidCommand,
				"usage: env set <name> <key> <value>", nil)
		}
		return ec.setVariable(args[1], args[2], args[3])
	case "unset":
		if len(args) < 3 {
			return errors.NewError(errors.ErrInvalidCommand,
				"usage: env unset <name> <key>", nil)
		}
		return ec.unsetVariable(args[1], args[2])
	case "show":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand,
				"usage: env show <name>", nil)
		}
		ec.envName = args[1]
		return ec.showEnvironment()
	case "apply":
		if len(args) < 3 {
			return errors.NewError(errors.ErrInvalidCommand,
				"usage: env apply <name> <command>", nil)
		}
		return ec.applyEnvironment(args[1], strings.Join(args[2:], " "))
	default:
		return errors.NewError(errors.ErrInvalidCommand,
			fmt.Sprintf("unknown action: %s", ec.action), nil)
	}
}

func (ec *EnvCommand) listEnvironments() error {
	// Search for environment commands by tag
	envCommands, err := ec.repository.SearchByTag("environment")
	if err != nil {
		ec.logger.Error("Failed to search environments", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to list environments", err)
	}

	if len(envCommands) == 0 {
		color.Yellow("📁 No environments found")
		color.Cyan("\nCreate your first environment:")
		color.White("  ambros env create development")
		return nil
	}

	// Group environments by name
	envMap := make(map[string][]models.Command)
	for _, cmd := range envCommands {
		if cmd.Category == "environment" {
			envName := extractEnvName(cmd.Name)
			envMap[envName] = append(envMap[envName], cmd)
		}
	}

	color.Cyan("📁 Available Environments (%d):", len(envMap))

	// Sort environment names
	var envNames []string
	for name := range envMap {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	for i, name := range envNames {
		fmt.Printf("%d. %s\n", i+1, color.GreenString(name))

		vars := envMap[name]
		fmt.Printf("   Variables: %s\n", color.CyanString("%d", len(vars)))

		if len(vars) > 0 {
			fmt.Printf("   Created: %s\n", vars[0].CreatedAt.Format("2006-01-02 15:04:05"))
		}

		if i < len(envNames)-1 {
			fmt.Println()
		}
	}

	return nil
}

func (ec *EnvCommand) createEnvironment() error {
	// Check if environment already exists
	envCommands, err := ec.repository.SearchByTag("environment")
	if err != nil {
		ec.logger.Error("Failed to search environments", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to check existing environments", err)
	}

	for _, cmd := range envCommands {
		if cmd.Category == "environment" && extractEnvName(cmd.Name) == ec.envName {
			return errors.NewError(errors.ErrInvalidCommand,
				fmt.Sprintf("environment '%s' already exists", ec.envName), nil)
		}
	}

	// Create environment metadata command
	envType := "user"
	if ec.global {
		envType = "global"
	}

	envCommand := models.Command{
		Entity: models.Entity{
			ID: fmt.Sprintf("ENV-%s-%d", ec.envName, generateTimestamp()),
		},
		Name:     fmt.Sprintf("env:%s:meta", ec.envName),
		Command:  "# Environment metadata",
		Category: "environment",
		Tags:     []string{"environment", ec.envName, "meta"},
		Status:   true,
		Variables: map[string]string{
			"env_name": ec.envName,
			"env_type": envType,
		},
	}

	if err := ec.repository.Put(context.Background(), envCommand); err != nil {
		ec.logger.Error("Failed to create environment",
			zap.String("envName", ec.envName), zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite,
			"failed to create environment", err)
	}

	color.Green("✅ Environment '%s' created successfully", ec.envName)
	color.Cyan("\nNext steps:")
	color.White("  ambros env set %s KEY value", ec.envName)
	color.White("  ambros env show %s", ec.envName)
	ec.logger.Info("Environment created", zap.String("envName", ec.envName), zap.String("envType", envType))

	// If interactive flag is set, prompt user to add variables now
	if ec.interactive {
		reader := bufio.NewReader(os.Stdin)
		color.Cyan("\nInteractive mode: enter variables in the form KEY=VALUE. Empty line to finish.")
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				color.Yellow("Invalid format, expected KEY=VALUE")
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if err := ec.setVariable(ec.envName, key, value); err != nil {
				color.Red("Failed to set variable: %v", err)
			}
		}
	}
	return nil
}

func (ec *EnvCommand) deleteEnvironment() error {
	// Find all commands for this environment
	envCommands, err := ec.repository.SearchByTag(ec.envName)
	if err != nil {
		ec.logger.Error("Failed to search environment", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to find environment", err)
	}

	envFound := false
	var commandsToDelete []models.Command

	for _, cmd := range envCommands {
		if cmd.Category == "environment" && extractEnvName(cmd.Name) == ec.envName {
			envFound = true
			commandsToDelete = append(commandsToDelete, cmd)
		}
	}

	if !envFound {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("environment not found: %s", ec.envName), nil)
	}

	// Delete all environment commands
	for _, cmd := range commandsToDelete {
		if err := ec.repository.Delete(cmd.ID); err != nil {
			ec.logger.Error("Failed to delete environment command",
				zap.String("commandId", cmd.ID), zap.Error(err))
			// Continue deleting other commands
		}
	}

	color.Green("🗑️  Environment '%s' deleted successfully", ec.envName)
	color.Cyan("Deleted %d variables", len(commandsToDelete)-1) // -1 for metadata

	ec.logger.Info("Environment deleted",
		zap.String("envName", ec.envName),
		zap.Int("deletedCommands", len(commandsToDelete)))
	return nil
}

func (ec *EnvCommand) setVariable(envName, key, value string) error {
	// Check if environment exists
	if !ec.environmentExists(envName) {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("environment not found: %s", envName), nil)
	}

	// Create or update variable command
	varCommand := models.Command{
		Entity: models.Entity{
			ID: fmt.Sprintf("ENV-%s-%s-%d", envName, key, generateTimestamp()),
		},
		Name:     fmt.Sprintf("env:%s:var:%s", envName, key),
		Command:  fmt.Sprintf("export %s=%s", key, value),
		Category: "environment",
		Tags:     []string{"environment", envName, "variable"},
		Status:   true,
		Variables: map[string]string{
			"env_name":  envName,
			"var_key":   key,
			"var_value": value,
		},
	}

	if err := ec.repository.Put(context.Background(), varCommand); err != nil {
		ec.logger.Error("Failed to set environment variable",
			zap.String("envName", envName),
			zap.String("key", key), zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite,
			"failed to set environment variable", err)
	}

	color.Green("✅ Variable set: %s=%s", color.YellowString(key), color.CyanString(value))

	ec.logger.Info("Environment variable set",
		zap.String("envName", envName),
		zap.String("key", key))
	return nil
}

func (ec *EnvCommand) unsetVariable(envName, key string) error {
	// Find the variable command
	envCommands, err := ec.repository.SearchByTag(envName)
	if err != nil {
		ec.logger.Error("Failed to search environment", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to find environment", err)
	}

	var varCommand *models.Command
	for _, cmd := range envCommands {
		if cmd.Category == "environment" &&
			strings.Contains(cmd.Name, fmt.Sprintf("var:%s", key)) {
			varCommand = &cmd
			break
		}
	}

	if varCommand == nil {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("variable '%s' not found in environment '%s'", key, envName), nil)
	}

	// Delete the variable command
	if err := ec.repository.Delete(varCommand.ID); err != nil {
		ec.logger.Error("Failed to delete environment variable",
			zap.String("commandId", varCommand.ID), zap.Error(err))
		return errors.NewError(errors.ErrRepositoryWrite,
			"failed to delete environment variable", err)
	}

	color.Green("🗑️  Variable '%s' removed from environment '%s'", key, envName)

	ec.logger.Info("Environment variable removed",
		zap.String("envName", envName),
		zap.String("key", key))
	return nil
}

func (ec *EnvCommand) showEnvironment() error {
	// Find environment commands
	envCommands, err := ec.repository.SearchByTag(ec.envName)
	if err != nil {
		ec.logger.Error("Failed to search environment", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to find environment", err)
	}

	var metaCommand *models.Command
	var varCommands []models.Command

	for _, cmd := range envCommands {
		if cmd.Category == "environment" && extractEnvName(cmd.Name) == ec.envName {
			if strings.Contains(cmd.Name, ":meta") {
				metaCommand = &cmd
			} else if strings.Contains(cmd.Name, ":var:") {
				varCommands = append(varCommands, cmd)
			}
		}
	}

	if metaCommand == nil {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("environment not found: %s", ec.envName), nil)
	}

	// Display environment details
	color.Cyan("📄 Environment Details:")
	fmt.Printf("Name: %s\n", color.GreenString(ec.envName))
	fmt.Printf("ID: %s\n", metaCommand.ID)
	fmt.Printf("Created: %s\n", metaCommand.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Variables: %s\n", color.CyanString("%d", len(varCommands)))

	if len(varCommands) > 0 {
		fmt.Printf("\n%s Environment Variables:\n", color.YellowString("🔧"))

		// Sort variables by key
		sort.Slice(varCommands, func(i, j int) bool {
			keyI := varCommands[i].Variables["var_key"]
			keyJ := varCommands[j].Variables["var_key"]
			return keyI < keyJ
		})

		for _, cmd := range varCommands {
			key := cmd.Variables["var_key"]
			value := cmd.Variables["var_value"]
			fmt.Printf("  %s = %s\n",
				color.YellowString(key),
				color.CyanString(value))
		}

		fmt.Printf("\n%s Usage:\n", color.GreenString("💡"))
		color.White("  ambros env apply %s \"your-command\"", ec.envName)
	}

	return nil
}

func (ec *EnvCommand) applyEnvironment(envName, command string) error {
	// Get environment variables
	envCommands, err := ec.repository.SearchByTag(envName)
	if err != nil {
		ec.logger.Error("Failed to search environment", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead,
			"failed to find environment", err)
	}

	envVars := make(map[string]string)
	envFound := false

	for _, cmd := range envCommands {
		if cmd.Category == "environment" && extractEnvName(cmd.Name) == envName {
			envFound = true
			if strings.Contains(cmd.Name, ":var:") {
				key := cmd.Variables["var_key"]
				value := cmd.Variables["var_value"]
				envVars[key] = value
			}
		}
	}

	if !envFound {
		return errors.NewError(errors.ErrCommandNotFound,
			fmt.Sprintf("environment not found: %s", envName), nil)
	}

	color.Cyan("🚀 Applying environment '%s' to command", envName)

	if len(envVars) > 0 {
		color.Yellow("Environment variables:")
		for key, value := range envVars {
			fmt.Printf("  %s = %s\n", color.YellowString(key), color.CyanString(value))
		}
	}

	// Execute the provided command with merged environment
	// Use the shell to allow complex commands and expansions
	var cmd *exec.Cmd
	shell := "sh"
	shellFlag := "-c"
	cmd = exec.Command(shell, shellFlag, command)

	// Merge current environment and override with envVars
	env := os.Environ()
	// Build map for easier override
	envMap := map[string]string{}
	for _, e := range env {
		if kv := strings.SplitN(e, "=", 2); len(kv) == 2 {
			envMap[kv[0]] = kv[1]
		}
	}
	for k, v := range envVars {
		envMap[k] = v
	}
	// Build final env slice
	finalEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = finalEnv

	// If stdin is a TTY, start in a pty for interactive behavior
	if isatty.IsTerminal(os.Stdin.Fd()) {
		ptmx, err := pty.Start(cmd)
		if err != nil {
			ec.logger.Error("Failed to start command in pty", zap.Error(err))
			return errors.NewError(errors.ErrExecutionFailed, "failed to start command", err)
		}
		// copy output to stdout/stderr
		go func() { _, _ = io.Copy(os.Stdout, ptmx) }()
		// Wait for completion
		err = cmd.Wait()
		if err != nil {
			ec.logger.Error("Command failed", zap.Error(err))
			return errors.NewError(errors.ErrExecutionFailed, "command failed", err)
		}
		return nil
	}

	// Non-tty: capture combined output
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		fmt.Print(string(out))
	}
	if err != nil {
		ec.logger.Error("Command execution failed", zap.Error(err))
		if exit, ok := extractExitStatus(err); ok {
			return errors.NewError(errors.ErrExecutionFailed, fmt.Sprintf("command exited with %d", exit), err)
		}
		return errors.NewError(errors.ErrExecutionFailed, "command execution failed", err)
	}

	ec.logger.Info("Environment applied to command",
		zap.String("envName", envName),
		zap.String("command", command),
		zap.Int("variables", len(envVars)))

	return nil
}

// Helper functions
func (ec *EnvCommand) environmentExists(envName string) bool {
	envCommands, err := ec.repository.SearchByTag("environment")
	if err != nil {
		return false
	}

	for _, cmd := range envCommands {
		if cmd.Category == "environment" && extractEnvName(cmd.Name) == envName {
			return true
		}
	}
	return false
}

func extractEnvName(cmdName string) string {
	// Extract environment name from command name like "env:production:meta" or "env:production:var:KEY"
	parts := strings.Split(cmdName, ":")
	if len(parts) >= 2 && parts[0] == "env" {
		return parts[1]
	}
	return ""
}

func generateTimestamp() int64 {
	return time.Now().UnixNano()
}

func (ec *EnvCommand) Command() *cobra.Command {
	return ec.cmd
}
