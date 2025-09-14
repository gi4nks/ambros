package commands

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
)

type DBCommand struct {
	*BaseCommand
}

func NewDBCommand(logger *zap.Logger, repo RepositoryInterface) *DBCommand {
	dc := &DBCommand{}

	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database utilities (init, prune)",
		Long:  `Database helpers: initialize repository and prune history.`,
	}

	// init subcommand
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize database (noop if already initialized)",
		RunE: func(cmd *cobra.Command, args []string) error {
			wipe, _ := cmd.Flags().GetBool("wipe")
			assumeYes, _ := cmd.Flags().GetBool("yes")

			// Basic check: try to read one command to ensure DB is available
			_, err := repo.GetAllCommands()
			if err == nil && !wipe {
				logger.Info("Database is accessible and appears initialized")
				return nil
			}

			if wipe {
				if !assumeYes {
					reader := bufio.NewReader(cmd.InOrStdin())
					fmt.Fprint(cmd.OutOrStdout(), "WARNING: This will wipe the database. Type 'yes' to confirm: ")
					resp, _ := reader.ReadString('\n')
					resp = strings.TrimSpace(resp)
					if resp != "yes" {
						fmt.Fprintln(cmd.OutOrStdout(), "Aborted")
						return nil
					}
				}
				// Attempt to call concrete implementation to delete schema
				if concrete, ok := repo.(interface{ DeleteSchema(bool) error }); ok {
					if err := concrete.DeleteSchema(true); err != nil {
						logger.Error("Failed to wipe database", zap.Error(err))
						return errors.NewError(errors.ErrRepositoryWrite, "failed to wipe database", err)
					}
					logger.Info("Database wiped/reinitialized")
					return nil
				}
				return errors.NewError(errors.ErrRepositoryWrite, "repository does not support wipe/reinit", nil)
			}

			// If repo.GetAllCommands returned an error (cannot access DB), return that
			if err != nil {
				logger.Error("Database access failed", zap.Error(err))
				return errors.NewError(errors.ErrRepositoryWrite, "database access failed", err)
			}
			return nil
		},
	}
	initCmd.Flags().Bool("wipe", false, "Wipe existing database data and reinitialize (destructive)")
	initCmd.Flags().Bool("yes", false, "Assume yes for destructive operations (skip confirmation)")

	// prune subcommand
	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune command history older than a given date",
		Long:  "Prune stored command history older than the specified date. Date format: YYYY-MM-DD or RFC3339.",
		RunE: func(cmd *cobra.Command, args []string) error {
			beforeStr, _ := cmd.Flags().GetString("before")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			if beforeStr == "" {
				return errors.NewError(errors.ErrInvalidCommand, "--before is required (YYYY-MM-DD or RFC3339)", nil)
			}

			before, err := parseDate(beforeStr)
			if err != nil {
				return errors.NewError(errors.ErrInvalidCommand, "invalid date format for --before", err)
			}

			// Fetch all commands and delete those older than 'before'
			commands, err := repo.GetAllCommands()
			if err != nil {
				return errors.NewError(errors.ErrRepositoryRead, "failed to list commands", err)
			}

			var toDelete []string
			for _, c := range commands {
				if c.CreatedAt.Before(before) {
					toDelete = append(toDelete, c.ID)
				}
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Found %d commands to delete (dry-run):\n", len(toDelete))
				for _, id := range toDelete {
					fmt.Fprintln(cmd.OutOrStdout(), id)
				}
				return nil
			}

			deleted := 0
			for _, id := range toDelete {
				if err := repo.Delete(id); err != nil {
					logger.Warn("Failed to delete command", zap.String("id", id), zap.Error(err))
					continue
				}
				deleted++
			}

			logger.Info("Prune completed", zap.Int("deleted", deleted))
			fmt.Fprintf(cmd.OutOrStdout(), "Pruned %d commands\n", deleted)
			return nil
		},
	}
	pruneCmd.Flags().String("before", "", "Prune commands created before this date (YYYY-MM-DD or RFC3339)")
	pruneCmd.Flags().Bool("dry-run", false, "List commands that would be deleted without removing them")
	pruneCmd.Flags().String("tag", "", "Only prune commands that contain this tag")
	pruneCmd.Flags().String("status", "", "Only prune commands with status: success|failed")

	cmd.AddCommand(initCmd)
	cmd.AddCommand(pruneCmd)

	dc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	dc.cmd = cmd
	return dc
}

func parseDate(s string) (time.Time, error) {
	// Try YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// Try RFC3339
	return time.Parse(time.RFC3339, s)
}
