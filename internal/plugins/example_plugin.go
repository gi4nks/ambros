package plugins

import (
	"context"
	"fmt"
	"io"

	"github.com/gi4nks/ambros/v3/internal/models"
)

// ExamplePlugin is a sample Go internal plugin that demonstrates
// how to implement the GoPlugin interface.
// This plugin provides a simple "hello" command and responds to hooks.
type ExamplePlugin struct {
	api CoreAPI
}

// NewExamplePlugin creates a new instance of the example plugin.
func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{}
}

// GetManifest implements GoPlugin.GetManifest
func (p *ExamplePlugin) GetManifest() models.Plugin {
	return models.Plugin{
		Name:        "example-go-plugin",
		Version:     "1.0.0",
		Description: "An example Go internal plugin demonstrating the plugin API",
		Author:      "Ambros Team",
		Enabled:     true,
		Type:        models.PluginTypeGoInternal,
		Commands: []models.PluginCommandDef{
			{
				Name:        "hello",
				Description: "Print a greeting message",
				Usage:       "ambros example-go-plugin hello [name]",
				Args:        []string{"name"},
			},
			{
				Name:        "info",
				Description: "Show plugin information",
				Usage:       "ambros example-go-plugin info",
				Args:        []string{},
			},
		},
		Hooks: []string{"pre-run", "post-run"},
	}
}

// Init implements GoPlugin.Init
func (p *ExamplePlugin) Init(api CoreAPI) error {
	p.api = api
	if p.api != nil {
		p.api.Logger().Info("Example Go plugin initialized")
	}
	return nil
}

// Run implements GoPlugin.Run
func (p *ExamplePlugin) Run(ctx context.Context, commandName string, args []string, stdout, stderr io.Writer) error {
	switch commandName {
	case "hello":
		name := "World"
		if len(args) > 0 {
			name = args[0]
		}
		fmt.Fprintf(stdout, "Hello, %s! 👋\n", name)
		fmt.Fprintf(stdout, "This message is from the example Go plugin.\n")
		return nil

	case "info":
		manifest := p.GetManifest()
		fmt.Fprintf(stdout, "Plugin: %s\n", manifest.Name)
		fmt.Fprintf(stdout, "Version: %s\n", manifest.Version)
		fmt.Fprintf(stdout, "Description: %s\n", manifest.Description)
		fmt.Fprintf(stdout, "Author: %s\n", manifest.Author)
		fmt.Fprintf(stdout, "Type: %s\n", manifest.Type)
		fmt.Fprintf(stdout, "Commands:\n")
		for _, cmd := range manifest.Commands {
			fmt.Fprintf(stdout, "  - %s: %s\n", cmd.Name, cmd.Description)
		}
		fmt.Fprintf(stdout, "Hooks subscribed: %v\n", manifest.Hooks)
		return nil

	default:
		return fmt.Errorf("unknown command: %s", commandName)
	}
}

// HandleHook implements GoPlugin.HandleHook
func (p *ExamplePlugin) HandleHook(ctx context.Context, hookName string, args []string) error {
	if p.api != nil {
		p.api.Logger().Debug("Example plugin received hook") // Using structured logging

	}
	// Example: Log the hook event
	// In a real plugin, you might:
	// - Send notifications
	// - Update external systems
	// - Trigger additional processing
	// - Record metrics
	return nil
}

// init registers the example plugin with the global registry when the package is loaded.
// Uncomment the init function below to automatically register this plugin.
// For production use, you might want to make registration more explicit.
/*
func init() {
	if err := RegisterPlugin(NewExamplePlugin()); err != nil {
		// Log registration failure (cannot use logger here as it's not initialized)
		fmt.Fprintf(os.Stderr, "Failed to register example Go plugin: %v\n", err)
	}
}
*/
