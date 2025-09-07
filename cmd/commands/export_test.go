package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/ambros/internal/repos/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestExportCommand_ValidateFlags(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		filter  string
		from    string
		to      string
		wantErr bool
	}{
		{
			name:    "valid json format",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "valid yaml format",
			format:  "yaml",
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "csv",
			wantErr: true,
		},
		{
			name:    "valid success filter",
			format:  "json",
			filter:  "success",
			wantErr: false,
		},
		{
			name:    "valid failed filter",
			format:  "json",
			filter:  "failed",
			wantErr: false,
		},
		{
			name:    "invalid filter",
			format:  "json",
			filter:  "invalid",
			wantErr: true,
		},
		{
			name:    "valid dates",
			format:  "json",
			from:    "2024-01-01",
			to:      "2024-01-31",
			wantErr: false,
		},
		{
			name:    "invalid from date",
			format:  "json",
			from:    "2024/01/01",
			wantErr: true,
		},
		{
			name:    "invalid to date",
			format:  "json",
			to:      "2024/01/31",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			mockRepo := new(mocks.MockRepository)
			cmd := NewExportCommand(logger, mockRepo)

			cmd.format = tt.format
			cmd.filter = tt.filter
			cmd.fromDate = tt.from
			cmd.toDate = tt.to

			err := cmd.validateFlags()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExportCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Sample commands for export
	sampleCommands := []models.Command{
		{
			Entity: models.Entity{
				ID:        "cmd1",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Name:      "echo",
			Arguments: []string{"hello"},
			Tags:      []string{"test"},
			Status:    true,
			Output:    "hello\n",
		},
		{
			Entity: models.Entity{
				ID:        "cmd2",
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
			Name:      "ls",
			Arguments: []string{"-la"},
			Tags:      []string{"file"},
			Status:    false,
			Error:     "permission denied",
		},
	}

	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "ambros-export-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		outputFile    string
		format        string
		tag           string
		filter        string
		expectedError bool
		verifyOutput  func(*testing.T, string)
	}{
		{
			name:       "Success - Export all commands as JSON",
			outputFile: filepath.Join(tmpDir, "all.json"),
			format:     "json",
			verifyOutput: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				assert.NoError(t, err)

				var exported map[string]interface{}
				err = json.Unmarshal(data, &exported)
				assert.NoError(t, err)

				commands, ok := exported["commands"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, commands, 2)

				metadata, ok := exported["metadata"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, float64(2), metadata["total"])
				assert.Equal(t, "json", metadata["format"])
			},
		},
		{
			name:       "Success - Export by tag",
			outputFile: filepath.Join(tmpDir, "tagged.json"),
			format:     "json",
			tag:        "test",
			verifyOutput: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				assert.NoError(t, err)

				var exported map[string]interface{}
				err = json.Unmarshal(data, &exported)
				assert.NoError(t, err)

				commands, ok := exported["commands"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, commands, 1)
			},
		},
		{
			name:       "Success - Export by status",
			outputFile: filepath.Join(tmpDir, "success.json"),
			format:     "json",
			filter:     "success",
			verifyOutput: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				assert.NoError(t, err)

				var exported map[string]interface{}
				err = json.Unmarshal(data, &exported)
				assert.NoError(t, err)

				commands, ok := exported["commands"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, commands, 1)
			},
		},
		{
			name:          "Error - Repository error",
			outputFile:    filepath.Join(tmpDir, "error.json"),
			format:        "json",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new mock for each test to avoid conflicts
			testMockRepo := new(mocks.MockRepository)
			cmd := NewExportCommand(logger, testMockRepo)

			// Set command fields
			cmd.outputFile = tt.outputFile
			cmd.format = tt.format
			cmd.tag = tt.tag
			cmd.filter = tt.filter

			// Setup mocks based on test requirements
			if tt.name == "Success - Export all commands as JSON" {
				testMockRepo.On("GetAllCommands").Return(sampleCommands, nil)
			} else if tt.name == "Success - Export by tag" {
				testMockRepo.On("SearchByTag", "test").Return(
					[]models.Command{sampleCommands[0]}, nil,
				)
			} else if tt.name == "Success - Export by status" {
				testMockRepo.On("SearchByStatus", true).Return(
					[]models.Command{sampleCommands[0]}, nil,
				)
			} else if tt.name == "Error - Repository error" {
				testMockRepo.On("GetAllCommands").Return(nil, assert.AnError)
			}

			// Execute command
			err := cmd.runE(cmd.cmd, []string{})

			// Verify error
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify output file
				if tt.verifyOutput != nil {
					tt.verifyOutput(t, tt.outputFile)
				}
			}

			// Verify mocks
			testMockRepo.AssertExpectations(t)
		})
	}
}

