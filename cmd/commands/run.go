package commands

import (
	"fmt"
	"os"
	"strings" // New import
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
)

type RunCommand struct {
	*BaseCommand
	opts     RunOptions
	wrapper  *CommandWrapper
	executor *Executor // New field for shared execution logic
	exitFunc func(int) // For testable exit behavior
}

type RunOptions struct {
	store    bool
	tag      []string
	category string
	template string
	dryRun   bool
	auto     bool
}

func NewRunCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *RunCommand {
	rc := &RunCommand{
		wrapper:  NewCommandWrapper(logger, repo),
		executor: NewExecutor(logger), // Initialize the executor
		exitFunc: os.Exit,             // Default to os.Exit
	}

	cmd := &cobra.Command{
		Use:   "run [flags] [--] <command> [args...]",
		Short: "Run a command and optionally store its execution details",
		Long: `Execute a command and capture its output, error, and execution details.
Optionally store the execution for future reference.

The -- separator is optional for simple commands without flags, but required when the 
command you're running has flags that might conflict with ambros flags.

Examples:
  ambros run echo "hello"                  # Simple command, no -- needed
  ambros run -- ls -la                    # Required: ls flags would conflict
  ambros run --store -- echo "hello"      # Store the command execution
  ambros run -t dev,test -- make build    # Run with tags
  ambros run --dry-run -- rm -rf /        # Show what would be executed

Note: If you get "unknown flag" errors, add -- before your command.`,
		RunE: rc.runE,
		Args: cobra.MinimumNArgs(1),
	}

	rc.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	rc.cmd = cmd
	rc.setupFlags(cmd)
	return rc
}

func (rc *RunCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&rc.opts.store, "store", true,
		"Store the command execution details")
	cmd.Flags().StringSliceVarP(&rc.opts.tag, "tag", "t", nil,
		"Add tags to the command")
	cmd.Flags().StringVarP(&rc.opts.category, "category", "c", "",
		"Assign a category to the command")
	cmd.Flags().StringVarP(&rc.opts.template, "template", "p", "",
		"Use a command template")
	cmd.Flags().BoolVar(&rc.opts.dryRun, "dry-run", false,
		"Show what would be executed without running")
	cmd.Flags().BoolVar(&rc.opts.auto, "auto", false,
		"Transparent execution: stream output and preserve exit code")
}

func (rc *RunCommand) runE(cmd *cobra.Command, args []string) error {
	// Delegate to the new Execute API which returns an exit code and error
	exitCode, err := rc.Execute(args)
	if err != nil {
		// If Execute returned an error (parsing or execution), log and wrap
		rc.logger.Error("Run execution error", zap.Error(err))
		return err
	}

	// Preserve previous behavior: for --auto, exit with child's code; for non-auto, exit 1 on failure
	if rc.opts.auto {
		if exitCode != 0 {
			rc.logger.Debug("Command failed in auto mode, exiting with child exit code",
				zap.Int("exitCode", exitCode))
			rc.exitFunc(exitCode)
		}
		return nil
	}

	if exitCode != 0 {
		rc.logger.Debug("Command failed, exiting with code 1")
		rc.exitFunc(1)
	}

	return nil
}

