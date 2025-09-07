package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/gi4nks/ambros/internal/errors"
	"github.com/gi4nks/ambros/internal/models"
)

type ImportCommand struct {
	*BaseCommand
	inputFile    string
	format       string
	merge        bool
	dryRun       bool
	skipExisting bool
}

func NewImportCommand(logger *zap.Logger, repo RepositoryInterface) *ImportCommand {
	ic := &ImportCommand{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import commands from file",
		Long: `Import commands from a file in various formats.
Supports JSON and YAML formats with options for merging and validation.

Examples:
  ambros import -i commands.json                  # Import from JSON file
  ambros import -i backup.yaml -f yaml           # Import from YAML file
  ambros import -i data.json --merge              # Merge with existing commands
  ambros import -i data.json --dry-run            # Preview import without changes
  ambros import -i data.json --skip-existing      # Skip commands that already exist`,
		RunE: ic.runE,
	}

	ic.BaseCommand = NewBaseCommand(cmd, logger, repo)
	ic.cmd = cmd
	ic.setupFlags(cmd)
	return ic
}

func (ic *ImportCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&ic.inputFile, "input", "i", "", "Input file path")
	cmd.Flags().StringVarP(&ic.format, "format", "f", "json", "Input format (json or yaml)")
	cmd.Flags().BoolVar(&ic.merge, "merge", false, "Merge with existing commands")
	cmd.Flags().BoolVar(&ic.dryRun, "dry-run", false, "Show what would be imported without making changes")
	cmd.Flags().BoolVar(&ic.skipExisting, "skip-existing", false, "Skip commands that already exist")

	cmd.MarkFlagRequired("input")
}

func (ic *ImportCommand) runE(cmd *cobra.Command, args []string) error {
	ic.logger.Debug("Import command invoked",
		zap.String("inputFile", ic.inputFile),
		zap.String("format", ic.format),
		zap.Bool("merge", ic.merge),
		zap.Bool("dryRun", ic.dryRun),
		zap.Bool("skipExisting", ic.skipExisting))

	// Validate flags
	if err := ic.validateFlags(); err != nil {
		return err
	}

	// Read and parse input file
	commands, err := ic.readImportFile()
	if err != nil {
		return err
	}

	ic.logger.Info("Parsed import file",
		zap.String("file", ic.inputFile),
		zap.Int("commandCount", len(commands)))

	// Process commands
	return ic.processCommands(commands)
}

func (ic *ImportCommand) validateFlags() error {
	if ic.format != "json" && ic.format != "yaml" {
		return errors.NewError(errors.ErrInvalidCommand, "unsupported format", nil)
	}

	if _, err := os.Stat(ic.inputFile); os.IsNotExist(err) {
		return errors.NewError(errors.ErrInvalidCommand, "input file does not exist", err)
	}

	return nil
}

func (ic *ImportCommand) readImportFile() ([]models.Command, error) {
	data, err := os.ReadFile(ic.inputFile)
	if err != nil {
		ic.logger.Error("Failed to read input file",
			zap.String("file", ic.inputFile),
			zap.Error(err))
		return nil, errors.NewError(errors.ErrInternalServer, "failed to read input file", err)
	}

	var importData struct {
		Commands []models.Command `json:"commands" yaml:"commands"`
	}

	switch ic.format {
	case "json":
		err = json.Unmarshal(data, &importData)
	case "yaml":
		err = yaml.Unmarshal(data, &importData)
	default:
		return nil, errors.NewError(errors.ErrInvalidCommand, "unsupported format", nil)
	}

	if err != nil {
		ic.logger.Error("Failed to parse input file",
			zap.String("file", ic.inputFile),
			zap.String("format", ic.format),
			zap.Error(err))
		return nil, errors.NewError(errors.ErrInternalServer, "failed to parse input file", err)
	}

	return importData.Commands, nil
}

func (ic *ImportCommand) processCommands(commands []models.Command) error {
	if ic.dryRun {
		return ic.previewImport(commands)
	}

	imported := 0
	skipped := 0
	errors := 0

	for _, command := range commands {
		// Check if command already exists if skipExisting is enabled
		if ic.skipExisting && ic.commandExists(command.ID) {
			skipped++
			ic.logger.Debug("Skipping existing command",
				zap.String("commandId", command.ID))
			continue
		}

		// Update timestamps for import
		command.CreatedAt = time.Now()
		if command.TerminatedAt.IsZero() {
			command.TerminatedAt = command.CreatedAt
		}

		// Import the command
		if err := ic.repository.Put(context.Background(), command); err != nil {
			errors++
			ic.logger.Error("Failed to import command",
				zap.String("commandId", command.ID),
				zap.Error(err))
			continue
		}

		imported++
		ic.logger.Debug("Imported command",
			zap.String("commandId", command.ID),
			zap.String("name", command.Name))
	}

	// Display summary
	fmt.Printf("Import completed:\n")
	fmt.Printf("Total commands: %d\n", len(commands))
	fmt.Printf("Imported: %d\n", imported)
	if skipped > 0 {
		fmt.Printf("Skipped: %d\n", skipped)
	}
	if errors > 0 {
		fmt.Printf("Errors: %d\n", errors)
	}

	ic.logger.Info("Import process completed",
		zap.Int("total", len(commands)),
		zap.Int("imported", imported),
		zap.Int("skipped", skipped),
		zap.Int("errors", errors))

	return nil
}

func (ic *ImportCommand) previewImport(commands []models.Command) error {
	fmt.Printf("Import preview for file: %s\n", ic.inputFile)
	fmt.Printf("Format: %s\n", ic.format)
	fmt.Printf("Total commands to import: %d\n\n", len(commands))

	existing := 0
	for i, command := range commands {
		if ic.commandExists(command.ID) {
			existing++
		}

		fmt.Printf("%d. %s %s\n", i+1, command.Name, fmt.Sprintf("(ID: %s)", command.ID))
		fmt.Printf("   Created: %s\n", command.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Status: %s\n", ic.formatStatus(command.Status))

		if len(command.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", fmt.Sprintf("[%s]", fmt.Sprintf("%v", command.Tags)))
		}

		if ic.commandExists(command.ID) {
			fmt.Printf("   Note: Command already exists\n")
		}

		if i < len(commands)-1 {
			fmt.Println()
		}
	}

	if existing > 0 {
		fmt.Printf("\nNote: %d command(s) already exist\n", existing)
		if ic.skipExisting {
			fmt.Printf("These will be skipped due to --skip-existing flag\n")
		} else {
			fmt.Printf("These will be overwritten\n")
		}
	}

	ic.logger.Info("Import preview completed",
		zap.Int("totalCommands", len(commands)),
		zap.Int("existingCommands", existing))

	return nil
}

func (ic *ImportCommand) commandExists(id string) bool {
	_, err := ic.repository.Get(id)
	return err == nil
}

func (ic *ImportCommand) formatStatus(status bool) string {
	if status {
		return "Success"
	}
	return "Failed"
}

func (ic *ImportCommand) Command() *cobra.Command {
	return ic.cmd
}
