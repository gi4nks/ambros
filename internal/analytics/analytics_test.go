package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/models"
)

func TestAnalytics_AnalyzeCommands(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	analytics := NewAnalytics(logger)

	commands := []models.Command{
		{Name: "ls", Status: true},
		{Name: "ls", Status: true},
		{Name: "grep", Status: true},
		{Name: "rm", Status: false},
		{Name: "ls", Status: false},
	}

	report := analytics.AnalyzeCommands(commands)

	assert.Equal(t, 5, report.TotalCommands, "TotalCommands should be 5")
	assert.Equal(t, 3, report.SuccessfulCommands, "SuccessfulCommands should be 3")
	assert.Equal(t, 2, report.FailedCommands, "FailedCommands should be 2")
	assert.InDelta(t, 60.0, report.SuccessRate, 0.01, "SuccessRate should be 60.0")

	expectedTopCommands := map[string]int{
		"ls":   3,
		"grep": 1,
		"rm":   1,
	}
	assert.Equal(t, expectedTopCommands, report.TopCommands, "TopCommands should be calculated correctly")
}

func TestAnalytics_AnalyzeCommands_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	analytics := NewAnalytics(logger)

	var commands []models.Command

	report := analytics.AnalyzeCommands(commands)

	assert.Equal(t, 0, report.TotalCommands, "TotalCommands should be 0 for empty input")
	assert.Equal(t, 0, report.SuccessfulCommands, "SuccessfulCommands should be 0 for empty input")
	assert.Equal(t, 0, report.FailedCommands, "FailedCommands should be 0 for empty input")
	assert.Equal(t, 0.0, report.SuccessRate, "SuccessRate should be 0.0 for empty input")
	assert.Empty(t, report.TopCommands, "TopCommands should be empty for empty input")
}
