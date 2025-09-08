package repos_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type testFixture struct {
	repo       *repos.Repository
	logger     *zap.Logger
	testDBPath string
}

func setupTest(t *testing.T) *testFixture {
	logger := zaptest.NewLogger(t)
	testDBPath := filepath.Join(t.TempDir(), "test.db")

	repo, err := repos.NewRepository(testDBPath, logger)
	require.NoError(t, err)

	return &testFixture{
		repo:       repo,
		logger:     logger,
		testDBPath: testDBPath,
	}
}

func (f *testFixture) tearDown() {
	f.repo.Close()
	os.RemoveAll(filepath.Dir(f.testDBPath))
}

func TestPutAndGet(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	ctx := context.Background()

	tests := []struct {
		name    string
		cmd     models.Command
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: models.Command{
				Entity: models.Entity{
					ID:           "test1",
					CreatedAt:    time.Now(),
					TerminatedAt: time.Now(),
				},
				Name:      "test",
				Arguments: []string{"arg1", "arg2"},
				Status:    true,
			},
			wantErr: false,
		},
		{
			name: "command with tags",
			cmd: models.Command{
				Entity: models.Entity{
					ID:           "test2",
					CreatedAt:    time.Now(),
					TerminatedAt: time.Now(),
				},
				Name:      "tagged-cmd",
				Arguments: []string{"arg1"},
				Tags:      []string{"test", "demo"},
				Category:  "testing",
				Status:    true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.repo.Put(ctx, tt.cmd)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Test Get method
			found, err := f.repo.Get(tt.cmd.ID)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, tt.cmd.ID, found.ID)
			assert.Equal(t, tt.cmd.Name, found.Name)
			assert.Equal(t, tt.cmd.Arguments, found.Arguments)
			assert.Equal(t, tt.cmd.Tags, found.Tags)
			assert.Equal(t, tt.cmd.Category, found.Category)

			// Test FindById method (legacy)
			foundLegacy, err := f.repo.FindById(tt.cmd.ID)
			assert.NoError(t, err)
			assert.Equal(t, tt.cmd.ID, foundLegacy.ID)
			assert.Equal(t, tt.cmd.Name, foundLegacy.Name)
		})
	}
}

func TestGetLimitCommands(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	ctx := context.Background()

	// Insert test commands with different timestamps
	baseTime := time.Now()
	commands := []models.Command{
		{
			Entity: models.Entity{
				ID:           "1",
				CreatedAt:    baseTime.Add(-2 * time.Hour),
				TerminatedAt: baseTime.Add(-2 * time.Hour),
			},
			Name: "cmd1",
		},
		{
			Entity: models.Entity{
				ID:           "2",
				CreatedAt:    baseTime.Add(-1 * time.Hour),
				TerminatedAt: baseTime.Add(-1 * time.Hour),
			},
			Name: "cmd2",
		},
		{
			Entity: models.Entity{
				ID:           "3",
				CreatedAt:    baseTime,
				TerminatedAt: baseTime,
			},
			Name: "cmd3",
		},
	}

	for _, cmd := range commands {
		require.NoError(t, f.repo.Put(ctx, cmd))
	}

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{"get all commands", 3, 3},
		{"get single command", 1, 1},
		{"get zero commands", 0, 0},
		{"get more than existing", 5, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := f.repo.GetLimitCommands(tt.limit)
			assert.NoError(t, err)
			assert.Len(t, results, tt.wantCount)

			// Verify ordering (most recent first)
			if len(results) > 1 {
				assert.True(t, results[0].CreatedAt.After(results[1].CreatedAt) ||
					results[0].CreatedAt.Equal(results[1].CreatedAt))
			}
		})
	}
}

