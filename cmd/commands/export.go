package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
)

// ExportCommand represents the export command
type ExportCommand struct {
	*BaseCommand
	format     string
	outputFile string
	filter     string
	tag        string
	fromDate   string
	toDate     string
	history    bool
}

// NewExportCommand creates a new export command
func NewExportCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *ExportCommand {
	ec := &ExportCommand{}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export commands to file",
		Long: `Export stored commands to a file in various formats.
Supports filtering by date range, tags, and history status.
Examples:
  ambros export -o commands.json                  # Export all commands to JSON
  ambros export -o history.yaml -f yaml -H        # Export command history in YAML
  ambros export -o tagged.json -t "backup"        # Export commands with tag "backup"
  ambros export -o recent.json --from 2024-01-01  # Export commands from Jan 1st 2024`,
		RunE: ec.runE,
	}

	ec.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	ec.cmd = cmd
	ec.setupFlags(ec.cmd)
	return ec
}

func (c *ExportCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.outputFile, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&c.format, "format", "f", "json", "Output format (json or yaml)")
	cmd.Flags().StringVar(&c.filter, "filter", "", "Filter commands by status (success|failed)")
	cmd.Flags().StringVarP(&c.tag, "tag", "t", "", "Filter commands by tag")
	cmd.Flags().StringVar(&c.fromDate, "from", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&c.toDate, "to", "", "End date (YYYY-MM-DD)")
	cmd.Flags().BoolVarP(&c.history, "history", "H", false, "Export command history")

	if err := cmd.MarkFlagRequired("output"); err != nil {
		c.logger.Error("Failed to mark flag as required", zap.Error(err))
	}
}

func (c *ExportCommand) runE(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := c.validateFlags(); err != nil {
		return err
	}

	// Get commands based on filters
	commands, err := c.getFilteredCommands()
	if err != nil {
		return err
	}

	// Prepare export data
	data := c.prepareExportData(commands)

	// Export based on format
	if err := c.exportData(data); err != nil {
		return err
	}

	c.logger.Info("Export completed",
		zap.String("file", c.outputFile),
		zap.Int("commands", len(commands)),
	)

	return nil
}

func (c *ExportCommand) validateFlags() error {
	if c.format != "json" && c.format != "yaml" {
		return errors.NewError(errors.ErrInvalidCommand, "unsupported format", nil)
	}

	if c.filter != "" && c.filter != "success" && c.filter != "failed" {
		return errors.NewError(errors.ErrInvalidCommand, "invalid filter value", nil)
	}

	if c.fromDate != "" {
		if _, err := time.Parse("2006-01-02", c.fromDate); err != nil {
			return errors.NewError(errors.ErrInvalidCommand, "invalid from date format", err)
		}
	}

	if c.toDate != "" {
		if _, err := time.Parse("2006-01-02", c.toDate); err != nil {
			return errors.NewError(errors.ErrInvalidCommand, "invalid to date format", err)
		}
	}

	return nil
}

func (c *ExportCommand) getFilteredCommands() ([]models.Command, error) {
	var commands []models.Command
	var err error

	if c.tag != "" {
		commands, err = c.repository.SearchByTag(c.tag)
	} else if c.filter != "" {
		success := c.filter == "success"
		commands, err = c.repository.SearchByStatus(success)
	} else {
		if c.history {
			commands, err = c.repository.GetAllCommands()
		} else {
			commands, err = c.repository.GetAllCommands()
		}
	}

	if err != nil {
		return nil, errors.NewError(errors.ErrRepositoryRead, "failed to retrieve commands", err)
	}

	// Apply date filters if specified
	if c.fromDate != "" || c.toDate != "" {
		commands = c.filterByDate(commands)
	}

	return commands, nil
}

func (c *ExportCommand) filterByDate(commands []models.Command) []models.Command {
	var fromTime, toTime time.Time
	var err error

	if c.fromDate != "" {
		fromTime, err = time.Parse("2006-01-02", c.fromDate)
		if err != nil {
			return commands
		}
	}

	if c.toDate != "" {
		toTime, err = time.Parse("2006-01-02", c.toDate)
		if err != nil {
			return commands
		}
		toTime = toTime.Add(24 * time.Hour) // Include the entire day
	}

	filtered := make([]models.Command, 0)
	for _, cmd := range commands {
		if c.fromDate != "" && cmd.CreatedAt.Before(fromTime) {
			continue
		}
		if c.toDate != "" && cmd.CreatedAt.After(toTime) {
			continue
		}
		filtered = append(filtered, cmd)
	}

	return filtered
}

func (c *ExportCommand) prepareExportData(commands []models.Command) interface{} {
	return struct {
		ExportDate time.Time        `json:"export_date" yaml:"export_date"`
		Commands   []models.Command `json:"commands" yaml:"commands"`
		Metadata   struct {
			Total    int    `json:"total" yaml:"total"`
			Format   string `json:"format" yaml:"format"`
			Filter   string `json:"filter,omitempty" yaml:"filter,omitempty"`
			Tag      string `json:"tag,omitempty" yaml:"tag,omitempty"`
			FromDate string `json:"from_date,omitempty" yaml:"from_date,omitempty"`
			ToDate   string `json:"to_date,omitempty" yaml:"to_date,omitempty"`
			History  bool   `json:"history" yaml:"history"`
		} `json:"metadata" yaml:"metadata"`
	}{
		ExportDate: time.Now(),
		Commands:   commands,
		Metadata: struct {
			Total    int    `json:"total" yaml:"total"`
			Format   string `json:"format" yaml:"format"`
			Filter   string `json:"filter,omitempty" yaml:"filter,omitempty"`
			Tag      string `json:"tag,omitempty" yaml:"tag,omitempty"`
			FromDate string `json:"from_date,omitempty" yaml:"from_date,omitempty"`
			ToDate   string `json:"to_date,omitempty" yaml:"to_date,omitempty"`
			History  bool   `json:"history" yaml:"history"`
		}{
			Total:    len(commands),
			Format:   c.format,
			Filter:   c.filter,
			Tag:      c.tag,
			FromDate: c.fromDate,
			ToDate:   c.toDate,
			History:  c.history,
		},
	}
}

func (c *ExportCommand) exportData(data interface{}) error {
	var exportedData []byte
	var err error

	switch c.format {
	case "json":
		exportedData, err = json.MarshalIndent(data, "", "  ")
	case "yaml":
		exportedData, err = yaml.Marshal(data)
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unsupported format", nil)
	}

	if err != nil {
		return errors.NewError(errors.ErrInternalServer, "failed to marshal data", err)
	}

	dir := filepath.Dir(c.outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewError(errors.ErrInternalServer, "failed to create output directory", err)
	}

	if err := os.WriteFile(c.outputFile, exportedData, 0644); err != nil {
		return errors.NewError(errors.ErrInternalServer, "failed to write output file", err)
	}

	return nil
}

func (c *ExportCommand) Command() *cobra.Command {
	return c.cmd
}
