package models

import (
	"fmt"
	"strings"
)

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

	usedArgs := make([]bool, len(args))

	// Replace positional arguments
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		if strings.Contains(command, placeholder) {
			command = strings.ReplaceAll(command, placeholder, arg)
			usedArgs[i] = true
		}
	}

	parts := strings.Fields(command)

	// Append any remaining args that were not used as placeholders
	for i, arg := range args {
		if !usedArgs[i] {
			parts = append(parts, arg)
		}
	}

	return parts
}
