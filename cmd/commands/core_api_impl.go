package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins"
	"github.com/gi4nks/ambros/v3/internal/repos"
	"github.com/gi4nks/ambros/v3/internal/utils"
)

// coreAPIImpl is a concrete implementation of the plugins.CoreAPI interface.
// It provides Go plugins with controlled access to core Ambros functionalities.
type coreAPIImpl struct {
	logger    *zap.Logger
	config    *utils.Configuration
	repo      repos.RepositoryInterface
	executor  *Executor
	rootCmd   *cobra.Command   // Needed for RegisterCommand
	utilities *utils.Utilities // For ID generation
}

// NewCoreAPIImpl creates a new instance of coreAPIImpl.
func NewCoreAPIImpl(logger *zap.Logger, config *utils.Configuration, repo repos.RepositoryInterface, executor *Executor, rootCmd *cobra.Command) plugins.CoreAPI {
	return &coreAPIImpl{
		logger:    logger,
		config:    config,
		repo:      repo,
		executor:  executor,
		rootCmd:   rootCmd,
		utilities: utils.NewUtilities(logger),
	}
}

// Logger implements plugins.CoreAPI.
func (c *coreAPIImpl) Logger() *zap.Logger {
	return c.logger
}

// GetConfig implements plugins.CoreAPI.
func (c *coreAPIImpl) GetConfig() *utils.Configuration {
	return c.config
}

// GetRepository implements plugins.CoreAPI.
func (c *coreAPIImpl) GetRepository() repos.RepositoryInterface {
	return c.repo
}

// ExecuteCommand implements plugins.CoreAPI.
func (c *coreAPIImpl) ExecuteCommand(ctx context.Context, name string, args []string, opts ...plugins.CommandExecuteOption) (exitCode int, output string, err error) {
	// Apply options
	execOpts := &plugins.CommandExecuteOptions{
		Store:  true,
		DryRun: false,
		Auto:   false,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	for _, opt := range opts {
		opt(execOpts)
	}

	// Handle dry run
	if execOpts.DryRun {
		c.logger.Info("Dry run - command would be executed",
			zap.String("command", name),
			zap.Strings("args", args))
		return 0, fmt.Sprintf("(dry run) Would execute: %s %v", name, args), nil
	}

	// Create command record for storage if needed
	var command *models.Command
	if execOpts.Store && c.repo != nil {
		command = &models.Command{
			Entity: models.Entity{
				ID:        c.utilities.Random(),
				CreatedAt: time.Now(),
			},
			Name:      name,
			Arguments: args,
			Tags:      execOpts.Tag,
			Category:  execOpts.Category,
		}
	}

	// Execute the command
	var execErr error
	if execOpts.Auto {
		// ExecuteCommandAuto streams directly to stdout/stderr
		c.logger.Debug("Executing command in auto mode",
			zap.String("command", name), zap.Strings("args", args))
		exitCode, execErr = c.executor.ExecuteCommandAuto(name, args)
		output = "" // No output string captured in auto mode
	} else {
		// ExecuteCapture captures combined stdout/stderr
		exitCode, output, execErr = c.executor.ExecuteCapture(name, args)
	}

	// Store the command execution if requested
	if execOpts.Store && c.repo != nil && command != nil {
		command.TerminatedAt = time.Now()
		command.Status = exitCode == 0 && execErr == nil
		command.Output = output
		if execErr != nil {
			command.Error = execErr.Error()
		}

		if storeErr := c.repo.Put(ctx, *command); storeErr != nil {
			c.logger.Warn("Failed to store command execution",
				zap.Error(storeErr),
				zap.String("commandId", command.ID))
		} else {
			c.logger.Debug("Stored command execution",
				zap.String("commandId", command.ID),
				zap.Strings("tags", command.Tags),
				zap.String("category", command.Category))
		}
	}

	return exitCode, output, execErr
}

// RegisterCommand implements plugins.CoreAPI.
func (c *coreAPIImpl) RegisterCommand(cmd *cobra.Command) error {
	if c.rootCmd == nil {
		return fmt.Errorf("root command not set in CoreAPIImpl, cannot register plugin command")
	}
	// TODO: Add proper validation to prevent command name collisions
	c.rootCmd.AddCommand(cmd)
	c.logger.Info("Plugin registered new command", zap.String("command_name", cmd.Name()))
	return nil
}

// TriggerHook implements plugins.CoreAPI.
// It triggers both shell-based plugins (via PluginCommand) and Go internal plugins (via registry).
func (c *coreAPIImpl) TriggerHook(hookName string, args []string) error {
	c.logger.Info("Triggering Ambros hook", zap.String("hook_name", hookName), zap.Strings("args", args))

	var errs []error

	// Trigger Go internal plugins via the global registry
	registry := plugins.GetGlobalRegistry()
	for _, plugin := range registry.GetAllPlugins() {
		manifest := plugin.GetManifest()
		// Check if this plugin is subscribed to this hook
		for _, subscribedHook := range manifest.Hooks {
			if subscribedHook == hookName {
				c.logger.Debug("Executing Go plugin hook",
					zap.String("plugin", manifest.Name),
					zap.String("hook", hookName))
				if err := plugin.HandleHook(context.Background(), hookName, args); err != nil {
					c.logger.Warn("Go plugin hook execution failed",
						zap.String("plugin", manifest.Name),
						zap.String("hook", hookName),
						zap.Error(err))
					errs = append(errs, fmt.Errorf("plugin '%s' hook '%s': %w", manifest.Name, hookName, err))
				}
				break // Don't execute same plugin multiple times for same hook
			}
		}
	}

	// Note: Shell plugins are triggered via ExecuteHooks(pc, hookName, args) in plugin_runner.go
	// which is called from main.go for "pre-run" and "post-run" hooks.
	// Plugins can trigger additional hooks via this CoreAPI method.

	if len(errs) > 0 {
		return fmt.Errorf("some hook handlers failed: %v", errs)
	}
	return nil
}
