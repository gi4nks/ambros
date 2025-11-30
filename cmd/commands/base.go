package commands

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/plugins"
	"github.com/gi4nks/ambros/v3/internal/repos"
)

// BaseCommand is the base structure for all commands
type BaseCommand struct {
	cmd        *cobra.Command
	logger     *zap.Logger
	repository repos.RepositoryInterface
	api        plugins.CoreAPI // CoreAPI for plugin access
}

// NewBaseCommand creates a new base command with repository and optional CoreAPI
func NewBaseCommand(cmd *cobra.Command, logger *zap.Logger, repo repos.RepositoryInterface, api ...plugins.CoreAPI) *BaseCommand {
	bc := &BaseCommand{
		cmd:        cmd,
		logger:     logger,
		repository: repo,
	}
	if len(api) > 0 {
		bc.api = api[0]
	}
	return bc
}

// NewBaseCommandWithoutRepo creates a new base command without repository
func NewBaseCommandWithoutRepo(cmd *cobra.Command, logger *zap.Logger) *BaseCommand {
	return &BaseCommand{
		cmd:        cmd,
		logger:     logger,
		repository: nil,
	}
}

// Command returns the cobra command
func (bc *BaseCommand) Command() *cobra.Command {
	return bc.cmd
}

// HasRepository returns true if the command has a repository interface
func (bc *BaseCommand) HasRepository() bool {
	return bc.repository != nil
}

// Logger returns the logger instance
func (bc *BaseCommand) Logger() *zap.Logger {
	return bc.logger
}

// Repository returns the repository interface
func (bc *BaseCommand) Repository() repos.RepositoryInterface {
	return bc.repository
}

// API returns the CoreAPI instance
func (bc *BaseCommand) API() plugins.CoreAPI {
	return bc.api
}

// RepositoryInterface is an alias for the repository interface
type RepositoryInterface = repos.RepositoryInterface
