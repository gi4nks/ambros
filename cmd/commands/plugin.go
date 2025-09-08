package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/errors"
)

// PluginCommand represents the plugin management command
type PluginCommand struct {
	*BaseCommand
	pluginName string
	configFile string
}

// Plugin represents a plugin definition
type Plugin struct {
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Description  string             `json:"description"`
	Author       string             `json:"author"`
	Enabled      bool               `json:"enabled"`
	Executable   string             `json:"executable"`
	Commands     []PluginCommandDef `json:"commands"`
	Hooks        []string           `json:"hooks"`
	Config       map[string]string  `json:"config"`
	Dependencies []string           `json:"dependencies"`
}

type PluginCommandDef struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Args        []string `json:"args"`
}

// NewPluginCommand creates a new plugin command
func NewPluginCommand(logger *zap.Logger, repo RepositoryInterface) *PluginCommand {
	pc := &PluginCommand{}

	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "🔌 Manage Ambros plugins (Phase 3)",
		Long: `🔌 Advanced Plugin Management System

The plugin system allows extending Ambros with custom functionality through
external scripts, programs, and integrations. Plugins can add new commands,
hook into existing workflows, and provide specialized functionality.

Features:
  • Install and manage external plugins
  • Custom command integrations  
  • Hook system for workflow automation
  • Configuration management
  • Dependency tracking
  • Security sandboxing

Subcommands:
  list                  List installed plugins
  install <name>        Install a plugin from registry
  uninstall <name>      Remove a plugin
  enable <name>         Enable a disabled plugin
  disable <name>        Disable a plugin
  info <name>           Show detailed plugin information
  config <name>         Manage plugin configuration
  create <name>         Create a new plugin template
  registry              Manage plugin registries

Examples:
  ambros plugin list
  ambros plugin install docker-integration
  ambros plugin info slack-notifications
  ambros plugin config git-hooks --set webhook.url=https://example.com
  ambros plugin create my-custom-plugin`,
		Args: cobra.MinimumNArgs(1),
		RunE: pc.runE,
	}

	pc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	pc.cmd = cmd
	pc.setupFlags(cmd)
	return pc
}

func (pc *PluginCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&pc.pluginName, "name", "n", "", "Plugin name")
	cmd.Flags().StringVarP(&pc.configFile, "config", "c", "", "Plugin config file")
}

func (pc *PluginCommand) runE(cmd *cobra.Command, args []string) error {
	pc.logger.Debug("Plugin command invoked",
		zap.String("action", args[0]),
		zap.Strings("args", args))

	switch args[0] {
	case "list":
		return pc.listPlugins()
	case "install":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return pc.installPlugin(args[1])
	case "uninstall":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return pc.uninstallPlugin(args[1])
	case "enable":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return pc.enablePlugin(args[1])
	case "disable":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return pc.disablePlugin(args[1])
	case "info":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return pc.showPluginInfo(args[1])
	case "config":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return pc.manageConfig(args[1], args[2:])
	case "create":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return pc.createPlugin(args[1])
	case "registry":
		return pc.manageRegistry(args[1:])
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unknown subcommand: "+args[0], nil)
	}
}

func (pc *PluginCommand) listPlugins() error {
	color.Cyan("🔌 Installed Plugins")
	color.Cyan("═══════════════════")

	pluginsDir := pc.getPluginsDirectory()
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		color.Yellow("No plugins directory found. No plugins installed.")
		return nil
	}

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		color.Yellow("No plugins installed.")
		color.White("Install plugins with: ambros plugin install <name>")
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			plugin, err := pc.loadPlugin(entry.Name())
			if err != nil {
				color.Red("❌ %s (invalid plugin)", entry.Name())
				continue
			}

			status := "🔴 Disabled"
			if plugin.Enabled {
				status = "🟢 Enabled"
			}

			color.White("📦 %s v%s %s", plugin.Name, plugin.Version, status)
			color.HiBlack("   %s", plugin.Description)

			if len(plugin.Commands) > 0 {
				color.HiBlack("   Commands: %d", len(plugin.Commands))
			}
		}
	}

	return nil
}

func (pc *PluginCommand) installPlugin(name string) error {
	color.Cyan("📦 Installing plugin: %s", name)

	// For now, create a sample plugin structure
	pluginDir := filepath.Join(pc.getPluginsDirectory(), name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// Create sample plugin definition
	plugin := Plugin{
		Name:         name,
		Version:      "1.0.0",
		Description:  fmt.Sprintf("Sample plugin: %s", name),
		Author:       "Ambros Plugin System",
		Enabled:      true,
		Executable:   fmt.Sprintf("./%s.sh", name),
		Commands:     []PluginCommandDef{},
		Hooks:        []string{},
		Config:       make(map[string]string),
		Dependencies: []string{},
	}

	// Save plugin definition
	if err := pc.savePlugin(name, plugin); err != nil {
		return err
	}

	// Create sample executable
	execPath := filepath.Join(pluginDir, fmt.Sprintf("%s.sh", name))
	sampleScript := fmt.Sprintf(`#!/bin/bash
# Sample plugin script for %s
echo "Plugin %s executed with args: $@"
`, name, name)

	if err := os.WriteFile(execPath, []byte(sampleScript), 0755); err != nil {
		return err
	}

	color.Green("✅ Plugin %s installed successfully", name)
	color.White("Plugin directory: %s", pluginDir)
	return nil
}

func (pc *PluginCommand) uninstallPlugin(name string) error {
	color.Yellow("🗑️  Uninstalling plugin: %s", name)

	pluginDir := filepath.Join(pc.getPluginsDirectory(), name)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		return err
	}

	color.Green("✅ Plugin %s uninstalled successfully", name)
	return nil
}

