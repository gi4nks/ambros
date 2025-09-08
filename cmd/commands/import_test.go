package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func TestImportCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "ambros-import-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("successful JSON import", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		// Create test data
		testCommands := []models.Command{
			{
				Entity: models.Entity{ID: "cmd1"},
				Name:   "echo",
				Status: true,
			},
			{
				Entity: models.Entity{ID: "cmd2"},
				Name:   "ls",
				Status: false,
			},
		}

		importData := struct {
			Commands []models.Command `json:"commands"`
		}{Commands: testCommands}

		// Write test file
		jsonData, _ := json.Marshal(importData)
		testFile := filepath.Join(tmpDir, "test.json")
		err := os.WriteFile(testFile, jsonData, 0644)
		assert.NoError(t, err)

		// Setup mocks
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil).Twice()

		importCmd := NewImportCommand(logger, mockRepo)
		importCmd.inputFile = testFile
		importCmd.format = "json"

		err = importCmd.runE(importCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful YAML import", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		testCommands := []models.Command{
			{
				Entity: models.Entity{ID: "cmd1"},
				Name:   "echo",
				Status: true,
			},
		}

		importData := struct {
			Commands []models.Command `yaml:"commands"`
		}{Commands: testCommands}

		// Write test file
		yamlData, _ := yaml.Marshal(importData)
		testFile := filepath.Join(tmpDir, "test.yaml")
		err := os.WriteFile(testFile, yamlData, 0644)
		assert.NoError(t, err)

		// Setup mocks
		mockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil)

		importCmd := NewImportCommand(logger, mockRepo)
		importCmd.inputFile = testFile
		importCmd.format = "yaml"

		err = importCmd.runE(importCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("dry run mode", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		testCommands := []models.Command{
			{
				Entity: models.Entity{ID: "cmd1"},
				Name:   "echo",
				Status: true,
			},
		}

		importData := struct {
			Commands []models.Command `json:"commands"`
		}{Commands: testCommands}

		jsonData, _ := json.Marshal(importData)
		testFile := filepath.Join(tmpDir, "dryrun.json")
		err := os.WriteFile(testFile, jsonData, 0644)
		assert.NoError(t, err)

		// Setup mocks for existence check
		mockRepo.On("Get", "cmd1").Return(nil, assert.AnError)

		importCmd := NewImportCommand(logger, mockRepo)
		importCmd.inputFile = testFile
		importCmd.format = "json"
		importCmd.dryRun = true

		err = importCmd.runE(importCmd.cmd, []string{})
		assert.NoError(t, err)

		// Should not call Put in dry run mode
		mockRepo.AssertNotCalled(t, "Put")
		mockRepo.AssertExpectations(t)
	})

	t.Run("skip existing commands", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		testCommands := []models.Command{
			{
				Entity: models.Entity{ID: "existing"},
				Name:   "echo",
				Status: true,
			},
			{
				Entity: models.Entity{ID: "new"},
				Name:   "ls",
				Status: true,
			},
		}

		importData := struct {
			Commands []models.Command `json:"commands"`
		}{Commands: testCommands}

		jsonData, _ := json.Marshal(importData)
		testFile := filepath.Join(tmpDir, "skip.json")
		err := os.WriteFile(testFile, jsonData, 0644)
		assert.NoError(t, err)

		// Setup mocks
		existingCmd := &models.Command{Entity: models.Entity{ID: "existing"}}
		mockRepo.On("Get", "existing").Return(existingCmd, nil) // Command exists
		mockRepo.On("Get", "new").Return(nil, assert.AnError)   // Command doesn't exist
		mockRepo.On("Put", mock.Anything, mock.MatchedBy(func(cmd models.Command) bool {
			return cmd.ID == "new" // Only new command should be imported
		})).Return(nil)

		importCmd := NewImportCommand(logger, mockRepo)
		importCmd.inputFile = testFile
		importCmd.format = "json"
		importCmd.skipExisting = true

		err = importCmd.runE(importCmd.cmd, []string{})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("file not found error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		importCmd := NewImportCommand(logger, mockRepo)
		importCmd.inputFile = "/nonexistent/file.json"
		importCmd.format = "json"

		err = importCmd.runE(importCmd.cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "input file does not exist")
	})

	t.Run("invalid format error", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		assert.NoError(t, err)

		importCmd := NewImportCommand(logger, mockRepo)
		importCmd.inputFile = testFile
		importCmd.format = "invalid"

		err = importCmd.runE(importCmd.cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})

	t.Run("invalid JSON file", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)

		testFile := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(testFile, []byte("invalid json"), 0644)
		assert.NoError(t, err)

		importCmd := NewImportCommand(logger, mockRepo)
		importCmd.inputFile = testFile
		importCmd.format = "json"

		err = importCmd.runE(importCmd.cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse input file")
	})

	t.Run("command structure validation", func(t *testing.T) {
		mockRepo := new(mocks.MockRepository)
		importCmd := NewImportCommand(logger, mockRepo)

		assert.Equal(t, "import", importCmd.cmd.Use)
		assert.Equal(t, "Import commands from file", importCmd.cmd.Short)
		assert.NotNil(t, importCmd.Command())
		assert.Equal(t, importCmd.cmd, importCmd.Command())

		// Test flags
		assert.NotNil(t, importCmd.cmd.Flags().Lookup("input"))
		assert.NotNil(t, importCmd.cmd.Flags().Lookup("format"))
		assert.NotNil(t, importCmd.cmd.Flags().Lookup("merge"))
		assert.NotNil(t, importCmd.cmd.Flags().Lookup("dry-run"))
		assert.NotNil(t, importCmd.cmd.Flags().Lookup("skip-existing"))

		// Test flag defaults
		formatFlag := importCmd.cmd.Flags().Lookup("format")
		assert.Equal(t, "json", formatFlag.DefValue)
	})
}

func TestImportCommand_CommandExists(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	importCmd := NewImportCommand(logger, mockRepo)

	t.Run("command exists", func(t *testing.T) {
		cmd := &models.Command{Entity: models.Entity{ID: "existing"}}
		mockRepo.On("Get", "existing").Return(cmd, nil)

		exists := importCmd.commandExists("existing")
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("command does not exist", func(t *testing.T) {
		mockRepo.On("Get", "nonexistent").Return(nil, assert.AnError)

		exists := importCmd.commandExists("nonexistent")
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})
}

func TestImportCommand_FormatStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	importCmd := NewImportCommand(logger, mockRepo)

	tests := []struct {
		name     string
		status   bool
		expected string
	}{
		{
			name:     "success status",
			status:   true,
			expected: "✅ Success",
		},
		{
			name:     "failed status",
			status:   false,
			expected: "❌ Failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := importCmd.formatStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
