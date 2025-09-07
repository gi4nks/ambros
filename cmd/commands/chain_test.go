package commands

import (
	"testing"
	"time"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestChainCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("execute chain", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"exec", "testchain"})
		assert.NoError(t, err)
	})

	t.Run("execute chain with conditional flag", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)
		chainCmd.conditional = true

		err := chainCmd.runE(chainCmd.cmd, []string{"exec", "testchain"})
		assert.NoError(t, err)
	})

	t.Run("execute chain - missing name", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"exec"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chain name required")
	})

	t.Run("create chain", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		// Mock the commands that will be verified
		cmd1 := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
		}
		cmd2 := &models.Command{
			Entity: models.Entity{ID: "cmd2"},
			Name:   "ls",
		}

		mockRepo.On("Get", "cmd1").Return(cmd1, nil)
		mockRepo.On("Get", "cmd2").Return(cmd2, nil)

		chainCmd := NewChainCommand(logger, mockRepo)
		chainCmd.description = "Test chain description"

		err := chainCmd.runE(chainCmd.cmd, []string{"create", "mychain", "cmd1,cmd2"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create chain - command not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "nonexistent").Return(nil, assert.AnError)

		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"create", "mychain", "nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("create chain - missing arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"create", "mychain"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command IDs required")
	})

	t.Run("list chains", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"list"})
		assert.NoError(t, err)
	})

	t.Run("delete chain", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"delete", "testchain"})
		assert.NoError(t, err)
	})

	t.Run("delete chain - missing name", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"delete"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chain name required")
	})

	t.Run("unknown subcommand", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		err := chainCmd.runE(chainCmd.cmd, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand")
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)

		assert.Equal(t, "chain", chainCmd.cmd.Use)
		assert.Equal(t, "Execute or manage command chains", chainCmd.cmd.Short)
		assert.NotNil(t, chainCmd.Command())
		assert.Equal(t, chainCmd.cmd, chainCmd.Command())

		// Test flags
		assert.NotNil(t, chainCmd.cmd.Flags().Lookup("name"))
		assert.NotNil(t, chainCmd.cmd.Flags().Lookup("desc"))
		assert.NotNil(t, chainCmd.cmd.Flags().Lookup("conditional"))
		assert.NotNil(t, chainCmd.cmd.Flags().Lookup("store"))

		// Test flag shortcuts
		assert.NotNil(t, chainCmd.cmd.Flags().ShorthandLookup("n"))
		assert.NotNil(t, chainCmd.cmd.Flags().ShorthandLookup("d"))
		assert.NotNil(t, chainCmd.cmd.Flags().ShorthandLookup("c"))
		assert.NotNil(t, chainCmd.cmd.Flags().ShorthandLookup("s"))
	})
}

func TestChainCommand_CreateChain(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("create chain with multiple commands", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		commands := []string{"cmd1", "cmd2", "cmd3"}
		for _, cmdId := range commands {
			cmd := &models.Command{
				Entity: models.Entity{ID: cmdId},
				Name:   "test-" + cmdId,
			}
			mockRepo.On("Get", cmdId).Return(cmd, nil)
		}

		chainCmd := NewChainCommand(logger, mockRepo)
		chainCmd.description = "Multi-command chain"

		err := chainCmd.createChain("multichain", commands)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create chain with whitespace in command IDs", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		cmd1 := &models.Command{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
		}
		cmd2 := &models.Command{
			Entity: models.Entity{ID: "cmd2"},
			Name:   "ls",
		}

		mockRepo.On("Get", "cmd1").Return(cmd1, nil)
		mockRepo.On("Get", "cmd2").Return(cmd2, nil)

		chainCmd := NewChainCommand(logger, mockRepo)

		// Test with spaces around command IDs
		err := chainCmd.createChain("spacechain", []string{" cmd1 ", " cmd2 "})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestChainCommand_GenerateChainID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	chainCmd := NewChainCommand(logger, mockRepo)

	id1 := chainCmd.generateChainID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := chainCmd.generateChainID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "CHAIN-")
	assert.Contains(t, id2, "CHAIN-")
}

func TestChainCommand_ExecuteChain(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("execute chain with store flag", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)
		chainCmd.store = true

		err := chainCmd.executeChain("testchain")
		assert.NoError(t, err)
	})

	t.Run("execute chain without conditional flag", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		chainCmd := NewChainCommand(logger, mockRepo)
		chainCmd.conditional = false

		err := chainCmd.executeChain("testchain")
		assert.NoError(t, err)
	})
}
