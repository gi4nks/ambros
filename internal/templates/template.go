package templates

import (
	"bytes"
	"text/template"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/ambros/internal/repos"
)

type CommandTemplate struct {
	Name        string
	Description string
	Command     string
	Arguments   []string
	Variables   map[string]string
}

func (t *CommandTemplate) Execute(vars map[string]string) (*models.Command, error) {
	// Replace variables in command and arguments
	// Return a new command instance
	cmdTemplate, err := template.New("command").Parse(t.Command)
	if err != nil {
		return nil, err
	}

	var cmdBuffer bytes.Buffer
	if err := cmdTemplate.Execute(&cmdBuffer, vars); err != nil {
		return nil, err
	}

	// Process arguments
	processedArgs := make([]string, len(t.Arguments))
	for i, arg := range t.Arguments {
		argTemplate, err := template.New("arg").Parse(arg)
		if err != nil {
			return nil, err
		}

		var argBuffer bytes.Buffer
		if err := argTemplate.Execute(&argBuffer, vars); err != nil {
			return nil, err
		}
		processedArgs[i] = argBuffer.String()
	}

	return &models.Command{
		Name:      cmdBuffer.String(),
		Arguments: processedArgs,
		Variables: vars,
	}, nil
}

type TemplateManager struct {
	repo *repos.Repository
}

func NewTemplateManager(repo *repos.Repository) *TemplateManager {
	return &TemplateManager{repo: repo}
}

func (tm *TemplateManager) Execute(tmpl *models.CommandTemplate, vars map[string]string) (*models.Command, error) {
	// Process command template
	cmdTemplate, err := template.New("command").Parse(tmpl.Command)
	if err != nil {
		return nil, err
	}

	var cmdBuffer bytes.Buffer
	if err := cmdTemplate.Execute(&cmdBuffer, vars); err != nil {
		return nil, err
	}

	// Process arguments
	processedArgs := make([]string, len(tmpl.Arguments))
	for i, arg := range tmpl.Arguments {
		argTemplate, err := template.New("arg").Parse(arg)
		if err != nil {
			return nil, err
		}

		var argBuffer bytes.Buffer
		if err := argTemplate.Execute(&argBuffer, vars); err != nil {
			return nil, err
		}
		processedArgs[i] = argBuffer.String()
	}

	return &models.Command{
		Name:      cmdBuffer.String(),
		Arguments: processedArgs,
		Variables: vars,
	}, nil
}
