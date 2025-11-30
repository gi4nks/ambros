package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gi4nks/ambros/v3/internal/models" // New import
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
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
}
