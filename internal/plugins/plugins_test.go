package plugins

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gi4nks/ambros/v3/internal/models"
)

func TestNewInternalPluginRegistry(t *testing.T) {
	registry := NewInternalPluginRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.plugins)
	assert.Empty(t, registry.GetAllPlugins())
}

func TestInternalPluginRegistry_RegisterPlugin(t *testing.T) {
	registry := NewInternalPluginRegistry()
	plugin := NewExamplePlugin()

	err := registry.RegisterPlugin(plugin)
	require.NoError(t, err)

	// Verify plugin is registered
	retrieved, found := registry.GetPlugin("example-go-plugin")
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "example-go-plugin", retrieved.GetManifest().Name)
}

func TestInternalPluginRegistry_RegisterPlugin_Duplicate(t *testing.T) {
	registry := NewInternalPluginRegistry()
	plugin := NewExamplePlugin()

	// First registration should succeed
	err := registry.RegisterPlugin(plugin)
	require.NoError(t, err)

	// Second registration should fail
	err = registry.RegisterPlugin(plugin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestInternalPluginRegistry_GetPlugin_NotFound(t *testing.T) {
	registry := NewInternalPluginRegistry()

	retrieved, found := registry.GetPlugin("non-existent-plugin")
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestInternalPluginRegistry_GetAllPlugins(t *testing.T) {
	registry := NewInternalPluginRegistry()

	// Register example plugin
	plugin := NewExamplePlugin()
	err := registry.RegisterPlugin(plugin)
	require.NoError(t, err)

	// Get all plugins
	allPlugins := registry.GetAllPlugins()
	assert.Len(t, allPlugins, 1)
	assert.Equal(t, "example-go-plugin", allPlugins[0].GetManifest().Name)
}

func TestGlobalRegistry(t *testing.T) {
	// Get global registry
	registry := GetGlobalRegistry()
	assert.NotNil(t, registry)

	// Verify it's the same instance
	registry2 := GetGlobalRegistry()
	assert.Same(t, registry, registry2)
}

func TestExamplePlugin_GetManifest(t *testing.T) {
	plugin := NewExamplePlugin()
	manifest := plugin.GetManifest()

	assert.Equal(t, "example-go-plugin", manifest.Name)
	assert.Equal(t, "1.0.0", manifest.Version)
	assert.Equal(t, models.PluginTypeGoInternal, manifest.Type)
	assert.True(t, manifest.Enabled)
	assert.Len(t, manifest.Commands, 2)
	assert.Contains(t, manifest.Hooks, "pre-run")
	assert.Contains(t, manifest.Hooks, "post-run")
}

func TestExamplePlugin_Init(t *testing.T) {
	plugin := NewExamplePlugin()

	// Init without CoreAPI (nil) should not panic
	err := plugin.Init(nil)
	assert.NoError(t, err)
}

func TestExamplePlugin_Run_Hello(t *testing.T) {
	plugin := NewExamplePlugin()
	ctx := context.Background()

	var stdout, stderr bytes.Buffer

	// Test hello command without args
	err := plugin.Run(ctx, "hello", []string{}, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Hello, World!")

	// Reset buffer
	stdout.Reset()

	// Test hello command with args
	err = plugin.Run(ctx, "hello", []string{"Alice"}, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Hello, Alice!")
}

func TestExamplePlugin_Run_Info(t *testing.T) {
	plugin := NewExamplePlugin()
	ctx := context.Background()

	var stdout, stderr bytes.Buffer

	err := plugin.Run(ctx, "info", []string{}, &stdout, &stderr)
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Plugin: example-go-plugin")
	assert.Contains(t, output, "Version: 1.0.0")
	assert.Contains(t, output, "hello")
	assert.Contains(t, output, "info")
}

func TestExamplePlugin_Run_UnknownCommand(t *testing.T) {
	plugin := NewExamplePlugin()
	ctx := context.Background()

	var stdout, stderr bytes.Buffer

	err := plugin.Run(ctx, "unknown-command", []string{}, &stdout, &stderr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestExamplePlugin_HandleHook(t *testing.T) {
	plugin := NewExamplePlugin()
	ctx := context.Background()

	// HandleHook should not error
	err := plugin.HandleHook(ctx, "pre-run", []string{"test", "args"})
	assert.NoError(t, err)

	err = plugin.HandleHook(ctx, "post-run", []string{})
	assert.NoError(t, err)
}
