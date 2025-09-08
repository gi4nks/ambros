package commands

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

// TestCommand represents the test command
type TestCommand struct {
	*BaseCommand
	count   int
	cleanup bool
}

// NewTestCommand creates a new test command
func NewTestCommand(logger *zap.Logger, repo RepositoryInterface) *TestCommand {
	tc := &TestCommand{}

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Generate test commands",
		Long: `Generate and store test commands for development and testing purposes.
The commands are randomly generated with various properties and outcomes.

Examples:
  ambros test              # Generate 3 test commands
  ambros test -n 10       # Generate 10 test commands
  ambros test --cleanup   # Remove all test commands before generating new ones`,
		RunE: tc.runE,
	}

	tc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	tc.cmd = cmd

	tc.cmd.Flags().IntVarP(&tc.count, "number", "n", 3, "Number of test commands to generate")
	tc.cmd.Flags().BoolVar(&tc.cleanup, "cleanup", false, "Remove existing test commands before generating new ones")
	return tc
}

func (c *TestCommand) runE(cmd *cobra.Command, args []string) error {
	c.logger.Info("Test command invoked",
		zap.Int("count", c.count),
		zap.Bool("cleanup", c.cleanup))

	if c.cleanup {
		// In a real implementation, we would add a method to cleanup test commands
		c.logger.Debug("Cleanup requested but not implemented")
	}

	for i := 0; i < c.count; i++ {
		command := c.generateTestCommand()
		if err := c.repository.Put(context.Background(), *command); err != nil {
			return errors.NewError(errors.ErrRepositoryWrite,
				fmt.Sprintf("failed to store test command %d", i+1), err)
		}
		c.logger.Debug("Generated and stored test command",
			zap.String("commandId", command.ID),
			zap.String("name", command.Name),
			zap.Bool("status", command.Status))
	}

	c.logger.Info("Test command generation completed",
		zap.Int("generated", c.count))
	return nil
}

// generateTestCommand creates a fake command for testing
func (c *TestCommand) generateTestCommand() *models.Command {
	names := []string{"ls", "cat", "grep", "echo", "mkdir", "rm", "touch", "find"}
	args := [][]string{
		{"-l", "-a"},
		{"file.txt"},
		{"pattern", "file.txt"},
		{"Hello, World!"},
		{"test_dir"},
		{"-rf", "old_dir"},
		{"newfile.txt"},
		{".", "-name", "*.go"},
	}

	now := time.Now()
	status := rand.Float32() < 0.7 // 70% success rate
	var output, errorMsg string

	if status {
		output = fmt.Sprintf("Output from command execution at %s", now.Format(time.RFC3339))
	} else {
		errorMsg = fmt.Sprintf("Error: command failed at %s", now.Format(time.RFC3339))
	}

	execTime := time.Duration(rand.Intn(5000)) * time.Millisecond
	return &models.Command{
		Entity: models.Entity{
			ID:           fmt.Sprintf("TEST-%d", rand.Intn(10000)),
			CreatedAt:    now,
			TerminatedAt: now.Add(execTime),
		},
		Name:      names[rand.Intn(len(names))],
		Arguments: args[rand.Intn(len(args))],
		Status:    status,
		Output:    output,
		Error:     errorMsg,
		Tags:      []string{"test", "generated"},
	}
}