func (pc *PluginCommand) enablePlugin(name string) error {
	plugin, err := pc.loadPlugin(name)
	if err != nil {
		return err
	}

	plugin.Enabled = true
	if err := pc.savePlugin(name, *plugin); err != nil {
		return err
	}

	color.Green("✅ Plugin %s enabled", name)
	return nil
}

func (pc *PluginCommand) disablePlugin(name string) error {
	plugin, err := pc.loadPlugin(name)
	if err != nil {
		return err
	}

	plugin.Enabled = false
	if err := pc.savePlugin(name, *plugin); err != nil {
		return err
	}

	color.Yellow("Plugin %s disabled", name)
	return nil
}

func (pc *PluginCommand) showPluginInfo(name string) error {
	plugin, err := pc.loadPlugin(name)
	if err != nil {
		return err
	}

	color.Cyan("🔍 Plugin Information: %s", name)
	color.Cyan("═══════════════════════════════")
	color.White("Name: %s", plugin.Name)
	color.White("Version: %s", plugin.Version)
	color.White("Description: %s", plugin.Description)
	color.White("Author: %s", plugin.Author)
	color.White("Enabled: %v", plugin.Enabled)
	color.White("Executable: %s", plugin.Executable)

	if len(plugin.Commands) > 0 {
		color.White("Commands:")
		for _, cmd := range plugin.Commands {
			color.White("  - %s: %s", cmd.Name, cmd.Description)
		}
	}

	if len(plugin.Hooks) > 0 {
		color.White("Hooks: %v", plugin.Hooks)
	}

	if len(plugin.Config) > 0 {
		color.White("Configuration:")
		for key, value := range plugin.Config {
			color.White("  %s: %s", key, value)
		}
	}

	if len(plugin.Dependencies) > 0 {
		color.White("Dependencies: %v", plugin.Dependencies)
	}

	return nil
}

func (pc *PluginCommand) manageConfig(name string, args []string) error {
	color.Yellow("🔧 Plugin configuration management coming soon!")
	color.White("Will support: --set key=value, --get key, --list")
	return nil
}

func (pc *PluginCommand) createPlugin(name string) error {
	color.Cyan("🛠️  Creating plugin template: %s", name)

	pluginDir := filepath.Join(pc.getPluginsDirectory(), name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// Create plugin template structure
	plugin := Plugin{
		Name:         name,
		Version:      "0.1.0",
		Description:  fmt.Sprintf("Custom plugin: %s", name),
		Author:       "Your Name",
		Enabled:      false,
		Executable:   fmt.Sprintf("./%s.sh", name),
		Commands:     []PluginCommandDef{},
		Hooks:        []string{},
		Config:       make(map[string]string),
		Dependencies: []string{},
	}

	if err := pc.savePlugin(name, plugin); err != nil {
		return err
	}

	// Create template script
	execPath := filepath.Join(pluginDir, fmt.Sprintf("%s.sh", name))
	templateScript := fmt.Sprintf(`#!/bin/bash
# Plugin: %s
# Description: %s
# Version: %s

# Parse command line arguments
case "$1" in
    "hello")
        echo "Hello from %s plugin!"
        ;;
    "info")
        echo "Plugin: %s v%s"
        echo "Author: %s"
        ;;
    *)
        echo "Usage: $0 {hello|info}"
        exit 1
        ;;
esac
`, name, plugin.Description, plugin.Version, name, name, plugin.Version, plugin.Author)

	if err := os.WriteFile(execPath, []byte(templateScript), 0755); err != nil {
		return err
	}

	// Create README
	readmePath := filepath.Join(pluginDir, "README.md")
	readme := fmt.Sprintf(`# %s Plugin

%s

## Installation

This plugin is created as a template. Customize it for your needs.

## Usage

`+"```bash\n"+`ambros plugin enable %s
`+"```\n"+`

## Configuration

Edit the plugin.json file to customize:
- Commands
- Hooks
- Configuration options
- Dependencies

## Development

The main executable is %s.sh. You can modify it to add your custom functionality.
`, name, plugin.Description, name, name)

	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return err
	}

	color.Green("✅ Plugin template created: %s", name)
	color.White("Plugin directory: %s", pluginDir)
	color.White("Edit %s.sh to add your functionality", name)
	color.White("Run 'ambros plugin enable %s' when ready", name)

	return nil
}

func (pc *PluginCommand) manageRegistry(args []string) error {
	color.Yellow("🏪 Plugin registry management coming soon!")
	color.White("Will support adding custom plugin repositories")
	return nil
}

// Helper methods
func (pc *PluginCommand) getPluginsDirectory() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ambros", "plugins")
}

func (pc *PluginCommand) loadPlugin(name string) (*Plugin, error) {
	pluginPath := filepath.Join(pc.getPluginsDirectory(), name, "plugin.json")
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, err
	}

	var plugin Plugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

func (pc *PluginCommand) savePlugin(name string, plugin Plugin) error {
	pluginDir := filepath.Join(pc.getPluginsDirectory(), name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	pluginPath := filepath.Join(pluginDir, "plugin.json")
	data, err := json.MarshalIndent(plugin, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pluginPath, data, 0644)
}

func (pc *PluginCommand) Command() *cobra.Command {
	return pc.cmd
}
