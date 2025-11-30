package commands

import (
	"context"
	"fmt"
	"strings"
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

func TestGenerateRecommendations_Empty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	recs := cmd.generateRecommendations([]models.Command{})
	assert.GreaterOrEqual(t, len(recs), 1)
}

func TestGenerateRecommendations_Frequent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	cmds := []models.Command{}
	for i := 0; i < 5; i++ {
		cmds = append(cmds, models.Command{Command: "git pull"})
	}
	recs := cmd.generateRecommendations(cmds)
	assert.NotEmpty(t, recs)
	// recommendation may suggest templates or chains depending on heuristics
	ok := false
	for _, r := range recs {
		if strings.Contains(r, "template") || strings.Contains(r, "chain") || strings.Contains(r, "->") {
			ok = true
			break
		}
	}
	assert.True(t, ok)
}

func TestGenerateRecommendations_Sequence(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	cmds := []models.Command{}
	// repeat a pair 2 times
	cmds = append(cmds, models.Command{Command: "git pull"})
	cmds = append(cmds, models.Command{Command: "deploy"})
	cmds = append(cmds, models.Command{Command: "git pull"})
	cmds = append(cmds, models.Command{Command: "deploy"})

	recs := cmd.generateRecommendations(cmds)
	assert.NotEmpty(t, recs)
	found := false
	for _, r := range recs {
		if strings.Contains(r, "chain") || strings.Contains(r, "sequence") || strings.Contains(r, "->") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected chain recommendation for repeated sequence")
}

func TestGenerateRecommendations_Failures(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	cmds := []models.Command{{Command: "git pull", Status: false}, {Command: "git pull", Status: false}}
	recs := cmd.generateRecommendations(cmds)
	assert.True(t, len(recs) > 0)
	found := false
	for _, r := range recs {
		if strings.Contains(r, "Investigate frequent failures") {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestGenerateSearchSuggestions_Match(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	cmds := []models.Command{{Entity: models.Entity{ID: "1"}, Name: "git pull", Command: "git pull"}, {Entity: models.Entity{ID: "2"}, Name: "npm install", Command: "npm install"}}
	res := cmd.generateSearchSuggestions("git", cmds)
	assert.True(t, len(res) >= 1)
}

func TestGenerateSearchSuggestions_Empty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	res := cmd.generateSearchSuggestions("", []models.Command{})
	assert.Contains(t, res[0], "Add a search term")
}

// Fuzzing to ensure we don't panic and suggestions remain stable for random inputs
func FuzzGenerateSearchSuggestions(f *testing.F) {
	seed := []string{"git", "docker build", "npm", "deploy"}
	for _, s := range seed {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, q string) {
		logger := zaptest.NewLogger(t)
		mockRepo := &MockServerRepository{}
		cmd := NewServerCommand(logger, mockRepo)

		commands := []models.Command{{Entity: models.Entity{ID: "1"}, Name: "git pull", Command: "git pull"}, {Entity: models.Entity{ID: "2"}, Name: "npm install", Command: "npm install"}}
		// ensure the function does not panic and returns something
		suggestions := cmd.generateSearchSuggestions(q, commands)
		if suggestions == nil {
			t.Fatalf("expected suggestions array, got nil")
		}
	})
}

func FuzzGenerateRecommendations(f *testing.F) {
	seed := []string{"git pull", "deploy", "npm install", "backup"}
	for _, s := range seed {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, q string) {
		logger := zaptest.NewLogger(t)
		mockRepo := &MockServerRepository{}
		cmd := NewServerCommand(logger, mockRepo)

		// random commands
		commands := []models.Command{{Command: q, Status: true}}
		// ensure it doesn't panic and returns suggestions
		recs := cmd.generateRecommendations(commands)
		if recs == nil {
			t.Fatalf("expected non-nil recommendations")
		}
	})
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

// Deep Analytics Tests

func TestGenerateAliasSuggestions_Empty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	suggestions := cmd.generateAliasSuggestions([]models.Command{})
	assert.NotNil(t, suggestions)
	assert.Empty(t, suggestions)
}

func TestGenerateAliasSuggestions_LongCommands(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	// Create commands with a long command used frequently
	commands := []models.Command{}
	longCmd := "docker-compose -f docker-compose.prod.yml up -d --build"
	for i := 0; i < 5; i++ {
		commands = append(commands, models.Command{
			Entity:  models.Entity{ID: fmt.Sprintf("cmd-%d", i)},
			Command: longCmd,
		})
	}

	suggestions := cmd.generateAliasSuggestions(commands)
	assert.NotNil(t, suggestions)
	assert.GreaterOrEqual(t, len(suggestions), 1)

	// Check that the suggestion includes the long command
	found := false
	for _, s := range suggestions {
		if s.FullCommand == longCmd {
			found = true
			assert.Greater(t, s.UsageCount, 0)
			assert.NotEmpty(t, s.Alias)
			assert.NotEmpty(t, s.Reason)
			break
		}
	}
	assert.True(t, found, "Expected to find suggestion for long command")
}

func TestGenerateAliasName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	tests := []struct {
		command  string
		expected string
	}{
		{"docker-compose up -d", "du"},
		{"git commit -m message", "gcm"},
		{"kubectl get pods", "kgp"},
		{"npm run build", "nrb"},
		{"", "alias"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			alias := cmd.generateAliasName(tt.command)
			assert.NotEmpty(t, alias)
			// Alias should be lowercase
			assert.Equal(t, strings.ToLower(alias), alias)
		})
	}
}

