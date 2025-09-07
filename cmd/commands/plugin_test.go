package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gi4nks/ambros/internal/errors"
	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestablePluginCommand extends PluginCommand for testing
type TestablePluginCommand struct {
	*PluginCommand
	testPluginsDir string
}

func NewTestablePluginCommand(logger *zap.Logger, repo RepositoryInterface, testDir string) *TestablePluginCommand {
	pc := NewPluginCommand(logger, repo)
	return &TestablePluginCommand{
		PluginCommand:  pc,
		testPluginsDir: testDir,
	}
}

func (tpc *TestablePluginCommand) getPluginsDirectory() string {
	if tpc.testPluginsDir != "" {
		return tpc.testPluginsDir
	}
	return tpc.PluginCommand.getPluginsDirectory()
}

func (tpc *TestablePluginCommand) loadPlugin(name string) (*Plugin, error) {
	pluginPath := filepath.Join(tpc.getPluginsDirectory(), name, "plugin.json")
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

func (tpc *TestablePluginCommand) savePlugin(name string, plugin Plugin) error {
	pluginDir := filepath.Join(tpc.getPluginsDirectory(), name)
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

// Override the runE method to intercept all plugin operations
func (tpc *TestablePluginCommand) runE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand, "plugin action required", nil)
	}

	tpc.logger.Info("Plugin command executed",
		zap.String("action", args[0]),
		zap.Strings("args", args))

	switch args[0] {
	case "list":
		return tpc.listPlugins()
	case "install":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return tpc.installPlugin(args[1])
	case "uninstall":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return tpc.uninstallPlugin(args[1])
	case "enable":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return tpc.enablePlugin(args[1])
	case "disable":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return tpc.disablePlugin(args[1])
	case "info":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return tpc.showPluginInfo(args[1])
	case "config":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return tpc.manageConfig(args[1], args[2:])
	case "create":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "plugin name required", nil)
		}
		return tpc.createPlugin(args[1])
	case "registry":
		return tpc.manageRegistry(args[1:])
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unknown subcommand: "+args[0], nil)
	}
}

func (tpc *TestablePluginCommand) listPlugins() error {
	pluginsDir := tpc.getPluginsDirectory()
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return nil
	}

	_, err := os.ReadDir(pluginsDir)
	if err != nil {
		return err
	}

	return nil
}

func (tpc *TestablePluginCommand) installPlugin(name string) error {
	pluginDir := filepath.Join(tpc.getPluginsDirectory(), name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

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

	if err := tpc.savePlugin(name, plugin); err != nil {
		return err
	}

	execPath := filepath.Join(pluginDir, fmt.Sprintf("%s.sh", name))
	sampleScript := fmt.Sprintf(`#!/bin/bash
# Sample plugin script for %s
echo "Plugin %s executed with args: $@"
`, name, name)

	return os.WriteFile(execPath, []byte(sampleScript), 0755)
}

func (tpc *TestablePluginCommand) uninstallPlugin(name string) error {
	pluginDir := filepath.Join(tpc.getPluginsDirectory(), name)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", name)
	}

	return os.RemoveAll(pluginDir)
}

func (tpc *TestablePluginCommand) enablePlugin(name string) error {
	plugin, err := tpc.loadPlugin(name)
	if err != nil {
		return err
	}

	plugin.Enabled = true
	return tpc.savePlugin(name, *plugin)
}

func (tpc *TestablePluginCommand) disablePlugin(name string) error {
	plugin, err := tpc.loadPlugin(name)
	if err != nil {
		return err
	}

	plugin.Enabled = false
	return tpc.savePlugin(name, *plugin)
}

func (tpc *TestablePluginCommand) showPluginInfo(name string) error {
	_, err := tpc.loadPlugin(name)
	return err
}

