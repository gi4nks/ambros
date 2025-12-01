package models

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Plugin struct {
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Description  string             `json:"description"`
	Author       string             `json:"author"`
	Enabled      bool               `json:"enabled"`
	Type         PluginType         `json:"type"`       // New field for plugin type
	Executable   string             `json:"executable"` // For shell/external plugins
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

// PluginType defines the type of plugin
type PluginType string

const (
	PluginTypeShell      PluginType = "shell"       // Shell script based plugin
	PluginTypeGoInternal PluginType = "go_internal" // Go plugin built into Ambros binary
	// PluginTypeGoExternal PluginType = "go_external" // Go plugin as shared library (future)
	// PluginTypeGRPC       PluginType = "grpc"        // GRPC based plugin (future)
)

type Entity struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	TerminatedAt time.Time `json:"terminated_at"`
}

type Command struct {
	Entity
	Name      string            `json:"name"`
	Command   string            `json:"command"`
	Arguments []string          `json:"arguments"`
	Status    bool              `json:"status"`
	Output    string            `json:"output"`
	Error     string            `json:"error"`
	Tags      []string          `json:"tags,omitempty"`
	Category  string            `json:"category,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
}

type CommandChain struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Commands    []string  `json:"commands"` // Command IDs
	Conditional bool      `json:"conditional"`
	CreatedAt   time.Time `json:"created_at"`
}

type CommandTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Arguments   []string          `json:"arguments"`
	Variables   map[string]string `json:"variables"`
}

type SearchQuery struct {
	Text     string    `json:"text"`
	Tags     []string  `json:"tags,omitempty"`
	Category string    `json:"category,omitempty"`
	DateFrom time.Time `json:"date_from,omitempty"`
	DateTo   time.Time `json:"date_to,omitempty"`
	Status   *bool     `json:"status,omitempty"`
}

type ExecutedCommand struct {
	ID      string    `json:"id"`
	Command string    `json:"command"`
	Status  bool      `json:"status"`
	When    time.Time `json:"when"`
	Index   int       `json:"index"`
	Order   int       `json:"order"`
}

func (c *Command) Clone() *Command {
	// Create a new Command object with the same field values as the original
	clone := &Command{
		Entity: Entity{
			ID:           c.ID,
			CreatedAt:    c.CreatedAt,
			TerminatedAt: c.TerminatedAt,
		},
		Name:      c.Name,
		Arguments: make([]string, len(c.Arguments)),
		Status:    c.Status,
		Output:    c.Output,
		Error:     c.Error,
		Command:   c.Command,
		Tags:      append([]string{}, c.Tags...),
		Category:  c.Category,
		Variables: make(map[string]string),
	}

	// Copy the elements of the Arguments slice to the clone's Arguments slice
	copy(clone.Arguments, c.Arguments)

	return clone
}

func (c Command) String() (string, error) {
	b, err := json.Marshal(c)

	if err != nil {
		return "{}", err
	}
	return string(b), nil
}

func (c *Command) AsHistory() string {
	return "Name : " + c.Name + " --> Arguments : " + strings.Join(c.Arguments, " ")
}

func (c *Command) AsExecutedCommand(order int) ExecutedCommand {
	s := c.Name + " " + strings.Join(c.Arguments, " ")
	return ExecutedCommand{Order: order, ID: c.ID, Command: s, Status: c.Status, When: c.CreatedAt}
}

func (c Command) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ID":           c.ID,
		"Name":         c.Name,
		"Arguments":    c.Arguments,
		"Status":       c.Status,
		"Output":       c.Output,
		"Error":        c.Error,
		"Command":      c.Command,
		"Tags":         c.Tags,
		"Category":     c.Category,
		"Variables":    c.Variables,
		"CreatedAt":    c.CreatedAt,
		"TerminatedAt": c.TerminatedAt,
	}
}

func (c *Command) FromMap(frommap map[string]interface{}) {
	c.ID = frommap["ID"].(string)
	c.Name = frommap["Name"].(string)
	c.Arguments = frommap["Arguments"].([]string)
	c.Status = frommap["Status"].(bool)
	c.Output = frommap["Output"].(string)
	c.Error = frommap["Error"].(string)
	c.CreatedAt = frommap["CreatedAt"].(time.Time)
	c.TerminatedAt = frommap["TerminatedAt"].(time.Time)
}

func (c Command) AsStoredCommand() string {
	return "[" + c.ID + "] " + c.Name + " " + strings.Join(c.Arguments, " ")
}

func (c ExecutedCommand) AsFlatCommand() string {
	return "{" + c.When.Format("02.01.2006 15:04:05") + "} [id: " + c.ID + ", status: " + strconv.FormatBool(c.Status) + "] " + c.Command
}

func (c ExecutedCommand) Print(logger zap.Logger) {
	logger.Info("Executed Command",
		zap.String("when", c.When.Format("02.01.2006 15:04:05")),
		zap.String("id", c.ID),
		zap.Bool("status", c.Status),
		zap.String("command", c.Command),
	)
}
