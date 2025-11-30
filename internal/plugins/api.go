package plugins

import (
	"context"
	"io"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos"
	"github.com/gi4nks/ambros/v3/internal/utils"
)

// CoreAPI defines the interface for core Ambros functionalities exposed to plugins.
// Plugins can use these methods to interact with the Ambros system.
type CoreAPI interface {
	// Logger returns the application's logger.
	Logger() *zap.Logger

	// GetConfig returns the global application configuration.
	GetConfig() *utils.Configuration

	// GetRepository returns the main data repository interface.
	GetRepository() repos.RepositoryInterface

	// ExecuteCommand runs an external command.
	// It mirrors the functionality of `ambros run`.
	ExecuteCommand(ctx context.Context, name string, args []string, opts ...CommandExecuteOption) (exitCode int, output string, err error)

	// RegisterCommand allows a plugin to register a new Cobra command with the main application.
	RegisterCommand(cmd *cobra.Command) error

	// TriggerHook allows a plugin to trigger an Ambros hook.
	TriggerHook(hookName string, args []string) error
}

// CommandExecuteOption defines options for executing a command via CoreAPI.
type CommandExecuteOption func(*CommandExecuteOptions)

// CommandExecuteOptions holds options for command execution.
type CommandExecuteOptions struct {
	Store    bool
	Tag      []string
	Category string
	DryRun   bool
	Auto     bool // Attach TTY
	Stdout   io.Writer
	Stderr   io.Writer
}

// WithStore enables/disables storing command execution details.
func WithStore(store bool) CommandExecuteOption {
	return func(opts *CommandExecuteOptions) { opts.Store = store }
}

// WithTag adds tags to the command.
func WithTag(tags ...string) CommandExecuteOption {
	return func(opts *CommandExecuteOptions) { opts.Tag = append(opts.Tag, tags...) }
}

// WithCategory assigns a category to the command.
func WithCategory(category string) CommandExecuteOption {
	return func(opts *CommandExecuteOptions) { opts.Category = category }
}

// WithDryRun enables/disables dry run mode.
func WithDryRun(dryRun bool) CommandExecuteOption {
	return func(opts *CommandExecuteOptions) { opts.DryRun = dryRun }
}

// WithAuto enables/disables auto (TTY) mode.
func WithAuto(auto bool) CommandExecuteOption {
	return func(opts *CommandExecuteOptions) { opts.Auto = auto }
}

// WithStdout sets the stdout writer for the command.
func WithStdout(w io.Writer) CommandExecuteOption {
	return func(opts *CommandExecuteOptions) { opts.Stdout = w }
}

// WithStderr sets the stderr writer for the command.
func WithStderr(w io.Writer) CommandExecuteOption {
	return func(opts *CommandExecuteOptions) { opts.Stderr = w }
}

// GoPlugin defines the interface that all Go-based internal plugins must implement.
type GoPlugin interface {
	// GetManifest returns the plugin's manifest. This is used by Ambros to understand
	// the plugin's capabilities and metadata.
	GetManifest() models.Plugin

	// Init initializes the plugin, providing it with access to the CoreAPI.
	// This method is called once when the plugin is loaded.
	Init(api CoreAPI) error

	// Run executes a specific command provided by the plugin.
	// The commandName corresponds to one of the names defined in the manifest.
	// args are the arguments passed to that command.
	Run(ctx context.Context, commandName string, args []string, stdout, stderr io.Writer) error

	// HandleHook is called when an Ambros hook is triggered that the plugin
	// has subscribed to (via its manifest).
	HandleHook(ctx context.Context, hookName string, args []string) error
}