func TestExportCommand_FilterByDate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	commands := []models.Command{
		{
			Entity: models.Entity{
				ID:        "cmd1",
				CreatedAt: now,
			},
			Name: "test1",
		},
		{
			Entity: models.Entity{
				ID:        "cmd2",
				CreatedAt: yesterday,
			},
			Name: "test2",
		},
		{
			Entity: models.Entity{
				ID:        "cmd3",
				CreatedAt: twoDaysAgo,
			},
			Name: "test3",
		},
	}

	tests := []struct {
		name     string
		from     string
		to       string
		expected int
	}{
		{
			name:     "no filter",
			expected: 3,
		},
		{
			name:     "from only",
			from:     yesterday.Format("2006-01-02"),
			expected: 2,
		},
		{
			name:     "to only",
			to:       yesterday.Format("2006-01-02"),
			expected: 2,
		},
		{
			name:     "from and to",
			from:     twoDaysAgo.Format("2006-01-02"),
			to:       yesterday.Format("2006-01-02"),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewExportCommand(logger, mockRepo)
			cmd.fromDate = tt.from
			cmd.toDate = tt.to

			filtered := cmd.filterByDate(commands)
			assert.Equal(t, tt.expected, len(filtered))
		})
	}
}

func TestExportCommand_PrepareExportData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(mocks.MockRepository)
	cmd := NewExportCommand(logger, mockRepo)

	cmd.format = "json"
	cmd.tag = "test"
	cmd.filter = "success"
	cmd.fromDate = "2024-01-01"
	cmd.toDate = "2024-01-31"
	cmd.history = true

	sampleCommands := []models.Command{
		{
			Entity: models.Entity{ID: "cmd1"},
			Name:   "echo",
			Status: true,
		},
	}

	data := cmd.prepareExportData(sampleCommands)

	// Verify the structure
	exportData, ok := data.(struct {
		ExportDate time.Time        `json:"export_date" yaml:"export_date"`
		Commands   []models.Command `json:"commands" yaml:"commands"`
		Metadata   struct {
			Total    int    `json:"total" yaml:"total"`
			Format   string `json:"format" yaml:"format"`
			Filter   string `json:"filter,omitempty" yaml:"filter,omitempty"`
			Tag      string `json:"tag,omitempty" yaml:"tag,omitempty"`
			FromDate string `json:"from_date,omitempty" yaml:"from_date,omitempty"`
			ToDate   string `json:"to_date,omitempty" yaml:"to_date,omitempty"`
			History  bool   `json:"history" yaml:"history"`
		} `json:"metadata" yaml:"metadata"`
	})

	assert.True(t, ok)
	assert.Len(t, exportData.Commands, 1)
	assert.Equal(t, 1, exportData.Metadata.Total)
	assert.Equal(t, "json", exportData.Metadata.Format)
	assert.Equal(t, "test", exportData.Metadata.Tag)
	assert.Equal(t, "success", exportData.Metadata.Filter)
	assert.Equal(t, "2024-01-01", exportData.Metadata.FromDate)
	assert.Equal(t, "2024-01-31", exportData.Metadata.ToDate)
	assert.True(t, exportData.Metadata.History)
}
