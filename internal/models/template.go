package models

import "strings"

// Template represents a command template
type Template struct {
	Entity
	Name        string            `json:"name"`
	Pattern     string            `json:"pattern"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables"`
	Tags        []string          `json:"tags"`
}

// BuildCommand builds a command from the template with provided arguments
func (t *Template) BuildCommand(args []string) []string {
	command := t.Pattern

	// Replace variables in the pattern
	for key, value := range t.Variables {
		placeholder := "{" + key + "}"
		command = strings.ReplaceAll(command, placeholder, value)
	}

	// Replace positional arguments
	for i, arg := range args {
		placeholder := "{" + string(rune('0'+i)) + "}"
		command = strings.ReplaceAll(command, placeholder, arg)
	}

	// Split the command into parts
	return strings.Fields(command)
}
