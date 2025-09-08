package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos/mocks"
)

func TestTestCommand(t *testing.T) {
	suite := NewCommandTestSuite(t)

	tests := []struct {
		name          string
		args          []string
		flags         map[string]string
		setupMocks    func(*mocks.MockRepository)
		expectedError string
		expectedCalls int
	}{
		{
			name:          "default test command",
			expectedCalls: 3, // Default count is 3
			setupMocks: func(m *mocks.MockRepository) {
				m.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil).Times(3)
			},
		},
		{
			name: "custom count",
			flags: map[string]string{
				"number": "5",
			},
			expectedCalls: 5,
			setupMocks: func(m *mocks.MockRepository) {
				m.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil).Times(5)
			},
		},
		{
			name: "with cleanup",
			flags: map[string]string{
				"cleanup": "true",
			},
			expectedCalls: 3,
			setupMocks: func(m *mocks.MockRepository) {
				m.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil).Times(3)
			},
		},
		{
			name: "repository error",
			setupMocks: func(m *mocks.MockRepository) {
				m.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(
					errors.NewError(errors.ErrRepositoryWrite, "failed to store", nil))
			},
			expectedError: "failed to store test command 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite.MockRepo.ExpectedCalls = nil
			suite.MockRepo.Calls = nil

			// Create new command instance
			cmd := NewTestCommand(suite.Logger, suite.MockRepo)

			// Set flags if provided
			if tt.flags != nil {
				for name, value := range tt.flags {
					cmd.Command().Flags().Set(name, value)
				}
			}

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(suite.MockRepo)
			}

			// Execute command
			err := suite.ExecuteCommand(cmd.Command(), tt.args...)

			// Verify error
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCalls, len(suite.MockRepo.Calls))
			}

			// Verify mocks
			suite.MockRepo.AssertExpectations(t)
		})
	}
}

func TestGenerateTestCommand(t *testing.T) {
	suite := NewCommandTestSuite(t)
	cmd := NewTestCommand(suite.Logger, suite.MockRepo)

	// Generate multiple test commands to verify randomization
	cmds := make([]*models.Command, 10)
	for i := range cmds {
		cmds[i] = cmd.generateTestCommand()
	}

	// Verify common properties
	for _, c := range cmds {
		// Check basic structure
		assert.NotEmpty(t, c.ID)
		assert.True(t, c.ID[:5] == "TEST-")
		assert.NotEmpty(t, c.Name)
		assert.NotNil(t, c.Arguments)
		assert.NotZero(t, c.CreatedAt)
		assert.NotZero(t, c.TerminatedAt)
		assert.True(t, c.TerminatedAt.After(c.CreatedAt))

		// Check tags
		assert.Contains(t, c.Tags, "test")
		assert.Contains(t, c.Tags, "generated")

		// Check output/error consistency
		if c.Status {
			assert.NotEmpty(t, c.Output)
			assert.Empty(t, c.Error)
		} else {
			assert.Empty(t, c.Output)
			assert.NotEmpty(t, c.Error)
		}
	}

	// Verify uniqueness of IDs
	ids := make(map[string]bool)
	for _, c := range cmds {
		assert.False(t, ids[c.ID], "Duplicate ID found: %s", c.ID)
		ids[c.ID] = true
	}
}

func TestTestCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := NewCommandTestSuite(t)
	suite.MockRepo.On("Put", mock.Anything, mock.AnythingOfType("models.Command")).Return(nil)

	cmd := NewTestCommand(suite.Logger, suite.MockRepo)

	tests := []struct {
		name     string
		flags    map[string]string
		expected int
	}{
		{
			name:     "default generation",
			expected: 3,
		},
		{
			name: "custom count",
			flags: map[string]string{
				"number": "2",
			},
			expected: 2,
		},
		{
			name: "with cleanup",
			flags: map[string]string{
				"cleanup": "true",
				"number":  "1",
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite.MockRepo.Calls = nil

			if tt.flags != nil {
				for name, value := range tt.flags {
					cmd.Command().Flags().Set(name, value)
				}
			}

			err := suite.ExecuteCommand(cmd.Command())
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, len(suite.MockRepo.Calls))
		})
	}
}
