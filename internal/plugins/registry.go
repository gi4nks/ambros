package plugins

import (
	"fmt"
	"sync"

	"go.uber.org/zap" // New import
)

// InternalPluginRegistry manages the registration and retrieval of Go-based internal plugins.
// Plugins are typically registered during application initialization.
type InternalPluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]GoPlugin // Map from plugin name to its GoPlugin instance
	api     CoreAPI             // CoreAPI instance to be passed to plugins
}

// NewInternalPluginRegistry creates a new instance of the plugin registry.
func NewInternalPluginRegistry() *InternalPluginRegistry {
	return &InternalPluginRegistry{
		plugins: make(map[string]GoPlugin),
	}
}

// SetCoreAPI sets the CoreAPI instance that will be provided to registered plugins.
// This should be called once during Ambros's initialization.
func (r *InternalPluginRegistry) SetCoreAPI(api CoreAPI) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.api = api
	// Initialize all already registered plugins with the CoreAPI
	for _, plugin := range r.plugins {
		if err := plugin.Init(r.api); err != nil {
			r.api.Logger().Error("Failed to initialize Go plugin",
				zap.String("plugin", plugin.GetManifest().Name),
				zap.Error(err))
		}
	}
}

// RegisterPlugin registers a new Go-based internal plugin with the registry.
// Plugins should call this function in their `init()` methods.
func (r *InternalPluginRegistry) RegisterPlugin(plugin GoPlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.GetManifest().Name
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("Go plugin with name '%s' already registered", name)
	}

	r.plugins[name] = plugin

	// If CoreAPI is already set, initialize the plugin immediately
	if r.api != nil {
		if err := plugin.Init(r.api); err != nil {
			r.api.Logger().Error("Failed to initialize Go plugin after registration",
				zap.String("plugin", name),
				zap.Error(err))
			return fmt.Errorf("failed to initialize Go plugin '%s': %w", name, err)
		}
	}
	return nil
}

// GetPlugin retrieves a registered Go-based internal plugin by its name.
func (r *InternalPluginRegistry) GetPlugin(name string) (GoPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// GetAllPlugins returns a slice of all registered Go-based internal plugins.
func (r *InternalPluginRegistry) GetAllPlugins() []GoPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	allPlugins := make([]GoPlugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		allPlugins = append(allPlugins, plugin)
	}
	return allPlugins
}

// Global registry instance
var globalRegistry = NewInternalPluginRegistry()

// RegisterPlugin is a convenience function to register a plugin with the global registry.
func RegisterPlugin(plugin GoPlugin) error {
	return globalRegistry.RegisterPlugin(plugin)
}

// GetGlobalRegistry returns the singleton instance of the InternalPluginRegistry.
func GetGlobalRegistry() *InternalPluginRegistry {
	return globalRegistry
}
