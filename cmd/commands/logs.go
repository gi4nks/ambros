package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gi4nks/ambros/v3/internal/plugins"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// LogsCommand represents the logs command
type LogsCommand struct {
	*BaseCommand
	follow bool
	limit  int
}

// NewLogsCommand creates a new logs command
func NewLogsCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *LogsCommand {
	lc := &LogsCommand{}

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Display internal Ambros logs",
		Long:  `View, filter, or follow the internal logs for debugging and diagnostic purposes.`,
		RunE:  lc.runE,
	}

	lc.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	lc.cmd = cmd
	lc.setupFlags(cmd)
	return lc
}

func (lc *LogsCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&lc.follow, "follow", "f", false, "Follow log output (tail -f)")
	cmd.Flags().IntVarP(&lc.limit, "limit", "n", 50, "Number of lines to show")
}

func (lc *LogsCommand) runE(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".ambros", "ambros.log")

	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No log file found.")
			return nil
		}
		return err
	}
	defer file.Close()

	if lc.follow {
		fmt.Printf("Following logs from: %s\n", logPath)
		fmt.Println("Press Ctrl+C to stop.")

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		return nil
	}

	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > lc.limit {
			lines = lines[1:]
		}
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}

func (lc *LogsCommand) Command() *cobra.Command {
	return lc.cmd
}
