package commands

import (
	"testing"
	"time"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestStoreCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("store simple command", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "echo" && len(cmd.Arguments) == 1 && cmd.Arguments[0] == "hello"
		})).Return(nil)

		storeCmd := NewStoreCommand(logger, mockRepo)

		err := storeCmd.runE(storeCmd.cmd, []string{"echo", "hello"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("store command with name", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Get", "my-echo").Return(nil, assert.AnError) // Command doesn't exist
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.ID == "my-echo" && cmd.Name == "echo"
		})).Return(nil)

		storeCmd := NewStoreCommand(logger, mockRepo)
		storeCmd.name = "my-echo"

		err := storeCmd.runE(storeCmd.cmd, []string{"echo", "hello"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("store command with metadata", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "ls" &&
				len(cmd.Tags) == 3 && // "tag1", "tag2", "stored"
				cmd.Category == "files" &&
				cmd.Output == "List files in directory"
		})).Return(nil)

		storeCmd := NewStoreCommand(logger, mockRepo)
		storeCmd.tags = []string{"tag1", "tag2"}
		storeCmd.category = "files"
		storeCmd.description = "List files in directory"

		err := storeCmd.runE(storeCmd.cmd, []string{"ls", "-la"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("store command with existing name without force", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		existingCmd := &models.Command{Entity: models.Entity{ID: "existing"}}
		mockRepo.On("Get", "existing").Return(existingCmd, nil) // Command exists

		storeCmd := NewStoreCommand(logger, mockRepo)
		storeCmd.name = "existing"
		storeCmd.force = false

		err := storeCmd.runE(storeCmd.cmd, []string{"echo", "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("store command with existing name with force", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		// No Get expectation needed when force=true, as existence check is skipped
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.ID == "existing" && cmd.Name == "echo"
		})).Return(nil)

		storeCmd := NewStoreCommand(logger, mockRepo)
		storeCmd.name = "existing"
		storeCmd.force = true

		err := storeCmd.runE(storeCmd.cmd, []string{"echo", "overwrite"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(assert.AnError)

		storeCmd := NewStoreCommand(logger, mockRepo)

		err := storeCmd.runE(storeCmd.cmd, []string{"echo", "test"})
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		storeCmd := NewStoreCommand(logger, mockRepo)

		assert.Equal(t, "store [command...]", storeCmd.cmd.Use)
		assert.Equal(t, "Store a command for future use", storeCmd.cmd.Short)
		assert.NotNil(t, storeCmd.Command())
		assert.Equal(t, storeCmd.cmd, storeCmd.Command())

		// Test flags
		assert.NotNil(t, storeCmd.cmd.Flags().Lookup("name"))
		assert.NotNil(t, storeCmd.cmd.Flags().Lookup("description"))
		assert.NotNil(t, storeCmd.cmd.Flags().Lookup("tag"))
		assert.NotNil(t, storeCmd.cmd.Flags().Lookup("category"))
		assert.NotNil(t, storeCmd.cmd.Flags().Lookup("force"))

		// Test short flags
		nFlag := storeCmd.cmd.Flags().ShorthandLookup("n")
		assert.NotNil(t, nFlag)
		assert.Equal(t, nFlag, storeCmd.cmd.Flags().Lookup("name"))

		dFlag := storeCmd.cmd.Flags().ShorthandLookup("d")
		assert.NotNil(t, dFlag)
		assert.Equal(t, dFlag, storeCmd.cmd.Flags().Lookup("description"))

		tFlag := storeCmd.cmd.Flags().ShorthandLookup("t")
		assert.NotNil(t, tFlag)
		assert.Equal(t, tFlag, storeCmd.cmd.Flags().Lookup("tag"))

		cFlag := storeCmd.cmd.Flags().ShorthandLookup("c")
		assert.NotNil(t, cFlag)
		assert.Equal(t, cFlag, storeCmd.cmd.Flags().Lookup("category"))
	})
}

func TestStoreCommand_CommandExists(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	storeCmd := NewStoreCommand(logger, mockRepo)

	t.Run("command exists", func(t *testing.T) {
		existingCmd := &models.Command{Entity: models.Entity{ID: "existing"}}
		mockRepo.On("Get", "existing").Return(existingCmd, nil)

		exists := storeCmd.commandExists("existing")
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command does not exist", func(t *testing.T) {
		mockRepo.On("Get", "nonexistent").Return(nil, assert.AnError)

		exists := storeCmd.commandExists("nonexistent")
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})
}

func TestStoreCommand_GenerateCommandID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	storeCmd := NewStoreCommand(logger, mockRepo)

	id1 := storeCmd.generateCommandID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := storeCmd.generateCommandID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "STORED-")
	assert.Contains(t, id2, "STORED-")
}

func TestStoreCommand_WithComplexArguments(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("store command with multiple arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		// Command doesn't exist, so Get returns an error
		mockRepo.On("Get", "backup-sync").Return(nil, assert.AnError)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "rsync" &&
				len(cmd.Arguments) == 3 &&
				cmd.Arguments[0] == "-av" &&
				cmd.Arguments[1] == "src/" &&
				cmd.Arguments[2] == "dest/"
		})).Return(nil)

		storeCmd := NewStoreCommand(logger, mockRepo)
		storeCmd.name = "backup-sync"
		storeCmd.tags = []string{"backup", "sync"}
		storeCmd.description = "Backup source to destination"

		err := storeCmd.runE(storeCmd.cmd, []string{"rsync", "-av", "src/", "dest/"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("store command with no arguments", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.Name == "pwd" && len(cmd.Arguments) == 0
		})).Return(nil)

		storeCmd := NewStoreCommand(logger, mockRepo)

		err := storeCmd.runE(storeCmd.cmd, []string{"pwd"})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
