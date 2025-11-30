package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

// PluginCommand represents the plugin management command
type PluginCommand struct {
	*BaseCommand
	pluginName string
	configFile string
	baseDir_   string // For testing
}

// RegistryConfig holds the list of plugin registries
type RegistryConfig struct {
	Registries []Registry `json:"registries"`
}

// Registry represents a single plugin registry
type Registry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
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
  list                      List installed plugins
  install <source-path>     Install a plugin from a local path or from a registry
  create <name>    Create a new plugin from a template
  uninstall <name>          Remove a plugin
  enable <name>             Enable a disabled plugin
  disable <name>            Disable a plugin
  info <name>               Show detailed plugin information
  config <name>             Manage plugin configuration
  registry                  Manage plugin registries

Examples:
  ambros plugin list
  ambros plugin install /path/to/my/plugin # Install from a local directory
  ambros plugin install my-plugin # Install from a registry
  ambros plugin create my-custom-plugin
  ambros plugin info slack-notifications
  ambros plugin config git-hooks --set webhook.url=https://example.com`,
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
			return errors.NewError(errors.ErrInvalidCommand, "source path or plugin name required", nil)
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

func (pc *PluginCommand) uninstallPlugin(name string) error {
	color.Yellow("🗑️  Uninstalling plugin: %s", name)

	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return errors.NewError(errors.ErrCommandNotFound, fmt.Sprintf("plugin not found: %s", name), nil)
	}

	if err := pc.ensureDirIsSafe(pluginDir); err != nil {
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

func (pc *PluginCommand) manageConfig(pluginName string, args []string) error {
	if len(args) == 0 {
		return pc.listConfig(pluginName)
	}

	subcommand := args[0]
	switch subcommand {
	case "get":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "config key required for get", nil)
		}
		return pc.getConfig(pluginName, args[1])
	case "set":
		if len(args) < 3 {
			return errors.NewError(errors.ErrInvalidCommand, "config key and value required for set", nil)
		}
		return pc.setConfig(pluginName, args[1], args[2])
	case "list":
		return pc.listConfig(pluginName)
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unknown config subcommand: "+subcommand, nil)
	}
}

func (pc *PluginCommand) listConfig(pluginName string) error {
	plugin, err := pc.loadPlugin(pluginName)
	if err != nil {
		return err
	}

	color.Cyan("⚙️  Configuration for %s:", pluginName)
	if len(plugin.Config) == 0 {
		color.Yellow("No configuration set.")
		return nil
	}

	for key, value := range plugin.Config {
		fmt.Printf("  %s: %s\n", key, value)
	}
	return nil
}

func (pc *PluginCommand) getConfig(pluginName, key string) error {
	plugin, err := pc.loadPlugin(pluginName)
	if err != nil {
		return err
	}

	value, ok := plugin.Config[key]
	if !ok {
		return errors.NewError(errors.ErrNotFound, fmt.Sprintf("config key '%s' not found for plugin '%s'", key, pluginName), nil)
	}

	fmt.Println(value)
	return nil
}

func (pc *PluginCommand) setConfig(pluginName, key, value string) error {
	plugin, err := pc.loadPlugin(pluginName)
	if err != nil {
		return err
	}

	if plugin.Config == nil {
		plugin.Config = make(map[string]string)
	}
	plugin.Config[key] = value

	if err := pc.savePlugin(pluginName, *plugin); err != nil {
		return err
	}

	color.Green("✅ Config set: %s = %s", key, value)
	return nil
}

func (pc *PluginCommand) createPlugin(name string) error {
	if !pc.isValidPluginName(name) {
		return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("invalid plugin name: %s (allowed: letters, numbers, dot, dash, underscore)", name), nil)
	}
	color.Cyan("🛠️  Creating plugin template: %s", name)

	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return err
	}

	if err := pc.ensureDirIsSafe(pluginDir); err != nil {
		return err
	}

	if err := os.MkdirAll(pluginDir, 0750); err != nil {
		return err
	}

	plugin := models.Plugin{
		Name:         name,
		Version:      "0.1.0",
		Description:  fmt.Sprintf("Custom plugin: %s", name),
		Author:       "Your Name",
		Enabled:      false,
		Type:         models.PluginTypeShell, // Default to shell type for new plugins
		Executable:   fmt.Sprintf("./%s.sh", name),
		Commands:     []models.PluginCommandDef{}, // Use models.PluginCommandDef
		Hooks:        []string{},
		Config:       make(map[string]string),
		Dependencies: []string{},
	}

	// Create the executable script FIRST, before savePlugin validates it
	execPath := filepath.Join(pluginDir, fmt.Sprintf("%s.sh", name))
	templateScript := fmt.Sprintf(`#!/bin/bash
# Plugin: %s
# Description: %s
# Version: %s

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
`, name, plugin.Description, plugin.Version, name, name, plugin.Version, name)

	if err := os.WriteFile(execPath, []byte(templateScript), 0750); err != nil {
		return err
	}

	// Now save plugin.json (which validates the executable exists)
	if err := pc.savePlugin(name, plugin); err != nil {
		return err
	}

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

