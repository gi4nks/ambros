package commands

import (
	"context"
	"encoding/json" // New import
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"  // New import
	"github.com/gi4nks/ambros/v3/internal/plugins" // For InternalPluginRegistry
)

// RegisterPluginCommands scans the plugin directory and registers a command for each enabled plugin.
func RegisterPluginCommands(rootCmd *cobra.Command, pc *PluginCommand) {
	plugins, err := pc.getEnabledPlugins()
	if err != nil {
		pc.logger.Warn("Failed to get enabled plugins", zap.Error(err))
		return
	}

	for _, plugin := range plugins {
		// Capture variables for the closure
		p := plugin

		// Create a top-level command for the plugin itself, e.g., `ambros my-plugin`
		pluginCmd := &cobra.Command{
			Use:   p.Name,
			Short: fmt.Sprintf("Commands provided by the '%s' plugin", p.Name),
		}

		if len(p.Commands) == 0 {
			// If the plugin has no subcommands, make the top-level command do something.
			// For now, we can just have it print info.
			pluginCmd.RunE = func(cmd *cobra.Command, args []string) error {
				fmt.Fprintf(cmd.OutOrStdout(), "Plugin '%s': %s\n", p.Name, p.Description)
				fmt.Fprintln(cmd.OutOrStdout(), "This plugin provides no direct commands.")
				return nil
			}
		}

		// Register each command defined in the plugin's manifest
		for _, cmdDef := range p.Commands {
			// Capture variables for the closure
			def := cmdDef

			subCmd := &cobra.Command{
				Use:   def.Name,
				Short: def.Description,
				Long:  def.Usage,
				RunE: func(cmd *cobra.Command, args []string) error {
					return executePluginCommand(pc, p, def.Name, args, cmd.OutOrStdout(), cmd.ErrOrStderr())
				},
			}
			pluginCmd.AddCommand(subCmd)
		}

		rootCmd.AddCommand(pluginCmd)
	}
}

// executePluginCommand handles the execution of a plugin's command.
func executePluginCommand(pc *PluginCommand, p *models.Plugin, commandName string, args []string, stdout, stderr io.Writer) error {
	return executePlugin(pc, p, commandName, args, stdout, stderr)
}

// ExecuteHooks runs all hooks for a given event.
func ExecuteHooks(pc *PluginCommand, event string, args []string) {
	plugins, err := pc.getEnabledPlugins()
	if err != nil {
		pc.logger.Warn("Failed to get enabled plugins for hooks", zap.Error(err))
		return
	}

	for _, plugin := range plugins {
		for _, hook := range plugin.Hooks {
			if hook == event {
				pc.logger.Info("Executing hook",
					zap.String("plugin", plugin.Name),
					zap.String("hook", event),
				)
				// Execute the hook. The first argument is the hook event name.
				err := executePlugin(pc, plugin, event, args, os.Stdout, os.Stderr)
				if err != nil {
					pc.logger.Warn("Hook execution failed",
						zap.String("plugin", plugin.Name),
						zap.String("hook", event),
						zap.Error(err),
					)
				}
			}
		}
	}
}

// executePlugin handles the execution of a plugin's executable.
func executePlugin(pc *PluginCommand, p *models.Plugin, commandName string, args []string, stdout, stderr io.Writer) error {
	pc.logger.Info("Executing plugin",
		zap.String("plugin", p.Name),
		zap.String("command", commandName),
		zap.Strings("args", args))

	switch p.Type {
	case models.PluginTypeShell:
		// For shell plugins, execute the script
		// Validate the executable path to ensure it's safe.
		if err := pc.validateExecutablePath(p.Name, p.Executable); err != nil {
			return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("plugin '%s' has an invalid executable path", p.Name), err)
		}

		pluginDir, err := pc.pluginDirPath(p.Name)
		if err != nil {
			return err
		}
		execPath := filepath.Join(pluginDir, p.Executable)

		// Prepare the command to run. The first argument to the script is the command name.
		cmdArgs := append([]string{commandName}, args...)
		cmd := exec.Command(execPath, cmdArgs...)
		cmd.Dir = pluginDir // Run the command from the plugin's directory.

		// Set environment variables for the plugin
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("AMBROS_PLUGIN_NAME=%s", p.Name))
		cmd.Env = append(cmd.Env, fmt.Sprintf("AMBROS_PLUGIN_COMMAND=%s", commandName))
		// TODO: Add other AMBROS_ environment variables as needed.
		// Add plugin config as env var
		if configJSON, err := json.Marshal(p.Config); err == nil {
			cmd.Env = append(cmd.Env, fmt.Sprintf("AMBROS_PLUGIN_CONFIG=%s", configJSON))
		} else {
			pc.logger.Warn("Failed to marshal plugin config to JSON for environment variable", zap.String("plugin", p.Name), zap.Error(err))
		}

		// Attach stdin, stdout, and stderr to allow interactive plugins.
		cmd.Stdin = os.Stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				pc.logger.Warn("Plugin exited with an error",
					zap.String("plugin", p.Name),
					zap.String("command", commandName),
					zap.Int("exit_code", exitErr.ExitCode()))
				return nil
			}
			// A different error occurred (e.g., command not found).
			return errors.NewError(errors.ErrExecutionFailed, fmt.Sprintf("failed to run plugin '%s'", p.Name), err)
		}

	case models.PluginTypeGoInternal:
		// For Go internal plugins, use the InternalPluginRegistry
		registry := plugins.GetGlobalRegistry()
		goPlugin, found := registry.GetPlugin(p.Name)
		if !found {
			pc.logger.Error("Go internal plugin not found in registry",
				zap.String("plugin", p.Name))
			return errors.NewError(errors.ErrNotFound,
				fmt.Sprintf("Go internal plugin '%s' not found in registry", p.Name), nil)
		}

		// Execute the plugin command
		ctx := context.Background()
		if err := goPlugin.Run(ctx, commandName, args, stdout, stderr); err != nil {
			pc.logger.Warn("Go plugin execution failed",
				zap.String("plugin", p.Name),
				zap.String("command", commandName),
				zap.Error(err))
			return errors.NewError(errors.ErrExecutionFailed,
				fmt.Sprintf("Go internal plugin '%s' command '%s' failed", p.Name, commandName), err)
		}

		pc.logger.Info("Go plugin executed successfully",
			zap.String("plugin", p.Name),
			zap.String("command", commandName))

	default:
		return errors.NewError(errors.ErrExecutionFailed,
			fmt.Sprintf("unsupported plugin type '%s' for plugin '%s'", p.Type, p.Name), nil)
	}

	return nil
}

// getEnabledPlugins returns a list of all enabled plugins.
func (pc *PluginCommand) getEnabledPlugins() ([]*models.Plugin, error) {
	pluginsDir := pc.getPluginsDirectory()
	logger := pc.logger

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No plugins directory, no enabled plugins.
		}
		return nil, errors.NewError(errors.ErrInternalServer, "failed to read plugins directory", err)
	}

	var enabledPlugins []*models.Plugin
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginName := entry.Name()
		plugin, err := pc.loadPlugin(pluginName)
		if err != nil {
			logger.Warn("Failed to load invalid plugin", zap.String("plugin", pluginName), zap.Error(err))
			continue
		}

		if plugin.Enabled {
			enabledPlugins = append(enabledPlugins, plugin)
		}
	}

	return enabledPlugins, nil
}
