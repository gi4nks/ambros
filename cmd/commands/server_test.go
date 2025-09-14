package commands

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/gi4nks/ambros/v3/internal/models"
)

// MockServerRepository for testing server command
type MockServerRepository struct {
	mock.Mock
}

func (m *MockServerRepository) Put(ctx context.Context, command models.Command) error {
	args := m.Called(ctx, command)
	return args.Error(0)
}

func (m *MockServerRepository) Get(id string) (*models.Command, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Command), args.Error(1)
}

func (m *MockServerRepository) FindById(id string) (models.Command, error) {
	args := m.Called(id)
	return args.Get(0).(models.Command), args.Error(1)
}

func (m *MockServerRepository) GetLimitCommands(limit int) ([]models.Command, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Command), args.Error(1)
}

func (m *MockServerRepository) GetAllCommands() ([]models.Command, error) {
	args := m.Called()
	return args.Get(0).([]models.Command), args.Error(1)
}

func (m *MockServerRepository) SearchByTag(tag string) ([]models.Command, error) {
	args := m.Called(tag)
	return args.Get(0).([]models.Command), args.Error(1)
}

func (m *MockServerRepository) SearchByStatus(success bool) ([]models.Command, error) {
	args := m.Called(success)
	return args.Get(0).([]models.Command), args.Error(1)
}

func (m *MockServerRepository) GetTemplate(name string) (*models.Template, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockServerRepository) Push(command models.Command) error {
	args := m.Called(command)
	return args.Error(0)
}

func (m *MockServerRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestNewServerCommand(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}

	cmd := NewServerCommand(logger, mockRepo)

	assert.NotNil(t, cmd)
	assert.Equal(t, "server", cmd.cmd.Use)
	assert.Contains(t, cmd.cmd.Short, "Ambros web dashboard")
	assert.NotNil(t, cmd.BaseCommand)
	assert.Equal(t, logger, cmd.logger)
}

func TestServerCommand_SetupFlags(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}

	cmd := NewServerCommand(logger, mockRepo)

	// Test that flags are properly set up
	assert.True(t, cmd.cmd.Flags().HasAvailableFlags())

	// Test port flag
	portFlag := cmd.cmd.Flags().Lookup("port")
	assert.NotNil(t, portFlag)
	assert.Equal(t, "8080", portFlag.DefValue)

	// Test host flag
	hostFlag := cmd.cmd.Flags().Lookup("host")
	assert.NotNil(t, hostFlag)
	assert.Equal(t, "localhost", hostFlag.DefValue)

	// Test dev flag
	devFlag := cmd.cmd.Flags().Lookup("dev")
	assert.NotNil(t, devFlag)
	assert.Equal(t, "false", devFlag.DefValue)

	// Test cors flag
	corsFlag := cmd.cmd.Flags().Lookup("cors")
	assert.NotNil(t, corsFlag)
	assert.Equal(t, "false", corsFlag.DefValue)

	// Test verbose flag
	verboseFlag := cmd.cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestServerCommand_GetMostUsedCommands(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	// Test data with proper model structure
	commands := []models.Command{
		{
			Entity: models.Entity{
				ID:        "cmd-1",
				CreatedAt: time.Now(),
			},
			Command: "echo hello",
			Tags:    []string{"test"},
			Status:  true,
		},
		{
			Entity: models.Entity{
				ID:        "cmd-2",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			Command: "ls -la",
			Tags:    []string{"filesystem"},
			Status:  false,
		},
		{
			Entity: models.Entity{
				ID:        "cmd-3",
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
			Command: "echo hello",
			Tags:    []string{"test"},
			Status:  true,
		},
	}

	result := cmd.getMostUsedCommands(commands, 7)

	// Should return most used commands
	assert.NotNil(t, result)
	assert.IsType(t, []string{}, result)

	// Ensure the result strings include a count like "name (N)"
	for _, v := range result {
		assert.Contains(t, v, "(")
		assert.Contains(t, v, ")")
	}
}

func TestServerCommand_CalculateAverageExecutionTime(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	// Test with empty commands
	emptyResult := cmd.calculateAverageExecutionTime([]models.Command{})
	assert.Equal(t, time.Duration(0), emptyResult)

	// Test with commands
	commands := []models.Command{
		{
			Entity: models.Entity{
				ID:           "cmd-1",
				CreatedAt:    time.Now(),
				TerminatedAt: time.Now().Add(100 * time.Millisecond),
			},
			Command: "echo hello",
			Status:  true,
		},
		{
			Entity: models.Entity{
				ID:           "cmd-2",
				CreatedAt:    time.Now(),
				TerminatedAt: time.Now().Add(200 * time.Millisecond),
			},
			Command: "ls -la",
			Status:  false,
		},
	}

	result := cmd.calculateAverageExecutionTime(commands)
	assert.Greater(t, result, time.Duration(0))
}

func TestServerCommand_AnalyzeCommandPatterns(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	commands := []models.Command{
		{
			Entity: models.Entity{
				ID: "cmd-1",
			},
			Command: "echo hello",
			Status:  true,
		},
		{
			Entity: models.Entity{
				ID: "cmd-2",
			},
			Command: "ls -la",
			Status:  false,
		},
	}

	result := cmd.analyzeCommandPatterns(commands)

	assert.NotNil(t, result)
	assert.IsType(t, CommandPatterns{}, result)
}

func TestServerCommand_AnalyzeExecutionTrends(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	commands := []models.Command{
		{
			Entity: models.Entity{
				ID:        "cmd-1",
				CreatedAt: time.Now(),
			},
			Command: "echo hello",
			Status:  true,
		},
	}

	result := cmd.analyzeExecutionTrends(commands)

	assert.NotNil(t, result)
	assert.IsType(t, ExecutionTrends{}, result)
}

func TestServerCommand_AnalyzeFailures(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	commands := []models.Command{
		{
			Entity: models.Entity{
				ID: "cmd-1",
			},
			Command: "echo hello",
			Status:  true,
		},
		{
			Entity: models.Entity{
				ID: "cmd-2",
			},
			Command: "false", // Command that fails
			Status:  false,
		},
	}

	result := cmd.analyzeFailures(commands)

	assert.NotNil(t, result)
	assert.IsType(t, FailureAnalysis{}, result)
}

func TestServerCommand_Command(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}

	cmd := NewServerCommand(logger, mockRepo)
	cobraCmd := cmd.Command()

	assert.NotNil(t, cobraCmd)
	assert.Equal(t, "server", cobraCmd.Use)
	assert.Contains(t, cobraCmd.Short, "dashboard")
}

// Benchmark tests
func BenchmarkServerCommand_GetMostUsedCommands(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	// Create test data
	commands := make([]models.Command, 1000)
	for i := 0; i < 1000; i++ {
		commands[i] = models.Command{
			Entity: models.Entity{
				ID:        "cmd-" + string(rune(i)),
				CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			},
			Command: "test command " + string(rune(i%10)),
			Status:  i%2 == 0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.getMostUsedCommands(commands, 30)
	}
}

func BenchmarkServerCommand_CalculateAverageExecutionTime(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	// Create test data
	commands := make([]models.Command, 1000)
	for i := 0; i < 1000; i++ {
		commands[i] = models.Command{
			Entity: models.Entity{
				ID:           "cmd-" + string(rune(i)),
				CreatedAt:    time.Now(),
				TerminatedAt: time.Now().Add(time.Duration(i) * time.Millisecond),
			},
			Command: "test command",
			Status:  true,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.calculateAverageExecutionTime(commands)
	}
}
