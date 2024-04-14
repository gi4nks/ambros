package commands

import (
	"fmt"
	"math/rand"
	"time"

	models "github.com/gi4nks/ambros/internal/models"
	"github.com/spf13/cobra"
)

// fakeCommand generates a fake command for testing
func fakeCommand() *models.Command {
	// Create a random ID for the command
	id := fmt.Sprintf("CMD-%d", rand.Intn(1000))

	// Generate random name and arguments for the command
	names := []string{"ls", "cat", "grep", "echo", "mkdir", "rm", "touch"}
	args := [][]string{
		{"-l"},
		{"file.txt"},
		{"pattern", "file.txt"},
		{"Hello, World!"},
		{"new_directory"},
		{"-rf", "directory"},
		{"file.txt"},
	}
	name := names[rand.Intn(len(names))]
	arguments := args[rand.Intn(len(args))]

	// Generate random timestamps for CreatedAt and TerminatedAt
	createdAt := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)  // Random time within the last hour
	terminatedAt := createdAt.Add(time.Duration(rand.Intn(3600)) * time.Second) // Random time after CreatedAt

	// Generate random status, output, and error for the command
	status := rand.Float32() < 0.5 // Randomly decide if the command succeeded or failed
	output := ""
	if status {
		output = "Output of the command"
	} else {
		output = ""
	}
	errorMsg := ""
	if !status {
		errorMsg = "Error message"
	}

	// Create and return the fake command object
	return &models.Command{
		Entity: models.Entity{
			ID:           id,
			CreatedAt:    createdAt,
			TerminatedAt: terminatedAt,
		},
		Name:      name,
		Arguments: arguments,
		Status:    status,
		Output:    output,
		Error:     errorMsg,
	}
}

// testCmd represents the output command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test",
	Long:  `Test command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Info("Test command invoked")

			fakeCmd1 := fakeCommand()
			fakeCmd2 := fakeCommand()
			fakeCmd3 := fakeCommand()

			Repository.Put(*fakeCmd1)
			Repository.Put(*fakeCmd2)
			Repository.Put(*fakeCmd3)
		})
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