// Execute parses flags from args, executes the command according to options and returns the
// child's exit code and any error encountered during setup/execution. It does not call rc.exitFunc.
func (rc *RunCommand) Execute(args []string) (int, error) {
	rc.logger.Debug("Execute invoked",
		zap.Strings("rawArgs", args),
		zap.String("template", rc.opts.template),
		zap.Bool("dryRun", rc.opts.dryRun),
		zap.Bool("store", rc.opts.store))

	// The args parameter is expected to be the command and its args (this matches the
	// previous behavior of runE when invoked directly in tests). We do NOT parse
	// cobra flags here because when Cobra calls runE it already parsed flags and
	// supplied only the non-flag args. Tests call runE directly and set rc.opts
	// manually.
	rem := args
	if len(rem) == 0 {
		return 1, errors.NewError(errors.ErrInvalidCommand,
			"No command specified. Use: ambros run [flags] -- <command> [args...]", nil)
	}

	// Handle template if specified
	if rc.opts.template != "" {
		template, err := rc.repository.GetTemplate(rc.opts.template)
		if err != nil {
			rc.logger.Error("Template not found",
				zap.String("template", rc.opts.template),
				zap.Error(err))
			return 1, errors.NewError(errors.ErrCommandNotFound,
				fmt.Sprintf("template not found: %s", rc.opts.template), err)
		}
		rem = template.BuildCommand(rem)
		rc.logger.Debug("Template applied",
			zap.String("template", rc.opts.template),
			zap.Strings("resultArgs", rem))
	}

	commandName := rem[0]
	commandArgs := rem[1:]

	// Create command record
	command := &models.Command{
		Entity: models.Entity{
			ID:        rc.generateCommandID(),
			CreatedAt: time.Now(),
		},
		Name:      commandName,
		Arguments: commandArgs,
		Tags:      rc.opts.tag,
		Category:  rc.opts.category,
	}

	if rc.opts.dryRun {
		// User output (keep fmt)
		fmt.Printf("Would execute: %s %s\n", commandName, strings.Join(commandArgs, " "))
		if len(rc.opts.tag) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(rc.opts.tag, ", "))
		}
		if rc.opts.category != "" {
			fmt.Printf("Category: %s\n", rc.opts.category)
		}

		rc.logger.Info("Dry run completed",
			zap.String("command", commandName),
			zap.Strings("args", commandArgs))
		return 0, nil
	}

	// ExecuteCapture runs the command similarly to Execute but always captures combined
	// stdout/stderr into a byte buffer and returns (exitCode, output, error).

	// Execute the command (support transparent --auto mode)
	var output, errorMsg string
	var success bool
	var execErr error
	var exitCode int

	// If the command looks interactive and user didn't request --auto, warn them
	if rc.executor.isLikelyInteractive(commandName, commandArgs) && !rc.opts.auto {
		// Hint the user about --auto which attaches a ttys/streams
		fmt.Fprintln(os.Stderr, color.YellowString("Warning: command looks interactive. Use --auto to attach a TTY and stream input/output."))
		rc.logger.Warn("Likely interactive command executed without --auto", zap.String("command", commandName))
	}

	if rc.opts.auto {
		// Default auto behavior streams output to the terminal. For callers who want
		// to capture the output (like the interactive UI), call ExecuteCapture
		// instead. Here we keep the default behavior for Execute.
		exitCode, execErr = rc.executor.ExecuteCommandAuto(commandName, commandArgs)
		success = execErr == nil && exitCode == 0
		// executeCommandAuto streams output directly; we don't capture combined output here
		output = ""
		if execErr != nil && exitCode == 0 {
			// start failed
			errorMsg = execErr.Error()
		}
	} else {
		// Execute the command normally and capture output
		output, errorMsg, success, execErr = rc.executor.ExecuteCommand(commandName, commandArgs)
	}

	if execErr != nil && !rc.opts.auto {
		rc.logger.Error("Command execution failed",
			zap.String("command", commandName),
			zap.Strings("args", commandArgs),
			zap.Error(execErr))
		return 1, errors.NewError(errors.ErrExecutionFailed, "failed to execute command", execErr)
	}

	// Update command with results
	command.TerminatedAt = time.Now()
	command.Status = success
	command.Output = output
	command.Error = errorMsg

	// Display output to user (keep fmt for user output)
	if output != "" {
		fmt.Print(output)
	}
	if errorMsg != "" {
		fmt.Fprint(os.Stderr, errorMsg)
	}

	// Store if requested
	if rc.opts.store {
		if err := rc.wrapper.FinalizeCommand(command); err != nil {
			rc.logger.Warn("Failed to store command execution",
				zap.Error(err),
				zap.String("commandId", command.ID))
		} else {
			// Show user-friendly confirmation message with color and timing
			duration := command.TerminatedAt.Sub(command.CreatedAt)
			if success {
				color.Green("[%s] ✅ Success (%v)", command.ID, duration.Round(time.Millisecond))
			} else {
				color.Red("[%s] ❌ Failed (%v)", command.ID, duration.Round(time.Millisecond))
			}
		}
	}

	// Log execution summary (only in debug mode)
	rc.logger.Debug("Command execution completed",
		zap.String("commandId", command.ID),
		zap.String("command", commandName),
		zap.Strings("args", commandArgs),
		zap.Bool("success", success),
		zap.Duration("duration", command.TerminatedAt.Sub(command.CreatedAt)),
		zap.Bool("stored", rc.opts.store))

	// Return child's exit code (0 for success, non-zero for failure)
	if rc.opts.auto {
		if execErr != nil && exitCode == 0 {
			// child failed to start
			return 1, execErr
		}
		return exitCode, nil
	}

	if !success {
		return 1, nil
	}

	return 0, nil
}

func (rc *RunCommand) generateCommandID() string {
	// Simple ID generation - could be enhanced
	return fmt.Sprintf("CMD-%d", time.Now().UnixNano())
}

func (rc *RunCommand) Command() *cobra.Command {
	return rc.cmd
}
