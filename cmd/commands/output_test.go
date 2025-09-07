package commands

import (
	"testing"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestOutputCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("successful output retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		cmd := &models.Command{
			Entity: models.Entity{
				ID: "test-id",
			},
			Name:   "test",
			Output: "test output",
		}
		mockRepo.On("Get", "test-id").Return(cmd, nil)

		outputCmd := NewOutputCommand(logger, mockRepo)
		outputCmd.commandId = "test-id"

		err := outputCmd.runE(outputCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command not found", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "nonexistent-id").Return(nil, assert.AnError)

		outputCmd := NewOutputCommand(logger, mockRepo)
		outputCmd.commandId = "nonexistent-id"

		err := outputCmd.runE(outputCmd.cmd, []string{})
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty command ID", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		outputCmd := NewOutputCommand(logger, mockRepo)
		outputCmd.commandId = ""

		err := outputCmd.runE(outputCmd.cmd, []string{})
		assert.Error(t, err)
	})

	t.Run("command with error output", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		cmd := &models.Command{
			Entity: models.Entity{
				ID: "error-cmd",
			},
			Name:   "failing-test",
			Output: "",
			Error:  "command failed with error",
		}
		mockRepo.On("Get", "error-cmd").Return(cmd, nil)

		outputCmd := NewOutputCommand(logger, mockRepo)
		outputCmd.commandId = "error-cmd"

		err := outputCmd.runE(outputCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command with both output and error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		cmd := &models.Command{
			Entity: models.Entity{
				ID: "mixed-cmd",
			},
			Name:   "mixed-output",
			Output: "some output",
			Error:  "some error",
		}
		mockRepo.On("Get", "mixed-cmd").Return(cmd, nil)

		outputCmd := NewOutputCommand(logger, mockRepo)
		outputCmd.commandId = "mixed-cmd"

		err := outputCmd.runE(outputCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		outputCmd := NewOutputCommand(logger, mockRepo)

		assert.Equal(t, "output", outputCmd.cmd.Use)
		assert.Equal(t, "Show the output of a stored command", outputCmd.cmd.Short)
		assert.NotNil(t, outputCmd.Command())
		assert.Equal(t, outputCmd.cmd, outputCmd.Command())

		// Test that the id flag exists
		idFlag := outputCmd.cmd.Flags().Lookup("id")
		assert.NotNil(t, idFlag)

		// Test short flag
		iFlag := outputCmd.cmd.Flags().ShorthandLookup("i")
		assert.NotNil(t, iFlag)
		assert.Equal(t, iFlag, idFlag)
	})
}
