package commands

import (
	"context"
	"testing"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// CommandTestSuite provides utilities for testing commands
type CommandTestSuite struct {
	T        *testing.T
	Logger   *zap.Logger
	MockRepo *mocks.MockRepository
}

// NewCommandTestSuite creates a new test suite
func NewCommandTestSuite(t *testing.T) *CommandTestSuite {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	return &CommandTestSuite{
		T:        t,
		Logger:   logger,
		MockRepo: mockRepo,
	}
}

// ExecuteCommand executes a command with arguments
func (s *CommandTestSuite) ExecuteCommand(cmd *cobra.Command, args ...string) error {
	cmd.SetArgs(args)
	return cmd.Execute()
}

// AssertCommandOutput verifies command output and error
func (s *CommandTestSuite) AssertCommandOutput(cmd *cobra.Command, args []string, expectedError error) {
	err := s.ExecuteCommand(cmd, args...)
	if expectedError != nil {
		assert.Error(s.T, err)
		assert.Equal(s.T, expectedError.Error(), err.Error())
	} else {
		assert.NoError(s.T, err)
	}
}

// SetupMockPut sets up mock expectations for Put operation
func (s *CommandTestSuite) SetupMockPut(cmd *models.Command, err error) {
	s.MockRepo.On("Put", context.Background(), *cmd).Return(err)
}

// SetupMockGet sets up mock expectations for Get operation
func (s *CommandTestSuite) SetupMockGet(id string, cmd *models.Command, err error) {
	s.MockRepo.On("Get", id).Return(cmd, err)
}

// SetupMockFindById sets up mock expectations for FindById operation
func (s *CommandTestSuite) SetupMockFindById(id string, cmd models.Command, err error) {
	s.MockRepo.On("FindById", id).Return(cmd, err)
}

// SetupMockGetAllCommands sets up mock expectations for GetAllCommands operation
func (s *CommandTestSuite) SetupMockGetAllCommands(commands []models.Command, err error) {
	s.MockRepo.On("GetAllCommands").Return(commands, err)
}

// SetupMockGetLimitCommands sets up mock expectations for GetLimitCommands operation
func (s *CommandTestSuite) SetupMockGetLimitCommands(limit int, commands []models.Command, err error) {
	s.MockRepo.On("GetLimitCommands", limit).Return(commands, err)
}

// SetupMockSearchByTag sets up mock expectations for SearchByTag operation
func (s *CommandTestSuite) SetupMockSearchByTag(tag string, commands []models.Command, err error) {
	s.MockRepo.On("SearchByTag", tag).Return(commands, err)
}

// SetupMockSearchByStatus sets up mock expectations for SearchByStatus operation
func (s *CommandTestSuite) SetupMockSearchByStatus(success bool, commands []models.Command, err error) {
	s.MockRepo.On("SearchByStatus", success).Return(commands, err)
}

// SetupMockGetTemplate sets up mock expectations for GetTemplate operation
func (s *CommandTestSuite) SetupMockGetTemplate(name string, template *models.Template, err error) {
	s.MockRepo.On("GetTemplate", name).Return(template, err)
}

// SetupMockPush sets up mock expectations for Push operation
func (s *CommandTestSuite) SetupMockPush(cmd models.Command, err error) {
	s.MockRepo.On("Push", cmd).Return(err)
}

// VerifyMocks verifies all mock expectations
func (s *CommandTestSuite) VerifyMocks() {
	s.MockRepo.AssertExpectations(s.T)
}

// CreateTestCommand creates a basic test command for testing
func (s *CommandTestSuite) CreateTestCommand(id, name string) *models.Command {
	return &models.Command{
		Entity: models.Entity{
			ID: id,
		},
		Name:   name,
		Status: true,
	}
}

// CommandInterface defines the standard interface for all commands
type CommandInterface interface {
	// Command returns the cobra command
	Command() *cobra.Command
}

// BaseCommandInterface extends CommandInterface with common functionality
type BaseCommandInterface interface {
	CommandInterface

	// HasRepository returns true if the command has a repository
	HasRepository() bool

	// Logger returns the logger instance
	Logger() *zap.Logger

	// Repository returns the repository interface
	Repository() RepositoryInterface
}
