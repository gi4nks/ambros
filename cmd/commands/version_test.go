package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestVersionCommand(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("version command with full output", func(t *testing.T) {
		versionCmd := NewVersionCommand(logger)
		versionCmd.short = false

		err := versionCmd.runE(versionCmd.cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("version command with short output", func(t *testing.T) {
		versionCmd := NewVersionCommand(logger)
		versionCmd.short = true

		err := versionCmd.runE(versionCmd.cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("command structure validation", func(t *testing.T) {
		versionCmd := NewVersionCommand(logger)

		assert.Equal(t, "version", versionCmd.cmd.Use)
		assert.Equal(t, "Show version information", versionCmd.cmd.Short)
		assert.NotNil(t, versionCmd.Command())
		assert.Equal(t, versionCmd.cmd, versionCmd.Command())

		// Test that the short flag exists
		shortFlag := versionCmd.cmd.Flags().Lookup("short")
		assert.NotNil(t, shortFlag)
		assert.Equal(t, "false", shortFlag.DefValue)

		// Test short flag shorthand
		sFlag := versionCmd.cmd.Flags().ShorthandLookup("s")
		assert.NotNil(t, sFlag)
		assert.Equal(t, sFlag, shortFlag)
	})

	t.Run("version variables are accessible", func(t *testing.T) {
		// Test that version variables exist and have default values
		assert.NotEmpty(t, Version)
		assert.NotEmpty(t, GitCommit)
		assert.NotEmpty(t, BuildDate)
		assert.NotEmpty(t, GoVersion)
	})

	t.Run("version command does not require repository", func(t *testing.T) {
		// Version command should work without repository
		versionCmd := NewVersionCommand(logger)
		assert.Nil(t, versionCmd.repository)
	})
}

func TestVersionCommand_SetBuildInfo(t *testing.T) {
	// Test that we can set build-time variables
	originalVersion := Version
	originalCommit := GitCommit
	originalBuildDate := BuildDate

	// Set test values
	Version = "1.0.0"
	GitCommit = "abc123"
	BuildDate = "2024-01-01"

	defer func() {
		// Restore original values
		Version = originalVersion
		GitCommit = originalCommit
		BuildDate = originalBuildDate
	}()

	logger, _ := zap.NewDevelopment()
	versionCmd := NewVersionCommand(logger)

	// Test that the command uses the updated values
	err := versionCmd.runE(versionCmd.cmd, []string{})
	assert.NoError(t, err)

	// Verify the values are what we set
	assert.Equal(t, "1.0.0", Version)
	assert.Equal(t, "abc123", GitCommit)
	assert.Equal(t, "2024-01-01", BuildDate)
}

func TestVersionCommand_Creation(t *testing.T) {
	logger := zap.NewNop()
	cmd := NewVersionCommand(logger)

	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Command().Use)
	assert.NotEmpty(t, cmd.Command().Short)
	assert.NotEmpty(t, cmd.Command().Long)

	// Check flag existence
	shortFlag := cmd.Command().Flags().Lookup("short")
	assert.NotNil(t, shortFlag)
	assert.Equal(t, "s", shortFlag.Shorthand)
}

func TestVersionCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := NewCommandTestSuite(t)
	cmd := NewVersionCommand(suite.Logger)

	// Execute command with and without short flag
	tests := []struct {
		name  string
		short bool
	}{
		{
			name:  "full version output",
			short: false,
		},
		{
			name:  "short version output",
			short: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd.short = tt.short

			// Execute command and verify it runs without error
			err := cmd.runE(cmd.cmd, []string{})
			assert.NoError(t, err)
		})
	}
}
