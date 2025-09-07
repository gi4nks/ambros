package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/utils"
)

type ConfigurationCommand struct {
	*BaseCommand
	format string
}

func NewConfigurationCommand(logger *zap.Logger) *ConfigurationCommand {
	cc := &ConfigurationCommand{}

	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Show application configuration",
		Long: `Display the current application configuration including repository settings,
paths, and other application parameters.

Examples:
  ambros configuration              # Show configuration
  ambros configuration --format json   # Show as JSON`,
		RunE: cc.runE,
	}

	// Configuration command doesn't need repository
	cc.BaseCommand = NewBaseCommandWithoutRepo(cmd, logger)
	cc.cmd = cmd
	cc.setupFlags(cmd)
	return cc
}

func (cc *ConfigurationCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cc.format, "format", "f", "text", "Output format (text or json)")
}

func (cc *ConfigurationCommand) runE(cmd *cobra.Command, args []string) error {
	cc.logger.Debug("Configuration command invoked",
		zap.String("format", cc.format))

	// Get configuration
	config := utils.NewConfiguration(cc.logger)

	switch cc.format {
	case "json":
		return cc.outputJSON(config)
	case "text":
		return cc.outputText(config)
	default:
		fmt.Printf("Unsupported format: %s\n", cc.format)
		return nil
	}
}

func (cc *ConfigurationCommand) outputJSON(config *utils.Configuration) error {
	configData := map[string]interface{}{
		"repositoryDirectory": config.RepositoryDirectory,
		"repositoryFile":      config.RepositoryFile,
		"repositoryFullPath":  config.RepositoryFullName(),
		"lastCountDefault":    config.LastCountDefault,
		"debugMode":           config.DebugMode,
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

func (cc *ConfigurationCommand) outputText(config *utils.Configuration) error {
	fmt.Printf("Ambros Configuration:\n")
	fmt.Printf("  Repository Directory: %s\n", config.RepositoryDirectory)
	fmt.Printf("  Repository File:      %s\n", config.RepositoryFile)
	fmt.Printf("  Full Repository Path: %s\n", config.RepositoryFullName())
	fmt.Printf("  Last Count Default:   %d\n", config.LastCountDefault)
	fmt.Printf("  Debug Mode:           %t\n", config.DebugMode)

	cc.logger.Debug("Configuration displayed in text format")
	return nil
}

func (cc *ConfigurationCommand) Command() *cobra.Command {
	return cc.cmd
}