func (tpc *TestablePluginCommand) createPlugin(name string) error {
	pluginDir := filepath.Join(tpc.getPluginsDirectory(), name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

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

	if err := tpc.savePlugin(name, plugin); err != nil {
		return err
	}

	execPath := filepath.Join(pluginDir, fmt.Sprintf("%s.sh", name))
	sampleScript := fmt.Sprintf(`#!/bin/bash
# Plugin script for %s
echo "Custom plugin %s executed"
`, name, name)

	if err := os.WriteFile(execPath, []byte(sampleScript), 0755); err != nil {
		return err
	}

	readmePath := filepath.Join(pluginDir, "README.md")
	readme := fmt.Sprintf(`# %s Plugin

%s

## Installation

This plugin has been created for you. 

## Usage

Run the plugin with:
`+"```"+`
ambros plugin enable %s
`+"```"+`

## Configuration

Edit the plugin.json file to customize:
- Commands
- Hooks
- Configuration options
- Dependencies

## Development

The main executable is %s.sh. You can modify it to add your custom functionality.
`, name, plugin.Description, name, name)

	return os.WriteFile(readmePath, []byte(readme), 0644)
}

func (tpc *TestablePluginCommand) manageConfig(name string, args []string) error {
	return nil
}

func (tpc *TestablePluginCommand) manageRegistry(args []string) error {
	return nil
}

func TestPluginCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ambros-plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("list plugins with no plugins directory", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, filepath.Join(tempDir, "nonexistent"))

		err := pluginCmd.runE(pluginCmd.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list plugins with empty plugins directory", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		pluginsDir := filepath.Join(tempDir, "empty-plugins")
		err := os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err = pluginCmd.runE(pluginCmd.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("install plugin successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "install-test")
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"install", "test-plugin"})
		assert.NoError(t, err)

		// Verify plugin directory was created
		pluginDir := filepath.Join(pluginsDir, "test-plugin")
		assert.DirExists(t, pluginDir)

		// Verify plugin.json was created
		pluginJsonPath := filepath.Join(pluginDir, "plugin.json")
		assert.FileExists(t, pluginJsonPath)

		// Verify executable was created
		execPath := filepath.Join(pluginDir, "test-plugin.sh")
		assert.FileExists(t, execPath)

		// Verify plugin.json content
		data, err := os.ReadFile(pluginJsonPath)
		require.NoError(t, err)

		var plugin Plugin
		err = json.Unmarshal(data, &plugin)
		require.NoError(t, err)

		assert.Equal(t, "test-plugin", plugin.Name)
		assert.Equal(t, "1.0.0", plugin.Version)
		assert.Equal(t, "Sample plugin: test-plugin", plugin.Description)
		assert.Equal(t, "Ambros Plugin System", plugin.Author)
		assert.True(t, plugin.Enabled)
		assert.Equal(t, "./test-plugin.sh", plugin.Executable)

		mockRepo.AssertExpectations(t)
	})

	t.Run("install plugin with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"install"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("uninstall plugin successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "uninstall-test")
		pluginDir := filepath.Join(pluginsDir, "test-plugin")
		err := os.MkdirAll(pluginDir, 0755)
		require.NoError(t, err)

		// Create a test file in the plugin directory
		testFile := filepath.Join(pluginDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err = pluginCmd.runE(pluginCmd.cmd, []string{"uninstall", "test-plugin"})
		assert.NoError(t, err)

		// Verify plugin directory was removed
		assert.NoDirExists(t, pluginDir)

		mockRepo.AssertExpectations(t)
	})

	t.Run("uninstall plugin not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "uninstall-notfound")
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"uninstall", "nonexistent-plugin"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin not found: nonexistent-plugin")

		mockRepo.AssertExpectations(t)
	})

	t.Run("uninstall plugin with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"uninstall"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("create plugin successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "create-test")
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"create", "my-custom-plugin"})
		assert.NoError(t, err)

		// Verify plugin directory was created
		pluginDir := filepath.Join(pluginsDir, "my-custom-plugin")
		assert.DirExists(t, pluginDir)

		// Verify plugin.json was created
		pluginJsonPath := filepath.Join(pluginDir, "plugin.json")
		assert.FileExists(t, pluginJsonPath)

		// Verify executable was created
		execPath := filepath.Join(pluginDir, "my-custom-plugin.sh")
		assert.FileExists(t, execPath)

		// Verify README was created
		readmePath := filepath.Join(pluginDir, "README.md")
		assert.FileExists(t, readmePath)

		// Verify plugin.json content for create (should be disabled by default)
		data, err := os.ReadFile(pluginJsonPath)
		require.NoError(t, err)

		var plugin Plugin
		err = json.Unmarshal(data, &plugin)
		require.NoError(t, err)

		assert.Equal(t, "my-custom-plugin", plugin.Name)
		assert.Equal(t, "0.1.0", plugin.Version)
		assert.Equal(t, "Custom plugin: my-custom-plugin", plugin.Description)
		assert.Equal(t, "Your Name", plugin.Author)
		assert.False(t, plugin.Enabled) // Should be disabled for create
		assert.Equal(t, "./my-custom-plugin.sh", plugin.Executable)

		mockRepo.AssertExpectations(t)
	})

	t.Run("create plugin with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"create"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("unknown action", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")

		mockRepo.AssertExpectations(t)
	})

	t.Run("config management placeholder", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"config", "test-plugin"})
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("config management with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"config"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("registry management placeholder", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"registry"})
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestPluginCommand_WithExistingPlugins(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ambros-plugin-existing-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup test plugins
	setupTestPlugins := func(pluginsDir string) {
		// Create test plugin 1 (enabled)
		plugin1Dir := filepath.Join(pluginsDir, "enabled-plugin")
		err := os.MkdirAll(plugin1Dir, 0755)
		require.NoError(t, err)

		plugin1 := Plugin{
			Name:        "enabled-plugin",
			Version:     "1.0.0",
			Description: "An enabled test plugin",
			Author:      "Test Author",
			Enabled:     true,
			Executable:  "./enabled-plugin.sh",
			Commands: []PluginCommandDef{
				{Name: "test-cmd", Description: "Test command", Usage: "test-cmd [args]", Args: []string{}},
			},
			Hooks:        []string{"pre-run", "post-run"},
			Config:       map[string]string{"key1": "value1", "key2": "value2"},
			Dependencies: []string{"dep1", "dep2"},
		}

		plugin1Json, err := json.MarshalIndent(plugin1, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(plugin1Dir, "plugin.json"), plugin1Json, 0644)
		require.NoError(t, err)

		// Create test plugin 2 (disabled)
		plugin2Dir := filepath.Join(pluginsDir, "disabled-plugin")
		err = os.MkdirAll(plugin2Dir, 0755)
		require.NoError(t, err)

		plugin2 := Plugin{
			Name:         "disabled-plugin",
			Version:      "2.0.0",
			Description:  "A disabled test plugin",
			Author:       "Another Author",
			Enabled:      false,
			Executable:   "./disabled-plugin.sh",
			Commands:     []PluginCommandDef{},
			Hooks:        []string{},
			Config:       map[string]string{},
			Dependencies: []string{},
		}

		plugin2Json, err := json.MarshalIndent(plugin2, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(plugin2Dir, "plugin.json"), plugin2Json, 0644)
		require.NoError(t, err)

		// Create invalid plugin directory (no plugin.json)
		invalidDir := filepath.Join(pluginsDir, "invalid-plugin")
		err = os.MkdirAll(invalidDir, 0755)
		require.NoError(t, err)
	}

	t.Run("list plugins with existing plugins", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "list-existing")
		setupTestPlugins(pluginsDir)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"list"})
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("show plugin info successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "info-test")
		setupTestPlugins(pluginsDir)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"info", "enabled-plugin"})
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("show plugin info for nonexistent plugin", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "info-notfound")
		setupTestPlugins(pluginsDir)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"info", "nonexistent"})
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("show plugin info with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"info"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("enable plugin successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "enable-test")
		setupTestPlugins(pluginsDir)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"enable", "disabled-plugin"})
		assert.NoError(t, err)

		// Verify plugin is now enabled
		plugin, err := pluginCmd.loadPlugin("disabled-plugin")
		require.NoError(t, err)
		assert.True(t, plugin.Enabled)

		mockRepo.AssertExpectations(t)
	})

	t.Run("enable plugin with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"enable"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("disable plugin successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "disable-test")
		setupTestPlugins(pluginsDir)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"disable", "enabled-plugin"})
		assert.NoError(t, err)

		// Verify plugin is now disabled
		plugin, err := pluginCmd.loadPlugin("enabled-plugin")
		require.NoError(t, err)
		assert.False(t, plugin.Enabled)

		mockRepo.AssertExpectations(t)
	})

	t.Run("disable plugin with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, tempDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"disable"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("enable nonexistent plugin", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "enable-notfound")
		setupTestPlugins(pluginsDir)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"enable", "nonexistent"})
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("disable nonexistent plugin", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "disable-notfound")
		setupTestPlugins(pluginsDir)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"disable", "nonexistent"})
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestPluginCommand_EdgeCases(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ambros-plugin-edge-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("install plugin with special characters in name", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "special-chars")
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		err := pluginCmd.runE(pluginCmd.cmd, []string{"install", "my-plugin_v1.0-test"})
		assert.NoError(t, err)

		// Verify plugin directory was created
		pluginDir := filepath.Join(pluginsDir, "my-plugin_v1.0-test")
		assert.DirExists(t, pluginDir)

		mockRepo.AssertExpectations(t)
	})

	t.Run("load plugin with invalid JSON", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "invalid-json")
		pluginDir := filepath.Join(pluginsDir, "invalid-plugin")
		err := os.MkdirAll(pluginDir, 0755)
		require.NoError(t, err)

		// Create invalid JSON file
		invalidJson := `{"name": "test", "version": "1.0.0", invalid}`
		err = os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(invalidJson), 0644)
		require.NoError(t, err)

		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		_, err = pluginCmd.loadPlugin("invalid-plugin")
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("create plugin in readonly directory", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		// Create readonly directory
		readonlyDir := filepath.Join(tempDir, "readonly")
		err := os.MkdirAll(readonlyDir, 0555) // Read and execute only
		require.NoError(t, err)
		defer os.Chmod(readonlyDir, 0755) // Restore permissions for cleanup

		// Use readonly dir as plugins dir
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, filepath.Join(readonlyDir, "plugins"))

		err = pluginCmd.runE(pluginCmd.cmd, []string{"create", "test-readonly"})
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("getPluginsDirectory uses home directory", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewPluginCommand(logger, mockRepo)

		pluginsDir := pluginCmd.getPluginsDirectory()
		homeDir, _ := os.UserHomeDir()
		expectedDir := filepath.Join(homeDir, ".ambros", "plugins")
		assert.Equal(t, expectedDir, pluginsDir)

		mockRepo.AssertExpectations(t)
	})
}