func (pc *PluginCommand) manageRegistry(args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand, "subcommand required: list, add, remove", nil)
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		return pc.listRegistries()
	case "add":
		if len(args) < 3 {
			return errors.NewError(errors.ErrInvalidCommand, "name and url required for add", nil)
		}
		return pc.addRegistry(args[1], args[2])
	case "remove":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "name required for remove", nil)
		}
		return pc.removeRegistry(args[1])
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unknown registry subcommand: "+subcommand, nil)
	}
}

func (pc *PluginCommand) getRegistryConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ambros", "registries.json"), nil
}

func (pc *PluginCommand) readRegistryConfig() (*RegistryConfig, error) {
	configPath, err := pc.getRegistryConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &RegistryConfig{Registries: []Registry{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config RegistryConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (pc *PluginCommand) writeRegistryConfig(config *RegistryConfig) error {
	configPath, err := pc.getRegistryConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0640)
}

func (pc *PluginCommand) listRegistries() error {
	config, err := pc.readRegistryConfig()
	if err != nil {
		return err
	}

	color.Cyan("🔌 Plugin Registries")
	color.Cyan("═══════════════════")

	if len(config.Registries) == 0 {
		color.Yellow("No registries configured.")
		color.White("Add a registry with: ambros plugin registry add <name> <url>")
		return nil
	}

	for _, registry := range config.Registries {
		color.White("📦 %s - %s", registry.Name, registry.URL)
	}

	return nil
}

func (pc *PluginCommand) addRegistry(name, url string) error {
	config, err := pc.readRegistryConfig()
	if err != nil {
		return err
	}

	for _, registry := range config.Registries {
		if registry.Name == name {
			return errors.NewError(errors.ErrCommandExists, fmt.Sprintf("registry with name '%s' already exists", name), nil)
		}
	}

	newRegistry := Registry{Name: name, URL: url}
	config.Registries = append(config.Registries, newRegistry)

	if err := pc.writeRegistryConfig(config); err != nil {
		return err
	}

	color.Green("✅ Registry '%s' added.", name)
	return nil
}

func (pc *PluginCommand) removeRegistry(name string) error {
	config, err := pc.readRegistryConfig()
	if err != nil {
		return err
	}

	var found bool
	var updatedRegistries []Registry
	for _, registry := range config.Registries {
		if registry.Name != name {
			updatedRegistries = append(updatedRegistries, registry)
		} else {
			found = true
		}
	}

	if !found {
		return errors.NewError(errors.ErrNotFound, fmt.Sprintf("registry with name '%s' not found", name), nil)
	}

	config.Registries = updatedRegistries
	if err := pc.writeRegistryConfig(config); err != nil {
		return err
	}

	color.Green("✅ Registry '%s' removed.", name)
	return nil
}

func (pc *PluginCommand) installPlugin(source string) error {
	// If source is not a local path, assume it's a plugin name from a registry
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return pc.installPluginFromRegistry(source)
	}
	return pc.installPluginFromLocalPath(source)
}

func (pc *PluginCommand) installPluginFromLocalPath(sourcePath string) error {
	srcInfo, err := os.Stat(sourcePath)
	if err != nil {
		return errors.NewError(errors.ErrNotFound, fmt.Sprintf("source path '%s' not found or accessible", sourcePath), err)
	}
	if !srcInfo.IsDir() {
		return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("source path '%s' is not a directory", sourcePath), nil)
	}

	manifestPath := filepath.Join(sourcePath, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.NewError(errors.ErrConfigInvalid, fmt.Sprintf("could not read plugin.json from '%s'", sourcePath), err)
	}
	var plugin models.Plugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return errors.NewError(errors.ErrConfigInvalid, fmt.Sprintf("invalid plugin.json format in '%s'", sourcePath), err)
	}

	if !pc.isValidPluginName(plugin.Name) {
		return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("invalid plugin name in manifest: %s", plugin.Name), nil)
	}

	if err := pc.checkAndInstallDependencies(&plugin); err != nil {
		return err
	}

	destPluginDir, err := pc.pluginDirPath(plugin.Name)
	if err != nil {
		return err
	}

	if err := pc.ensureDirIsSafe(destPluginDir); err != nil {
		return err
	}

	if _, err := os.Stat(destPluginDir); err == nil {
		return errors.NewError(errors.ErrCommandExists, fmt.Sprintf("plugin '%s' already exists at '%s'", plugin.Name, destPluginDir), nil)
	}

	if err := os.MkdirAll(destPluginDir, 0750); err != nil {
		return errors.NewError(errors.ErrInternalServer, fmt.Sprintf("failed to create destination directory '%s'", destPluginDir), err)
	}

	if err := pc.copyDir(sourcePath, destPluginDir); err != nil {
		_ = os.RemoveAll(destPluginDir)
		return errors.NewError(errors.ErrInternalServer, fmt.Sprintf("failed to copy plugin files from '%s' to '%s'", sourcePath, destPluginDir), err)
	}

	plugin.Enabled = true
	if err := pc.savePlugin(plugin.Name, plugin); err != nil {
		_ = os.RemoveAll(destPluginDir)
		return errors.NewError(errors.ErrInternalServer, fmt.Sprintf("failed to enable plugin '%s' after installation", plugin.Name), err)
	}

	color.Green("✅ Plugin '%s' installed successfully from '%s'", plugin.Name, sourcePath)
	color.White("Plugin directory: %s", destPluginDir)

	return nil
}

