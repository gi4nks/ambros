package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gi4nks/ambros/v3/cmd/commands"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestRepo(t *testing.T) (repos.RepositoryInterface, string) {
	tempDir, err := os.MkdirTemp("", "ambros-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tempDir, "test.db")
	logger, _ := zap.NewDevelopment()
	repo, err := repos.NewRepository(dbPath, logger)
	require.NoError(t, err)

	return repo, tempDir
}

func TestEndToEndCommandExecution(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)
	defer repo.(*repos.Repository).Close()

	logger, _ := zap.NewDevelopment()
	
	// Create a command wrapper for testing
	wrapper := commands.NewCommandWrapper(logger, repo)
	
	cmdName := "echo"
	cmdArgs := []string{"Hello", "Ambros", "Integration"}
	
	// 1. Initialize command
	command := wrapper.InitializeCommand(cmdName, cmdArgs)
	assert.NotEmpty(t, command.ID)
	assert.Equal(t, cmdName, command.Name)
	
	// 2. Execute command
	wrapper.ExecuteCommand(&command)
	assert.True(t, command.Status)
	assert.Contains(t, command.Output, "Hello Ambros Integration")
	
	// 3. Finalize and Store
	err := wrapper.FinalizeCommand(&command)
	require.NoError(t, err)
	
	// 4. Verify in Repository
	stored, err := repo.Get(command.ID)
	require.NoError(t, err)
	assert.Equal(t, command.Output, stored.Output)
	assert.Equal(t, command.Name, stored.Name)
	
	// 5. Search test
	results, err := repo.GetAllCommands()
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, command.ID, results[0].ID)
}

func TestBulkCommandStress(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)
	defer repo.(*repos.Repository).Close()

	ctx := context.Background()
	count := 100
	
	for i := 0; i < count; i++ {
		cmd := models.Command{
			Entity: models.Entity{
				ID:        strings.Join([]string{"CMD", string(rune(i))}, "-"),
				CreatedAt: time.Now(),
			},
			Name: "test-cmd",
			Status: true,
		}
		err := repo.Put(ctx, cmd)
		require.NoError(t, err)
	}
	
	results, err := repo.GetAllCommands()
	require.NoError(t, err)
	assert.Len(t, results, count)
}
