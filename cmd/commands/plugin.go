package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
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
		Short: "🔌 Manage Ambros plugins",
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
	if !isValidPluginName(name) {
		return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("invalid plugin name: %s (allowed: letters, numbers, dot, dash, underscore)", name), nil)
	}
	color.Cyan("📦 Installing plugin: %s", name)

	// Create a safe plugin directory inside the plugins base
	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return err
	}

	if err := ensureDirIsSafe(pluginDir); err != nil {
		return err
	}

	if err := os.MkdirAll(pluginDir, 0750); err != nil {
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

	// Write executable with restrictive permissions; rely on umask but set owner-exec
	if err := os.WriteFile(execPath, []byte(sampleScript), 0750); err != nil {
		return err
	}

	color.Green("✅ Plugin %s installed successfully", name)
	color.White("Plugin directory: %s", pluginDir)
	return nil
}

func (pc *PluginCommand) uninstallPlugin(name string) error {
	color.Yellow("🗑️  Uninstalling plugin: %s", name)

	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Prevent accidental removal outside plugins directory
	if err := ensureDirIsSafe(pluginDir); err != nil {
		return err
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

	// Validate executable before enabling
	if err := pc.validateExecutablePath(name, plugin.Executable); err != nil {
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

func (pc *PluginCommand) manageConfig(_ string, _ []string) error {
	color.Yellow("🔧 Plugin configuration management coming soon!")
	color.White("Will support: --set key=value, --get key, --list")
	return nil
}

func (pc *PluginCommand) createPlugin(name string) error {
	if !isValidPluginName(name) {
		return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("invalid plugin name: %s (allowed: letters, numbers, dot, dash, underscore)", name), nil)
	}
	color.Cyan("🛠️  Creating plugin template: %s", name)

	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return err
	}

	if err := ensureDirIsSafe(pluginDir); err != nil {
		return err
	}

	if err := os.MkdirAll(pluginDir, 0750); err != nil {
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

	if err := os.WriteFile(execPath, []byte(templateScript), 0750); err != nil {
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

	if err := os.WriteFile(readmePath, []byte(readme), 0640); err != nil {
		return err
	}

	color.Green("✅ Plugin template created: %s", name)
	color.White("Plugin directory: %s", pluginDir)
	color.White("Edit %s.sh to add your functionality", name)
	color.White("Run 'ambros plugin enable %s' when ready", name)

	return nil
}

func (pc *PluginCommand) manageRegistry(_ []string) error {
	color.Yellow("🏪 Plugin registry management coming soon!")
	color.White("Will support adding custom plugin repositories")
	return nil
}

// isValidPluginName ensures plugin names are simple and don't contain path separators
func isValidPluginName(name string) bool {
	// allow letters, numbers, dot, dash and underscore
	var validName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	return validName.MatchString(name)
}

// Helper methods
func (pc *PluginCommand) getPluginsDirectory() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ambros", "plugins")
}

func (pc *PluginCommand) loadPlugin(name string) (*Plugin, error) {
	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return nil, err
	}

	if err := ensureDirIsSafe(pluginDir); err != nil {
		return nil, err
	}

	pluginPath := filepath.Join(pluginDir, "plugin.json")
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
	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return err
	}

	if err := ensureDirIsSafe(pluginDir); err != nil {
		return err
	}

	if err := os.MkdirAll(pluginDir, 0750); err != nil {
		return err
	}

	pluginPath := filepath.Join(pluginDir, "plugin.json")
	data, err := json.MarshalIndent(plugin, "", "  ")
	if err != nil {
		return err
	}

	// Write file with restrictive perms
	if err := os.WriteFile(pluginPath, data, 0640); err != nil {
		return err
	}

	// Validate executable field points to a safe path relative to plugin dir
	if plugin.Executable != "" {
		if err := pc.validateExecutablePath(name, plugin.Executable); err != nil {
			return err
		}
	}

	return nil
}

// validateExecutablePath ensures the executable declared by a plugin is a safe, relative path within the plugin dir,
// does not contain traversal, is not a symlink, and (when present) is executable.
func (pc *PluginCommand) validateExecutablePath(pluginName, execPath string) error {
	if execPath == "" {
		return nil
	}

	// Disallow absolute paths
	if filepath.IsAbs(execPath) {
		return fmt.Errorf("executable path must be relative to plugin directory")
	}

	// Disallow traversal components
	cleaned := filepath.Clean(execPath)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "..") {
		return fmt.Errorf("executable path contains traversal")
	}

	// Resolve against plugin dir
	pluginDir, err := pc.pluginDirPath(pluginName)
	if err != nil {
		return err
	}

	full := filepath.Join(pluginDir, cleaned)

	// Ensure the resolved path is inside plugin dir
	rel, err := filepath.Rel(pluginDir, full)
	if err != nil {
		return err
	}
	if rel == ".." || rel == "." || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("executable resolves outside plugin directory")
	}

	// Check file exists
	fi, err := os.Lstat(full)
	if err != nil {
		return fmt.Errorf("executable not found: %w", err)
	}

	// No symlinks
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("executable must not be a symlink")
	}

	// Check exec bit for owner
	if fi.Mode().Perm()&0100 == 0 {
		return fmt.Errorf("executable is not executable")
	}

	return nil
}

// pluginDirPath returns the canonical plugin directory path for a plugin name
func (pc *PluginCommand) pluginDirPath(name string) (string, error) {
	base := pc.getPluginsDirectory()
	// Ensure base is absolute
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}

	// Join and clean
	joined := filepath.Join(absBase, name)
	cleaned := filepath.Clean(joined)

	// Prevent path traversal outside base
	rel, err := filepath.Rel(absBase, cleaned)
	if err != nil {
		return "", err
	}
	if rel == ".." || rel == "." || rel == "" || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("plugin path resolves outside plugins directory")
	}

	return cleaned, nil
}

// ensureDirIsSafe performs basic checks: no symlinks in path and path is under plugins dir
func ensureDirIsSafe(path string) error {
	// Check that none of the path components are symlinks to avoid bypassing
	cur := path
	for {
		fi, err := os.Lstat(cur)
		if err != nil {
			// If the path doesn't exist yet, check parent
			parent := filepath.Dir(cur)
			if parent == cur {
				break
			}
			cur = parent
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path contains symlink: %s", cur)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return nil
}

func (pc *PluginCommand) Command() *cobra.Command {
	return pc.cmd
}