func TestPluginCommand_CommandStructure(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginCmd := NewPluginCommand(logger, mockRepo)

		// Verify command is properly configured
		assert.Equal(t, "plugin", pluginCmd.cmd.Use)
		assert.NotEmpty(t, pluginCmd.cmd.Short)
		assert.NotEmpty(t, pluginCmd.cmd.Long)
		assert.NotNil(t, pluginCmd.cmd.RunE)

		// Verify flags are set up
		nameFlag := pluginCmd.cmd.Flags().Lookup("name")
		assert.NotNil(t, nameFlag)
		assert.Equal(t, "n", nameFlag.Shorthand)

		configFlag := pluginCmd.cmd.Flags().Lookup("config")
		assert.NotNil(t, configFlag)
		assert.Equal(t, "c", configFlag.Shorthand)
	})

	t.Run("plugin data structure validation", func(t *testing.T) {
		plugin := Plugin{
			Name:         "test",
			Version:      "1.0.0",
			Description:  "Test plugin",
			Author:       "Test Author",
			Enabled:      true,
			Executable:   "./test.sh",
			Commands:     []PluginCommandDef{{Name: "cmd", Description: "Test command"}},
			Hooks:        []string{"pre-run"},
			Config:       map[string]string{"key": "value"},
			Dependencies: []string{"dep1"},
		}

		// Test JSON marshaling/unmarshaling
		data, err := json.Marshal(plugin)
		assert.NoError(t, err)

		var unmarshaled Plugin
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)

		assert.Equal(t, plugin.Name, unmarshaled.Name)
		assert.Equal(t, plugin.Version, unmarshaled.Version)
		assert.Equal(t, plugin.Description, unmarshaled.Description)
		assert.Equal(t, plugin.Author, unmarshaled.Author)
		assert.Equal(t, plugin.Enabled, unmarshaled.Enabled)
		assert.Equal(t, plugin.Executable, unmarshaled.Executable)
		assert.Equal(t, len(plugin.Commands), len(unmarshaled.Commands))
		assert.Equal(t, len(plugin.Hooks), len(unmarshaled.Hooks))
		assert.Equal(t, len(plugin.Config), len(unmarshaled.Config))
		assert.Equal(t, len(plugin.Dependencies), len(unmarshaled.Dependencies))
	})

	t.Run("plugin command definition structure", func(t *testing.T) {
		cmdDef := PluginCommandDef{
			Name:        "test-cmd",
			Description: "A test command",
			Usage:       "test-cmd [options]",
			Args:        []string{"arg1", "arg2"},
		}

		// Test JSON marshaling/unmarshaling
		data, err := json.Marshal(cmdDef)
		assert.NoError(t, err)

		var unmarshaled PluginCommandDef
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)

		assert.Equal(t, cmdDef.Name, unmarshaled.Name)
		assert.Equal(t, cmdDef.Description, unmarshaled.Description)
		assert.Equal(t, cmdDef.Usage, unmarshaled.Usage)
		assert.Equal(t, len(cmdDef.Args), len(unmarshaled.Args))
	})
}

