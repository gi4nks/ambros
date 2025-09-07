package analytics

import (
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/models"
)

// Analytics provides analytics functionality
type Analytics struct {
	logger *zap.Logger
}

// NewAnalytics creates a new analytics instance
func NewAnalytics(logger *zap.Logger) *Analytics {
	return &Analytics{
		logger: logger,
	}
}

// AnalyzeCommands analyzes command execution patterns
func (a *Analytics) AnalyzeCommands(commands []models.Command) *AnalyticsReport {
	report := &AnalyticsReport{
		TotalCommands:      len(commands),
		SuccessfulCommands: 0,
		FailedCommands:     0,
		TopCommands:        make(map[string]int),
	}

	for _, cmd := range commands {
		if cmd.Status {
			report.SuccessfulCommands++
		} else {
			report.FailedCommands++
		}

		// Count command usage
		report.TopCommands[cmd.Name]++
	}

	if report.TotalCommands > 0 {
		report.SuccessRate = float64(report.SuccessfulCommands) / float64(report.TotalCommands) * 100
	}

	return report
}

// AnalyticsReport contains analytics results
type AnalyticsReport struct {
	TotalCommands      int            `json:"total_commands"`
	SuccessfulCommands int            `json:"successful_commands"`
	FailedCommands     int            `json:"failed_commands"`
	SuccessRate        float64        `json:"success_rate"`
	TopCommands        map[string]int `json:"top_commands"`
}
