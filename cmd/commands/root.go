package commands

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/repos"
	"github.com/gi4nks/ambros/v3/internal/utils"
)

var (
	cfgFile string
	debug   bool
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

	// Initialize logger
	logger := initLogger()

	// Initialize repository
	config := utils.NewConfiguration(logger)
	repo, err := repos.NewRepository(config.RepositoryFullName(), logger)
	if err != nil {
		logger.Fatal("Failed to initialize repository", zap.Error(err))
	}

	// Add subcommands
	addCommands(logger, repo)
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

func addCommands(logger *zap.Logger, repo RepositoryInterface) {
	// Add all command implementations
	rootCmd.AddCommand(NewRunCommand(logger, repo).Command())
	rootCmd.AddCommand(NewLastCommand(logger, repo).Command())
	rootCmd.AddCommand(NewSearchCommand(logger, repo).Command())
	rootCmd.AddCommand(NewOutputCommand(logger, repo).Command())
	rootCmd.AddCommand(NewTemplateCommand(logger, repo).Command())
	rootCmd.AddCommand(NewAnalyticsCommand(logger, repo).Command())
	rootCmd.AddCommand(NewEnvCommand(logger, repo).Command())
	rootCmd.AddCommand(NewInteractiveCommand(logger, repo).Command())
	rootCmd.AddCommand(NewIntegrateCommand(logger).Command())
	rootCmd.AddCommand(NewExportCommand(logger, repo).Command())
	rootCmd.AddCommand(NewImportCommand(logger, repo).Command())
	rootCmd.AddCommand(NewLogsCommand(logger, repo).Command())
	rootCmd.AddCommand(NewChainCommand(logger, repo).Command())
	rootCmd.AddCommand(NewStoreCommand(logger, repo).Command())
	rootCmd.AddCommand(NewRecallCommand(logger, repo).Command())
	rootCmd.AddCommand(NewReviveCommand(logger, repo).Command())
	rootCmd.AddCommand(NewSchedulerCommand(logger, repo).Command())
	// Database utilities
	rootCmd.AddCommand(NewDBCommand(logger, repo).Command())

	// Additional server and plugin commands
	rootCmd.AddCommand(NewServerCommand(logger, repo).Command())
	rootCmd.AddCommand(NewPluginCommand(logger, repo).Command())

	// Commands that don't need repository
	rootCmd.AddCommand(NewVersionCommand(logger).Command())
	rootCmd.AddCommand(NewConfigurationCommand(logger).Command())
}

func initConfig() {
	// Configuration initialization can be added here
	// For now, we'll use the default configuration
}
