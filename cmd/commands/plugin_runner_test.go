package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
)

func TestPluginCommandExecution(t *testing.T) {
	// 1. Setup a temporary directory for plugins
	tmpDir, err := os.MkdirTemp("", "ambros-plugins-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// --- Create a test plugin ---
	pluginName := "test-plugin"
	pluginDir := filepath.Join(tmpDir, pluginName)
	err = os.Mkdir(pluginDir, 0755)
	assert.NoError(t, err)

	// 2. Create the plugin manifest (plugin.json)
	manifest := `
{
	"name": "test-plugin",
	"version": "1.0.0",
	"description": "A test plugin",
	"author": "Test",
	"enabled": true,
	"type": "shell",
	"executable": "./test-script.sh",
	"commands": [
		{
			"name": "echo",
			"description": "Echoes a message"
		}
	]
}
`
	err = os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0644)
	assert.NoError(t, err)

	// 3. Create the executable script
	scriptContent := `#!/bin/sh
echo "Plugin executed with command: $1 and args: $2"
echo "Env var AMBROS_PLUGIN_NAME is: $AMBROS_PLUGIN_NAME"
`
	scriptPath := filepath.Join(pluginDir, "test-script.sh")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	assert.NoError(t, err)

	// --- Test the plugin registration and execution ---
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	// 4. Create a test root command and register plugins
	rootCmd := &cobra.Command{Use: "ambros"}

	pc := NewPluginCommand(logger, mockRepo)
	pc.baseDir_ = tmpDir

	RegisterPluginCommands(rootCmd, pc)

	// 5. Find the registered plugin command
	pluginCmd := findCommand(rootCmd, pluginName)
	assert.NotNil(t, pluginCmd, "Plugin command should be registered")
	assert.Equal(t, pluginName, pluginCmd.Use, "Plugin command should have the correct name")

	echoCmd := findCommand(pluginCmd, "echo")
	assert.NotNil(t, echoCmd, "Sub-command 'echo' should be registered under the plugin")

	// 6. Execute the command and capture output
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out) // Capture stderr as well

	rootCmd.SetArgs([]string{pluginName, "echo", "hello-world"})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// 7. Verify the output
	output := out.String()
	assert.Contains(t, output, "Plugin executed with command: echo and args: hello-world")
	assert.Contains(t, output, "Env var AMBROS_PLUGIN_NAME is: test-plugin")
}

// findCommand is a helper to find a subcommand by name.
func findCommand(root *cobra.Command, use string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Use == use {
			return cmd
		}
	}
	return nil
}