func (pc *PluginCommand) installPluginFromRegistry(pluginName string) error {
	config, err := pc.readRegistryConfig()
	if err != nil {
		return err
	}

	if len(config.Registries) == 0 {
		return errors.NewError(errors.ErrNotFound, "no registries configured. Add a registry with: ambros plugin registry add <name> <url>", nil)
	}

	for _, registry := range config.Registries {
		pluginURL := fmt.Sprintf("%s/%s/plugin.json", registry.URL, pluginName)
		color.White("Trying to fetch from %s", pluginURL)

		resp, err := http.Get(pluginURL)
		if err != nil {
			pc.logger.Debug("Failed to fetch plugin from registry", zap.String("registry", registry.Name), zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			pc.logger.Debug("Plugin not found in registry", zap.String("registry", registry.Name), zap.Int("status", resp.StatusCode))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			pc.logger.Debug("Failed to read plugin body from registry", zap.String("registry", registry.Name), zap.Error(err))
			continue
		}

		var plugin models.Plugin
		if err := json.Unmarshal(body, &plugin); err != nil {
			pc.logger.Debug("Failed to unmarshal plugin from registry", zap.String("registry", registry.Name), zap.Error(err))
			continue
		}

		if err := pc.checkAndInstallDependencies(&plugin); err != nil {
			return err
		}

		// Create a temporary directory to download the plugin files
		tmpDir, err := os.MkdirTemp("", "ambros-plugin-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		// Download all plugin files
		for _, command := range plugin.Commands {
			executablePath := command.Usage
			if !strings.HasPrefix(executablePath, "http") {
				executablePath = fmt.Sprintf("%s/%s/%s", registry.URL, pluginName, executablePath)
			}
			err = pc.downloadFile(filepath.Join(tmpDir, command.Usage), executablePath)
			if err != nil {
				return err
			}
		}

		// download the plugin executable
		executablePath := plugin.Executable
		if !strings.HasPrefix(executablePath, "http") {
			executablePath = fmt.Sprintf("%s/%s/%s", registry.URL, pluginName, executablePath)
		}
		err = pc.downloadFile(filepath.Join(tmpDir, plugin.Executable), executablePath)
		if err != nil {
			return err
		}

		// copy the plugin.json file
		err = os.WriteFile(filepath.Join(tmpDir, "plugin.json"), body, 0644)
		if err != nil {
			return err
		}

		return pc.installPluginFromLocalPath(tmpDir)
	}

	return errors.NewError(errors.ErrNotFound, fmt.Sprintf("plugin '%s' not found in any configured registry", pluginName), nil)
}

func (pc *PluginCommand) downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// copyDir recursively copies a directory from src to dest.
// Permissions are copied but mode bits like setuid/setgid are not preserved for safety.
func (pc *PluginCommand) copyDir(src string, dest string) error {
	src = filepath.Clean(src)
	dest = filepath.Clean(dest)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode().Perm())
		} else {
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			destFile, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			return err
		}
	})
}

