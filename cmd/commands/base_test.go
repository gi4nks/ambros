package commands

import (
	"testing"

	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewBaseCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	baseCmd := NewBaseCommand(cmd, logger, mockRepo)

	assert.NotNil(t, baseCmd)
	assert.Equal(t, cmd, baseCmd.cmd)
	assert.Equal(t, logger, baseCmd.logger)
	assert.Equal(t, mockRepo, baseCmd.repository)
	assert.True(t, baseCmd.HasRepository())
}

func TestNewBaseCommandWithoutRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	baseCmd := NewBaseCommandWithoutRepo(cmd, logger)

	assert.NotNil(t, baseCmd)
	assert.Equal(t, cmd, baseCmd.cmd)
	assert.Equal(t, logger, baseCmd.logger)
	assert.Nil(t, baseCmd.repository)
	assert.False(t, baseCmd.HasRepository())
}

func TestBaseCommand_Command(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	baseCmd := NewBaseCommand(cmd, logger, mockRepo)

	assert.Equal(t, cmd, baseCmd.Command())
}

func TestBaseCommand_Logger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	baseCmd := NewBaseCommand(cmd, logger, mockRepo)

	assert.Equal(t, logger, baseCmd.Logger())
}

func TestBaseCommand_Repository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	baseCmd := NewBaseCommand(cmd, logger, mockRepo)

	assert.Equal(t, mockRepo, baseCmd.Repository())
}

func TestBaseCommand_HasRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	t.Run("with repository", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		baseCmd := NewBaseCommand(cmd, logger, mockRepo)

		assert.True(t, baseCmd.HasRepository())
	})

	t.Run("without repository", func(t *testing.T) {
		baseCmd := NewBaseCommandWithoutRepo(cmd, logger)

		assert.False(t, baseCmd.HasRepository())
	})

	t.Run("with nil repository", func(t *testing.T) {
		baseCmd := NewBaseCommand(cmd, logger, nil)

		assert.False(t, baseCmd.HasRepository())
	})
}

func TestBaseCommand_NilHandling(t *testing.T) {
	t.Run("nil command", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		mockRepo := new(mocks.MockRepository)

		baseCmd := NewBaseCommand(nil, logger, mockRepo)

		assert.Nil(t, baseCmd.Command())
		assert.Equal(t, logger, baseCmd.Logger())
		assert.Equal(t, mockRepo, baseCmd.Repository())
	})

	t.Run("nil logger", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		cmd := &cobra.Command{Use: "test"}

		baseCmd := NewBaseCommand(cmd, nil, mockRepo)

		assert.Equal(t, cmd, baseCmd.Command())
		assert.Nil(t, baseCmd.Logger())
		assert.Equal(t, mockRepo, baseCmd.Repository())
	})

	t.Run("all nil", func(t *testing.T) {
		baseCmd := NewBaseCommand(nil, nil, nil)

		assert.Nil(t, baseCmd.Command())
		assert.Nil(t, baseCmd.Logger())
		assert.Nil(t, baseCmd.Repository())
		assert.False(t, baseCmd.HasRepository())
	})
}

func TestBaseCommand_Consistency(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	cmd := &cobra.Command{
		Use:   "consistency-test",
		Short: "Test command consistency",
		Long:  "Test that base command maintains consistency",
	}

	baseCmd := NewBaseCommand(cmd, logger, mockRepo)

	// Test that multiple calls return the same instances
	assert.Equal(t, baseCmd.Command(), baseCmd.Command())
	assert.Equal(t, baseCmd.Logger(), baseCmd.Logger())
	assert.Equal(t, baseCmd.Repository(), baseCmd.Repository())

	// Test that HasRepository is consistent
	hasRepo1 := baseCmd.HasRepository()
	hasRepo2 := baseCmd.HasRepository()
	assert.Equal(t, hasRepo1, hasRepo2)
	assert.True(t, hasRepo1)
}

func TestBaseCommand_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	// Test creating a command that could be used in practice
	cmd := &cobra.Command{
		Use:   "integration-test",
		Short: "Integration test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This would be where a real command implementation would go
			return nil
		},
	}

	baseCmd := NewBaseCommand(cmd, logger, mockRepo)

	// Verify all components are properly set up
	assert.NotNil(t, baseCmd.Command())
	assert.NotNil(t, baseCmd.Logger())
	assert.NotNil(t, baseCmd.Repository())
	assert.True(t, baseCmd.HasRepository())

	// Test that the command can be executed (even though it does nothing)
	err := baseCmd.Command().Execute()
	assert.NoError(t, err)
}

// TestRepositoryInterface ensures our interface has the expected methods
func TestRepositoryInterface(t *testing.T) {
	mockRepo := new(mocks.MockRepository)

	// This test ensures that our mock implements the interface correctly
	var _ RepositoryInterface = mockRepo

	// Test that we can assign the mock to the interface
	var repo RepositoryInterface = mockRepo
	assert.NotNil(t, repo)
}
