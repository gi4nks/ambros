package commands

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
	"github.com/gi4nks/ambros/v3/internal/repos"
	"github.com/gi4nks/ambros/v3/internal/utils"
)

var (
	cfgFile        string
	debug          bool
	pluginCmd      *PluginCommand
	globalConfig   *utils.Configuration // Global variable to hold the configuration
	globalExecutor *Executor            // Global variable to hold the executor
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ambros",
	Short: "The command butler!",
	Long: `Ambros creates a local history of executed commands, keeping track of output and metadata.
It helps you manage, search, and replay commands with ease.

Examples:
  ambros run -- ls -la              # Run and store a command
  ambros last                       # Show recent commands
  ambros search "grep"              # Search command history
  ambros export -o backup.json      # Export command history`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ambros.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")

	// Initialize logger (before config load, as config might influence logging)
	logger := initLogger()
	globalConfig = utils.NewConfiguration(logger) // Initialize global config with logger
	globalExecutor = NewExecutor(logger)          // Initialize global executor

	// Add subcommands that don't require repository (yet)
	rootCmd.AddCommand(NewConfigurationCommand(logger, globalConfig).Command()) // Pass globalConfig
}

// InitializeRepository initializes the repository and adds repository-dependent commands
func InitializeRepository() error {
	logger := initLogger()

	// Initialize repository
	repo, err := repos.NewRepository(globalConfig.RepositoryFullName(), logger)
	if err != nil {
		return err
	}

	// Create CoreAPI implementation
	coreAPI := NewCoreAPIImpl(logger, globalConfig, repo, globalExecutor, rootCmd)

	// Set CoreAPI in the global plugin registry
	plugins.GetGlobalRegistry().SetCoreAPI(coreAPI)

	// Add repository-dependent commands
	addRepositoryCommands(logger, repo, coreAPI)
	return nil
}

func addRepositoryCommands(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) {
	rootCmd.AddCommand(NewVersionCommand(logger, api).Command())
	rootCmd.AddCommand(NewIntegrateCommand(logger, api).Command())

	// Add all command implementations that require repository
	rootCmd.AddCommand(NewRunCommand(logger, repo, api).Command())
	rootCmd.AddCommand(NewLastCommand(logger, repo, api).Command())
	rootCmd.AddCommand(NewSearchCommand(logger, repo, api).Command())
	rootCmd.AddCommand(NewOutputCommand(logger, repo, api).Command())
	rootCmd.AddCommand(NewAnalyticsCommand(logger, repo, api).Command())
	rootCmd.AddCommand(NewInteractiveCommand(logger, repo, api).Command())
	rootCmd.AddCommand(NewExportCommand(logger, repo, api).Command())
	rootCmd.AddCommand(NewImportCommand(logger, repo, api).Command())
	// Unified rerun command replaces recall/revive
	rootCmd.AddCommand(NewRerunCommand(logger, repo, api).Command())
	// Database utilities
	rootCmd.AddCommand(NewDBCommand(logger, repo).Command())
	// MCP server
	rootCmd.AddCommand(NewMCPCommand(logger, repo, api).Command())

	// Plugin commands
	pluginCmd = NewPluginCommand(logger, repo) // Store in package-level variable
	rootCmd.AddCommand(pluginCmd.Command())
}

// GetPluginCommand returns the global PluginCommand instance for hook execution.
// May return nil if plugins are not yet initialized.
func GetPluginCommand() *PluginCommand {
	return pluginCmd
}

func initLogger() *zap.Logger {
	var logger *zap.Logger
	var err error

	if debug {
		logger, err = zap.NewDevelopment()
	} else {
		// For CLI tools, we want quieter logging in production
		// Only log warnings and errors to avoid cluttering user output
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		logger, err = config.Build()
	}

	if err != nil {
		panic(err)
	}

	return logger
}

func initConfig() {
	// Load configuration
	if err := globalConfig.Load(cfgFile); err != nil {
		globalConfig.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Re-initialize logger if debug mode changed after config load
	if globalConfig.DebugMode && !debug {
		debug = true // Force debug mode on if enabled in config
		globalConfig.Logger.Info("Debug mode enabled via config file, re-initializing logger")
		initLogger() // Re-init logger to apply debug settings
	}
}