// Helper methods
func (pc *PluginCommand) isValidPluginName(name string) bool {
	var validName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	return validName.MatchString(name)
}

func (pc *PluginCommand) getPluginsDirectory() string {
	if pc.baseDir_ != "" {
		return pc.baseDir_
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ambros", "plugins")
}

func (pc *PluginCommand) loadPlugin(name string) (*models.Plugin, error) {
	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return nil, err
	}

	if err := pc.ensureDirIsSafe(pluginDir); err != nil {
		return nil, err
	}

	pluginPath := filepath.Join(pluginDir, "plugin.json")
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, err
	}

	var plugin models.Plugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

func (pc *PluginCommand) savePlugin(name string, plugin models.Plugin) error {
	pluginDir, err := pc.pluginDirPath(name)
	if err != nil {
		return err
	}

	if err := pc.ensureDirIsSafe(pluginDir); err != nil {
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

	if err := os.WriteFile(pluginPath, data, 0640); err != nil {
		return err
	}

	if plugin.Executable != "" {
		if err := pc.validateExecutablePath(name, plugin.Executable); err != nil {
			return err
		}
	}

	return nil
}

func (pc *PluginCommand) validateExecutablePath(pluginName, execPath string) error {
	if execPath == "" {
		return nil
	}

	if filepath.IsAbs(execPath) {
		return errors.NewError(errors.ErrInvalidCommand, "executable path must be relative to plugin directory", nil)
	}

	cleaned := filepath.Clean(execPath)
	if cleaned == ".." || strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "..") {
		return errors.NewError(errors.ErrInvalidCommand, "executable path contains traversal", nil)
	}

	pluginDir, err := pc.pluginDirPath(pluginName)
	if err != nil {
		return err
	}

	full := filepath.Join(pluginDir, cleaned)

	rel, err := filepath.Rel(pluginDir, full)
	if err != nil {
		return err
	}
	if rel == ".." || rel == "." || strings.HasPrefix(rel, "..") {
		return errors.NewError(errors.ErrInvalidCommand, "executable resolves outside plugin directory", nil)
	}

	fi, err := os.Lstat(full)
	if err != nil {
		return errors.NewError(errors.ErrNotFound, "executable not found", err)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		return errors.NewError(errors.ErrInvalidCommand, "executable must not be a symlink", nil)
	}

	if fi.Mode().Perm()&0100 == 0 {
		return errors.NewError(errors.ErrInvalidCommand, "executable is not executable", nil)
	}

	return nil
}

func (pc *PluginCommand) pluginDirPath(name string) (string, error) {
	base := pc.getPluginsDirectory()
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}

	joined := filepath.Join(absBase, name)
	cleaned := filepath.Clean(joined)

	rel, err := filepath.Rel(absBase, cleaned)
	if err != nil {
		return "", err
	}
	if rel == ".." || rel == "." || rel == "" || strings.HasPrefix(rel, "..") {
		return "", errors.NewError(errors.ErrInvalidCommand, "plugin path resolves outside plugins directory", nil)
	}

	return cleaned, nil
}

func (pc *PluginCommand) ensureDirIsSafe(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("plugin directory must not be a symlink: %s", path), nil)
	}

	return nil
}

func (pc *PluginCommand) checkAndInstallDependencies(plugin *models.Plugin) error {
	for _, dep := range plugin.Dependencies {
		color.White("Checking dependency: %s", dep)
		// Check if the dependency is already installed
		if _, err := pc.loadPlugin(dep); err == nil {
			color.Green("  - Dependency '%s' already installed.", dep)
			continue
		}

		color.Yellow("  - Dependency '%s' not found, trying to install...", dep)
		if err := pc.installPluginFromRegistry(dep); err != nil {
			return errors.NewError(errors.ErrInternalServer, fmt.Sprintf("failed to install dependency '%s'", dep), err)
		}
		color.Green("  - Dependency '%s' installed successfully.", dep)
	}
	return nil
}

func (pc *PluginCommand) Command() *cobra.Command {
	return pc.cmd
}
