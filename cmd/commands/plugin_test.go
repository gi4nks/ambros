package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gi4nks/ambros/v3/internal/models" // New import
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPluginCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ambros-plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("list plugins with no plugins directory", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = filepath.Join(tempDir, "nonexistent")

		err := pc.runE(pc.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list plugins with empty plugins directory", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "empty-plugins")
		err := os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = pluginsDir

		err = pc.runE(pc.cmd, []string{"list"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create plugin successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "create-test")
		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = pluginsDir

		err := pc.runE(pc.cmd, []string{"create", "test-plugin"})
		assert.NoError(t, err)

		pluginDir := filepath.Join(pluginsDir, "test-plugin")
		assert.DirExists(t, pluginDir)

		pluginJsonPath := filepath.Join(pluginDir, "plugin.json")
		assert.FileExists(t, pluginJsonPath)

		// createPlugin creates <name>.sh in plugin root, not scripts/run.sh
		execPath := filepath.Join(pluginDir, "test-plugin.sh")
		assert.FileExists(t, execPath)

		data, err := os.ReadFile(pluginJsonPath)
		require.NoError(t, err)

		var plugin models.Plugin
		err = json.Unmarshal(data, &plugin)
		require.NoError(t, err)

		assert.Equal(t, "test-plugin", plugin.Name)
		// createPlugin sets Enabled=false by default
		assert.False(t, plugin.Enabled)

		mockRepo.AssertExpectations(t)
	})

	t.Run("install plugin successfully from local path", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "install-from-path-test")
		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = pluginsDir

		sourcePluginDir, err := os.MkdirTemp("", "source-plugin")
		require.NoError(t, err)
		defer os.RemoveAll(sourcePluginDir)

		manifest := `
{
	"name": "my-local-plugin",
	"version": "1.0.0",
	"description": "A local test plugin",
	"author": "Local Test",
	"enabled": false,
	"executable": "./my-script.sh",
	"commands": [
		{
			"name": "hello",
			"description": "Says hello"
		}
	]
}
`
		err = os.WriteFile(filepath.Join(sourcePluginDir, "plugin.json"), []byte(manifest), 0644)
		require.NoError(t, err)

		execContent := `#!/bin/sh
echo "Hello from local plugin"`
		err = os.WriteFile(filepath.Join(sourcePluginDir, "my-script.sh"), []byte(execContent), 0755)
		require.NoError(t, err)

		err = pc.runE(pc.cmd, []string{"install", sourcePluginDir})
		assert.NoError(t, err)

		installedPluginDir := filepath.Join(pluginsDir, "my-local-plugin")
		assert.DirExists(t, installedPluginDir)

		installedManifestPath := filepath.Join(installedPluginDir, "plugin.json")
		assert.FileExists(t, installedManifestPath)
		data, err := os.ReadFile(installedManifestPath)
		require.NoError(t, err)
		var installedPlugin models.Plugin
		err = json.Unmarshal(data, &installedPlugin)
		require.NoError(t, err)
		assert.True(t, installedPlugin.Enabled, "Installed plugin should be enabled")
		assert.Equal(t, "my-local-plugin", installedPlugin.Name)

		assert.FileExists(t, filepath.Join(installedPluginDir, "my-script.sh"))

		mockRepo.AssertExpectations(t)
	})

	t.Run("install plugin with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = tempDir

		err := pc.runE(pc.cmd, []string{"install"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source path or plugin name required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("run plugin with insufficient arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = tempDir

		// Missing command name
		err := pc.runE(pc.cmd, []string{"run", "my-plugin"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin name and command required")

		mockRepo.AssertExpectations(t)
	})

	t.Run("run plugin that does not exist", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "run-nonexistent-test")
		err := os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = pluginsDir

		err = pc.runE(pc.cmd, []string{"run", "nonexistent-plugin", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("run plugin that is disabled", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "run-disabled-test")
		err := os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		// Create a disabled plugin
		pluginDir := filepath.Join(pluginsDir, "disabled-plugin")
		err = os.MkdirAll(pluginDir, 0755)
		require.NoError(t, err)

		manifest := `{
	"name": "disabled-plugin",
	"version": "1.0.0",
	"description": "A disabled plugin",
	"enabled": false,
	"executable": "./run.sh"
}`
		err = os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
		require.NoError(t, err)

		// Create executable
		execContent := `#!/bin/sh
echo "Hello"`
		err = os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte(execContent), 0755)
		require.NoError(t, err)

		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = pluginsDir

		err = pc.runE(pc.cmd, []string{"run", "disabled-plugin", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "disabled")

		mockRepo.AssertExpectations(t)
	})

	t.Run("run plugin successfully and store result", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "run-success-test")
		err := os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		// Create a working plugin
		pluginDir := filepath.Join(pluginsDir, "working-plugin")
		err = os.MkdirAll(pluginDir, 0755)
		require.NoError(t, err)

		manifest := `{
	"name": "working-plugin",
	"version": "1.0.0",
	"description": "A working plugin",
	"enabled": true,
	"executable": "./run.sh",
	"commands": [
		{"name": "hello", "description": "Says hello"}
	]
}`
		err = os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
		require.NoError(t, err)

		// Create executable that outputs based on command
		execContent := `#!/bin/sh
case "$1" in
    "hello")
        echo "Hello from working-plugin!"
        exit 0
        ;;
    *)
        echo "Unknown command: $1"
        exit 1
        ;;
esac`
		err = os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte(execContent), 0755)
		require.NoError(t, err)

		// Mock the repository Put call
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "working-plugin" &&
				cmd.Category == "plugin" &&
				cmd.Status == true &&
				len(cmd.Tags) == 2 &&
				cmd.Tags[0] == "plugin" &&
				cmd.Tags[1] == "working-plugin"
		})).Return(nil)

		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = pluginsDir

		err = pc.runE(pc.cmd, []string{"run", "working-plugin", "hello"})
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("run plugin with failing command stores failure", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		pluginsDir := filepath.Join(tempDir, "run-fail-test")
		err := os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		// Create a plugin that will fail
		pluginDir := filepath.Join(pluginsDir, "failing-plugin")
		err = os.MkdirAll(pluginDir, 0755)
		require.NoError(t, err)

		manifest := `{
	"name": "failing-plugin",
	"version": "1.0.0",
	"description": "A plugin that fails",
	"enabled": true,
	"executable": "./run.sh"
}`
		err = os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
		require.NoError(t, err)

		// Create executable that always fails
		execContent := `#!/bin/sh
echo "This will fail"
exit 1`
		err = os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte(execContent), 0755)
		require.NoError(t, err)

		// Mock the repository Put call for the failed command
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "failing-plugin" &&
				cmd.Status == false &&
				cmd.Category == "plugin"
		})).Return(nil)

		pc := NewPluginCommand(logger, mockRepo)
		pc.baseDir_ = pluginsDir

		err = pc.runE(pc.cmd, []string{"run", "failing-plugin", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed with exit code")

		mockRepo.AssertExpectations(t)
	})
}
