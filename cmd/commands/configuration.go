package commands

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color" // New import
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/utils"
)

type ConfigurationCommand struct {
	*BaseCommand
	format       string
	globalConfig *utils.Configuration // Added to access global config instance
}

func NewConfigurationCommand(logger *zap.Logger, globalConfig *utils.Configuration) *ConfigurationCommand {
	cc := &ConfigurationCommand{
		globalConfig: globalConfig, // Store the global config
	}

	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Show application configuration",
		Long: `Display the current application configuration including repository settings,
paths, and other application parameters.

Examples:
  ambros configuration                       # Show configuration
  ambros configuration --format json         # Show as JSON
  ambros configuration validate              # Validate configuration file`,
		RunE: cc.runE,
	}

	// Add subcommands
	cmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate the current application configuration",
		Long:  `Checks the .ambros.yaml configuration file for syntax errors and valid values.`,
		RunE:  cc.validateConfig,
	})

	// Configuration command doesn't need repository
	cc.BaseCommand = NewBaseCommandWithoutRepo(cmd, logger)
	cc.cmd = cmd
	cc.setupFlags(cmd)
	return cc
}

func (cc *ConfigurationCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cc.format, "format", "f", "text", "Output format (text or json)")
}

func (cc *ConfigurationCommand) runE(cmd *cobra.Command, _ []string) error {
	cc.logger.Debug("Configuration command invoked",
		zap.String("format", cc.format))

	// Get configuration

	switch cc.format {
	case "json":
		return cc.outputJSON()
	case "text":
		return cc.outputText()
	default:
		fmt.Printf("Unsupported format: %s\n", cc.format)
		return nil
	}
}

func (cc *ConfigurationCommand) outputJSON() error {
	configData := map[string]interface{}{
		"repositoryDirectory": cc.globalConfig.RepositoryDirectory,
		"repositoryFile":      cc.globalConfig.RepositoryFile,
		"repositoryFullPath":  cc.globalConfig.RepositoryFullName(),
		"lastCountDefault":    cc.globalConfig.LastCountDefault,
		"debugMode":           cc.globalConfig.DebugMode,
		"server": map[string]interface{}{
			"host": cc.globalConfig.Server.Host,
			"port": cc.globalConfig.Server.Port,
			"cors": cc.globalConfig.Server.Cors,
		},
		"plugins": map[string]interface{}{
			"directory":   cc.globalConfig.Plugins.Directory,
			"allowUnsafe": cc.globalConfig.Plugins.AllowUnsafe,
		},
		"scheduler": map[string]interface{}{
			"enabled":       cc.globalConfig.Scheduler.Enabled,
			"checkInterval": cc.globalConfig.Scheduler.CheckInterval,
		},
		"analytics": map[string]interface{}{
			"enabled":       cc.globalConfig.Analytics.Enabled,
			"retentionDays": cc.globalConfig.Analytics.RetentionDays,
		},
	}

	jsonData, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		cc.logger.Error("Failed to marshal configuration to JSON", zap.Error(err))
		return err
	}

	fmt.Println(string(jsonData))

	cc.logger.Debug("Configuration displayed in JSON format")
	return nil
}

func (cc *ConfigurationCommand) outputText() error {
	fmt.Printf("Ambros Configuration:\n")
	fmt.Printf("  Repository Directory: %s\n", cc.globalConfig.RepositoryDirectory)
	fmt.Printf("  Repository File:      %s\n", cc.globalConfig.RepositoryFile)
	fmt.Printf("  Full Repository Path: %s\n", cc.globalConfig.RepositoryFullName())
	fmt.Printf("  Last Count Default:   %d\n", cc.globalConfig.LastCountDefault)
	fmt.Printf("  Debug Mode:           %t\n", cc.globalConfig.DebugMode)

	fmt.Printf("\n  Server:\n")
	fmt.Printf("    Host: %s\n", cc.globalConfig.Server.Host)
	fmt.Printf("    Port: %d\n", cc.globalConfig.Server.Port)
	fmt.Printf("    CORS: %t\n", cc.globalConfig.Server.Cors)

	fmt.Printf("\n  Plugins:\n")
	fmt.Printf("    Directory: %s\n", cc.globalConfig.Plugins.Directory)
	fmt.Printf("    Allow Unsafe: %t\n", cc.globalConfig.Plugins.AllowUnsafe)

	fmt.Printf("\n  Scheduler:\n")
	fmt.Printf("    Enabled: %t\n", cc.globalConfig.Scheduler.Enabled)
	fmt.Printf("    Check Interval: %s\n", cc.globalConfig.Scheduler.CheckInterval)

	fmt.Printf("\n  Analytics:\n")
	fmt.Printf("    Enabled: %t\n", cc.globalConfig.Analytics.Enabled)
	fmt.Printf("    Retention Days: %d\n", cc.globalConfig.Analytics.RetentionDays)

	cc.logger.Debug("Configuration displayed in text format")
	return nil
}

func (cc *ConfigurationCommand) validateConfig(cmd *cobra.Command, args []string) error {
	cc.logger.Debug("Validate configuration command invoked")

	// The configuration is already loaded by initConfig, so we just need to validate it
	if err := cc.globalConfig.Validate(); err != nil {
		color.Red("❌ Configuration validation failed!")
		fmt.Printf("Error: %v\n", err)
		cc.logger.Error("Configuration validation failed", zap.Error(err))
		return err // Return the error to Cobra for non-zero exit code
	}

	color.Green("✅ Configuration is valid!")
	cc.logger.Info("Configuration validated successfully")
	return nil
}