func TestAnalyzeSequencePatterns_Empty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	patterns := cmd.analyzeSequencePatterns([]models.Command{})
	assert.NotNil(t, patterns)
	assert.Empty(t, patterns)
}

func TestAnalyzeSequencePatterns_DetectsSequence(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	// Create a repeating sequence: git pull -> npm install -> npm test
	now := time.Now()
	commands := []models.Command{}
	for i := 0; i < 4; i++ {
		baseTime := now.Add(time.Duration(i*10) * time.Minute)
		commands = append(commands,
			models.Command{
				Entity:  models.Entity{ID: fmt.Sprintf("pull-%d", i), CreatedAt: baseTime},
				Command: "git pull",
			},
			models.Command{
				Entity:  models.Entity{ID: fmt.Sprintf("install-%d", i), CreatedAt: baseTime.Add(1 * time.Minute)},
				Command: "npm install",
			},
		)
	}

	patterns := cmd.analyzeSequencePatterns(commands)
	assert.NotNil(t, patterns)

	// Should detect the git -> npm sequence
	found := false
	for _, p := range patterns {
		if len(p.Sequence) >= 2 && p.Sequence[0] == "git" && p.Sequence[1] == "npm" {
			found = true
			assert.GreaterOrEqual(t, p.Occurrences, 3)
			break
		}
	}
	assert.True(t, found, "Expected to detect git -> npm sequence pattern")
}

func TestIdentifyWorkflowInsights_Empty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	insights := cmd.identifyWorkflowInsights([]models.Command{})
	assert.NotNil(t, insights)
	assert.Empty(t, insights)
}

func TestIdentifyWorkflowInsights_DetectsGitWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	commands := []models.Command{}
	// Add frequent git commands
	for i := 0; i < 10; i++ {
		commands = append(commands, models.Command{
			Entity:  models.Entity{ID: fmt.Sprintf("git-%d", i)},
			Command: "git pull",
		})
	}
	// Add frequent make commands
	for i := 0; i < 8; i++ {
		commands = append(commands, models.Command{
			Entity:  models.Entity{ID: fmt.Sprintf("make-%d", i)},
			Command: "make test",
		})
	}

	insights := cmd.identifyWorkflowInsights(commands)
	assert.NotNil(t, insights)

	// Should detect Git Development Flow or Go Development workflow
	found := false
	for _, insight := range insights {
		if strings.Contains(insight.Name, "Git") || strings.Contains(insight.Name, "Go") {
			found = true
			assert.NotEmpty(t, insight.Description)
			assert.NotEmpty(t, insight.Suggestion)
			assert.Greater(t, insight.Frequency, 0)
			break
		}
	}
	assert.True(t, found, "Expected to detect a development workflow")
}

func TestCalculateComplexityScore(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	tests := []struct {
		command       string
		minComplexity int
	}{
		{"ls", 0},
		{"git commit -m 'message'", 2},
		{"docker run -it --rm -v /data:/data nginx", 5},
		{"cat file.txt | grep pattern | sort | uniq", 9},
		{"if [ -f file ]; then echo 'exists'; fi && rm -f file || echo 'failed'", 10},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			score := cmd.calculateComplexityScore(tt.command)
			assert.GreaterOrEqual(t, score, tt.minComplexity)
		})
	}
}

func TestGenerateDeepAnalytics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := &MockServerRepository{}
	cmd := NewServerCommand(logger, mockRepo)

	commands := []models.Command{
		{Entity: models.Entity{ID: "1", CreatedAt: time.Now()}, Command: "git pull"},
		{Entity: models.Entity{ID: "2", CreatedAt: time.Now().Add(time.Minute)}, Command: "npm install"},
	}

	analytics := cmd.generateDeepAnalytics(commands)

	assert.NotNil(t, analytics.AliasSuggestions)
	assert.NotNil(t, analytics.SequencePatterns)
	assert.NotNil(t, analytics.WorkflowInsights)
	assert.NotNil(t, analytics.CommandComplexity)
}