func TestSearchByTag(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	ctx := context.Background()

	// Insert test commands with tags
	commands := []models.Command{
		{
			Entity: models.Entity{ID: "1", CreatedAt: time.Now()},
			Name:   "cmd1",
			Tags:   []string{"test", "demo"},
		},
		{
			Entity: models.Entity{ID: "2", CreatedAt: time.Now()},
			Name:   "cmd2",
			Tags:   []string{"prod", "deploy"},
		},
		{
			Entity: models.Entity{ID: "3", CreatedAt: time.Now()},
			Name:   "cmd3",
			Tags:   []string{"TEST", "debug"}, // Test case insensitive
		},
	}

	for _, cmd := range commands {
		require.NoError(t, f.repo.Put(ctx, cmd))
	}

	tests := []struct {
		name      string
		tag       string
		wantCount int
		wantIDs   []string
	}{
		{"find test tag", "test", 2, []string{"1", "3"}},
		{"find prod tag", "prod", 1, []string{"2"}},
		{"find non-existent tag", "nonexistent", 0, []string{}},
		{"case insensitive search", "TEST", 2, []string{"1", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := f.repo.SearchByTag(tt.tag)
			assert.NoError(t, err)
			assert.Len(t, results, tt.wantCount)

			if tt.wantCount > 0 {
				foundIDs := make([]string, len(results))
				for i, cmd := range results {
					foundIDs[i] = cmd.ID
				}
				assert.ElementsMatch(t, tt.wantIDs, foundIDs)
			}
		})
	}
}

func TestSearchByStatus(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	ctx := context.Background()

	// Insert test commands with different statuses
	commands := []models.Command{
		{
			Entity: models.Entity{ID: "success1", CreatedAt: time.Now()},
			Name:   "cmd1",
			Status: true,
		},
		{
			Entity: models.Entity{ID: "failed1", CreatedAt: time.Now()},
			Name:   "cmd2",
			Status: false,
		},
		{
			Entity: models.Entity{ID: "success2", CreatedAt: time.Now()},
			Name:   "cmd3",
			Status: true,
		},
	}

	for _, cmd := range commands {
		require.NoError(t, f.repo.Put(ctx, cmd))
	}

	tests := []struct {
		name      string
		success   bool
		wantCount int
	}{
		{"find successful commands", true, 2},
		{"find failed commands", false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := f.repo.SearchByStatus(tt.success)
			assert.NoError(t, err)
			assert.Len(t, results, tt.wantCount)

			for _, cmd := range results {
				assert.Equal(t, tt.success, cmd.Status)
			}
		})
	}
}

func TestPushAndStoredCommands(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	cmd := models.Command{
		Entity: models.Entity{
			ID:           "push-test",
			CreatedAt:    time.Now(),
			TerminatedAt: time.Now(),
		},
		Name: "test-push",
	}

	// Test pushing command
	err := f.repo.Push(cmd)
	assert.NoError(t, err)

	// Test retrieving stored command
	stored, err := f.repo.FindInStoreById(cmd.ID)
	assert.NoError(t, err)
	assert.Equal(t, cmd.ID, stored.ID)
	assert.Equal(t, cmd.Name, stored.Name)

	// Test getting all stored commands
	allStored, err := f.repo.GetAllStoredCommands()
	assert.NoError(t, err)
	assert.Len(t, allStored, 1)
	assert.Equal(t, cmd.ID, allStored[0].ID)
}

func TestDeleteStoredCommands(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	// Store multiple commands
	commands := []models.Command{
		{Entity: models.Entity{ID: "stored1"}, Name: "cmd1"},
		{Entity: models.Entity{ID: "stored2"}, Name: "cmd2"},
	}

	for _, cmd := range commands {
		require.NoError(t, f.repo.Push(cmd))
	}

	// Test deleting single stored command
	err := f.repo.DeleteStoredCommand("stored1")
	assert.NoError(t, err)

	// Verify it was deleted
	_, err = f.repo.FindInStoreById("stored1")
	assert.Error(t, err)

	// Verify other command still exists
	_, err = f.repo.FindInStoreById("stored2")
	assert.NoError(t, err)

	// Test deleting all stored commands
	err = f.repo.DeleteAllStoredCommands()
	assert.NoError(t, err)

	// Verify all are deleted
	stored, err := f.repo.GetAllStoredCommands()
	assert.NoError(t, err)
	assert.Len(t, stored, 0)
}