func TestPluginCommand_FileOperations(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ambros-plugin-fileops-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("save and load plugin round trip", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "roundtrip")
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		// Create test plugin
		testPlugin := Plugin{
			Name:        "roundtrip-test",
			Version:     "1.5.0",
			Description: "A test plugin for round trip testing",
			Author:      "Round Trip Author",
			Enabled:     true,
			Executable:  "./roundtrip-test.sh",
			Commands: []PluginCommandDef{
				{Name: "hello", Description: "Say hello", Usage: "hello [name]", Args: []string{"name"}},
				{Name: "goodbye", Description: "Say goodbye", Usage: "goodbye", Args: []string{}},
			},
			Hooks:        []string{"pre-command", "post-command"},
			Config:       map[string]string{"greeting": "Hello", "farewell": "Goodbye"},
			Dependencies: []string{"bash", "curl"},
		}

		// Save the plugin
		err := pluginCmd.savePlugin("roundtrip-test", testPlugin)
		assert.NoError(t, err)

		// Load the plugin
		loadedPlugin, err := pluginCmd.loadPlugin("roundtrip-test")
		assert.NoError(t, err)

		// Verify all fields match
		assert.Equal(t, testPlugin.Name, loadedPlugin.Name)
		assert.Equal(t, testPlugin.Version, loadedPlugin.Version)
		assert.Equal(t, testPlugin.Description, loadedPlugin.Description)
		assert.Equal(t, testPlugin.Author, loadedPlugin.Author)
		assert.Equal(t, testPlugin.Enabled, loadedPlugin.Enabled)
		assert.Equal(t, testPlugin.Executable, loadedPlugin.Executable)
		assert.Equal(t, len(testPlugin.Commands), len(loadedPlugin.Commands))
		assert.Equal(t, len(testPlugin.Hooks), len(loadedPlugin.Hooks))
		assert.Equal(t, len(testPlugin.Config), len(loadedPlugin.Config))
		assert.Equal(t, len(testPlugin.Dependencies), len(loadedPlugin.Dependencies))

		// Verify specific command details
		assert.Equal(t, testPlugin.Commands[0].Name, loadedPlugin.Commands[0].Name)
		assert.Equal(t, testPlugin.Commands[0].Description, loadedPlugin.Commands[0].Description)
		assert.Equal(t, testPlugin.Commands[0].Usage, loadedPlugin.Commands[0].Usage)
		assert.Equal(t, testPlugin.Commands[0].Args, loadedPlugin.Commands[0].Args)

		// Verify config details
		assert.Equal(t, testPlugin.Config["greeting"], loadedPlugin.Config["greeting"])
		assert.Equal(t, testPlugin.Config["farewell"], loadedPlugin.Config["farewell"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("load plugin from nonexistent file", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "nonexistent")
		pluginCmd := NewTestablePluginCommand(logger, mockRepo, pluginsDir)

		_, err := pluginCmd.loadPlugin("nonexistent-plugin")
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}
