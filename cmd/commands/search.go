package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

type SearchCommand struct {
	*BaseCommand
	opts SearchOptions
}

type SearchOptions struct {
	tag      []string
	category string
	since    string
	status   string
	limit    int
	format   string
}

func NewSearchCommand(logger *zap.Logger, repo RepositoryInterface) *SearchCommand {
	sc := &SearchCommand{}

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search through command history",
		Long:  `Search and filter through your command history using various criteria.`,
		RunE:  sc.runE,
	}

	sc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	sc.cmd = cmd
	sc.setupFlags(cmd)
	return sc
}

func (sc *SearchCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&sc.opts.tag, "tag", "t", nil,
		"Filter by tags")
	cmd.Flags().StringVarP(&sc.opts.category, "category", "c", "",
		"Filter by category")
	cmd.Flags().StringVarP(&sc.opts.since, "since", "s", "",
		"Show commands since (e.g., 24h, 7d)")
	cmd.Flags().StringVarP(&sc.opts.status, "status", "S", "",
		"Filter by status (success/failure)")
	cmd.Flags().IntVarP(&sc.opts.limit, "limit", "l", 10,
		"Limit the number of results")
	cmd.Flags().StringVarP(&sc.opts.format, "format", "f", "text",
		"Output format (text/json/yaml)")
}

func (sc *SearchCommand) runE(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	sc.logger.Debug("Search command invoked",
		zap.String("query", query),
		zap.Strings("tags", sc.opts.tag),
		zap.String("category", sc.opts.category),
		zap.String("status", sc.opts.status),
		zap.String("since", sc.opts.since),
		zap.Int("limit", sc.opts.limit),
		zap.String("format", sc.opts.format))

	// Get all commands first
	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		sc.logger.Error("Failed to retrieve commands", zap.Error(err))
		return errors.NewError(errors.ErrRepositoryRead, "failed to retrieve commands", err)
	}

	// Apply filters
	filtered := sc.filterCommands(commands, query)

	// Apply limit
	if sc.opts.limit > 0 && len(filtered) > sc.opts.limit {
		filtered = filtered[:sc.opts.limit]
	}

	sc.logger.Info("Search completed",
		zap.String("query", query),
		zap.Int("totalCommands", len(commands)),
		zap.Int("filteredResults", len(filtered)),
		zap.String("format", sc.opts.format))

	return sc.formatOutput(filtered, sc.opts.format)
}

func (sc *SearchCommand) filterCommands(commands []models.Command, query string) []models.Command {
	var filtered []models.Command
	var sinceTime time.Time

	// Parse since duration if provided
	if sc.opts.since != "" {
		duration, err := time.ParseDuration(sc.opts.since)
		if err == nil {
			sinceTime = time.Now().Add(-duration)
		}
	}

	for _, cmd := range commands {
		// Filter by text query
		if query != "" {
			if !strings.Contains(strings.ToLower(cmd.Name), strings.ToLower(query)) &&
				!strings.Contains(strings.ToLower(strings.Join(cmd.Arguments, " ")), strings.ToLower(query)) {
				continue
			}
		}

		// Filter by tags
		if len(sc.opts.tag) > 0 {
			tagMatch := false
			for _, searchTag := range sc.opts.tag {
				for _, cmdTag := range cmd.Tags {
					if strings.EqualFold(cmdTag, searchTag) {
						tagMatch = true
						break
					}
				}
				if tagMatch {
					break
				}
			}
			if !tagMatch {
				continue
			}
		}

		// Filter by category
		if sc.opts.category != "" && !strings.EqualFold(cmd.Category, sc.opts.category) {
			continue
		}

		// Filter by status
		if sc.opts.status != "" {
			if (sc.opts.status == "success" && !cmd.Status) ||
				(sc.opts.status == "failure" && cmd.Status) {
				continue
			}
		}

		// Filter by since date
		if !sinceTime.IsZero() && cmd.CreatedAt.Before(sinceTime) {
			continue
		}

		filtered = append(filtered, cmd)
	}

	return filtered
}

func (sc *SearchCommand) formatOutput(commands []models.Command, format string) error {
	switch format {
	case "json":
		return sc.outputJSON(commands)
	case "yaml":
		return sc.outputYAML(commands)
	case "text":
		return sc.outputText(commands)
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unsupported output format", nil)
	}
}

func (sc *SearchCommand) outputJSON(commands []models.Command) error {
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return errors.NewError(errors.ErrInternalServer, "failed to marshal JSON", err)
	}
	fmt.Println(string(data))
	return nil
}

func (sc *SearchCommand) outputYAML(commands []models.Command) error {
	data, err := yaml.Marshal(commands)
	if err != nil {
		return errors.NewError(errors.ErrInternalServer, "failed to marshal YAML", err)
	}
	fmt.Println(string(data))
	return nil
}

func (sc *SearchCommand) outputText(commands []models.Command) error {
	if len(commands) == 0 {
		// User output (keep fmt)
		fmt.Println("No commands found matching the search criteria.")
		return nil
	}

	// User output (keep fmt)
	fmt.Printf("Found %d command(s):\n\n", len(commands))

	for i, cmd := range commands {
		fmt.Printf("%d. %s %s\n", i+1, cmd.Name, strings.Join(cmd.Arguments, " "))

		// Display ID with color based on status
		if cmd.Status {
			color.Green("   ID: %s", cmd.ID)
		} else {
			color.Red("   ID: %s", cmd.ID)
		}

		fmt.Printf("   Created: %s\n", cmd.CreatedAt.Format("2006-01-02 15:04:05"))

		// Display status with color
		fmt.Print("   Status: ")
		if cmd.Status {
			color.Green("Success")
		} else {
			color.Red("Failed")
		}

		if len(cmd.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(cmd.Tags, ", "))
		}

		if cmd.Category != "" {
			fmt.Printf("   Category: %s\n", cmd.Category)
		}

		if i < len(commands)-1 {
			fmt.Println()
		}
	}

	sc.logger.Debug("Text output completed",
		zap.Int("commandsDisplayed", len(commands)))

	return nil
}

func (sc *SearchCommand) Command() *cobra.Command {
	return sc.cmd
}