func TestGetAllCommands(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	ctx := context.Background()

	// Add test commands
	expectedCmds := []models.Command{
		{
			Entity: models.Entity{ID: "test-1", CreatedAt: time.Now()},
			Name:   "cmd1",
		},
		{
			Entity: models.Entity{ID: "test-2", CreatedAt: time.Now()},
			Name:   "cmd2",
		},
	}

	for _, cmd := range expectedCmds {
		require.NoError(t, f.repo.Put(ctx, cmd))
	}

	// Test getting all commands
	commands, err := f.repo.GetAllCommands()
	assert.NoError(t, err)
	assert.Len(t, commands, len(expectedCmds))

	// Extract IDs for comparison
	foundIDs := make([]string, len(commands))
	expectedIDs := make([]string, len(expectedCmds))
	for i, cmd := range commands {
		foundIDs[i] = cmd.ID
	}
	for i, cmd := range expectedCmds {
		expectedIDs[i] = cmd.ID
	}
	assert.ElementsMatch(t, expectedIDs, foundIDs)
}

func TestGetExecutedCommands(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	ctx := context.Background()

	// Add test commands
	for i := 0; i < 3; i++ {
		cmd := models.Command{
			Entity: models.Entity{
				ID:        fmt.Sprintf("test-%d", i),
				CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
			},
			Name:   fmt.Sprintf("cmd%d", i),
			Status: i%2 == 0,
		}
		require.NoError(t, f.repo.Put(ctx, cmd))
	}

	// Test getting executed commands
	executed, err := f.repo.GetExecutedCommands(2)
	assert.NoError(t, err)
	assert.Len(t, executed, 2)

	// Verify order field is set
	for i, exec := range executed {
		assert.Equal(t, i, exec.Order)
		assert.NotEmpty(t, exec.ID)
		assert.NotEmpty(t, exec.Command)
	}
}

func TestBackupSchema(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	err := f.repo.BackupSchema()
	assert.NoError(t, err)

	// Verify backup file exists
	backupFile := f.testDBPath + ".bkp"
	_, err = os.Stat(backupFile)
	assert.NoError(t, err)

	// Cleanup backup file
	os.Remove(backupFile)
}

func TestDeleteSchema(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	ctx := context.Background()

	// Add some test data
	cmd := models.Command{
		Entity: models.Entity{ID: "test", CreatedAt: time.Now()},
		Name:   "test",
	}
	require.NoError(t, f.repo.Put(ctx, cmd))
	require.NoError(t, f.repo.Push(cmd))

	// Verify data exists
	commands, err := f.repo.GetAllCommands()
	require.NoError(t, err)
	assert.Len(t, commands, 1)

	// Test schema deletion
	err = f.repo.DeleteSchema(true)
	assert.NoError(t, err)

	// Verify data is deleted
	commands, err = f.repo.GetAllCommands()
	assert.NoError(t, err)
	assert.Len(t, commands, 0)
}

func TestGetTemplate(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	// Test template retrieval (placeholder implementation)
	template, err := f.repo.GetTemplate("test-template")
	assert.Error(t, err)
	assert.Nil(t, template)
	assert.Contains(t, err.Error(), "template not found")
}

func TestNotFoundErrors(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()

	// Test Get non-existent command
	cmd, err := f.repo.Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, cmd)

	// Test FindById non-existent command
	_, err = f.repo.FindById("nonexistent")
	assert.Error(t, err)

	// Test FindInStoreById non-existent command
	_, err = f.repo.FindInStoreById("nonexistent")
	assert.Error(t, err)
}
