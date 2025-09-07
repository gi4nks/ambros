package commands

import (
	"github.com/gi4nks/ambros/internal/repos"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type AnalyticsCommand struct {
	*BaseCommand
	opts AnalyticsOptions
}

type AnalyticsOptions struct {
	period string
	format string
	detail bool
}

func NewAnalyticsCommand(logger *zap.Logger, repo *repos.Repository) *AnalyticsCommand {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "Show command analytics",
		Long:  `Display analytics and statistics about command usage`,
	}

	ac := &AnalyticsCommand{
		BaseCommand: NewBaseCommand(cmd, logger, repo),
	}

	ac.setupFlags(cmd)
	ac.cmd = cmd
	return ac
}

func (ac *AnalyticsCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&ac.opts.period, "period", "p", "7d",
		"Analysis period (24h, 7d, 30d)")
	cmd.Flags().StringVarP(&ac.opts.format, "format", "f", "text",
		"Output format (text/json/yaml)")
	cmd.Flags().BoolVarP(&ac.opts.detail, "detail", "d", false,
		"Show detailed analytics")
}
