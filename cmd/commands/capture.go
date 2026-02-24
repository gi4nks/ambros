package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins"
)

type CaptureCommand struct {
	*BaseCommand
	opts     CaptureOptions
	wrapper  *CommandWrapper
	executor *Executor
	exitFunc func(int)
}

type CaptureOptions struct {
	store    bool
	tag      []string
	category string
	redact   bool
	maxBytes int64
}

func NewCaptureCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *CaptureCommand {
	cc := &CaptureCommand{
		wrapper:  NewCommandWrapper(logger, repo),
		executor: NewExecutor(logger),
		exitFunc: os.Exit,
	}

	cmd := &cobra.Command{
		Use:   "capture [flags] [--] <command> [args...]",
		Short: "Capture command output using pipes (ported from Ecto)",
		Long: `Execute a command and capture its output using pipes instead of PTY.
This is particularly useful for capturing clean output without escape sequences.
It also supports automatic redaction of secrets.`,
		RunE: cc.runE,
		Args: cobra.MinimumNArgs(1),
	}

	cc.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	cc.cmd = cmd
	cc.setupFlags(cmd)
	return cc
}

func (cc *CaptureCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&cc.opts.store, "store", true,
		"Store the command execution details")
	cmd.Flags().StringSliceVarP(&cc.opts.tag, "tag", "t", nil,
		"Add tags to the command")
	cmd.Flags().StringVarP(&cc.opts.category, "category", "c", "",
		"Assign a category to the command")
	cmd.Flags().BoolVar(&cc.opts.redact, "redact", true,
		"Redact secrets from captured output")
	cmd.Flags().Int64Var(&cc.opts.maxBytes, "max-bytes", 10*1024*1024,
		"Maximum bytes to capture (default 10MB)")
}

func (cc *CaptureCommand) runE(cmd *cobra.Command, args []string) error {
	cc.logger.Debug("Capture invoked",
		zap.Strings("rawArgs", args),
		zap.Bool("redact", cc.opts.redact),
		zap.Bool("store", cc.opts.store))

	if len(args) == 0 {
		return errors.NewError(errors.ErrInvalidCommand, "No command specified", nil)
	}

	commandName := args[0]
	commandArgs := args[1:]

	// Create command record
	command := &models.Command{
		Entity: models.Entity{
			ID:        cc.generateCommandID(),
			CreatedAt: time.Now(),
		},
		Name:      commandName,
		Arguments: commandArgs,
		Tags:      cc.opts.tag,
		Category:  cc.opts.category,
	}

	res, err := cc.executor.ExecutePipe(commandName, commandArgs, PipeOptions{
		MaxBytes:    cc.opts.maxBytes,
		Redact:      cc.opts.redact,
		PrintStderr: true,
	})

	command.TerminatedAt = time.Now()
	command.Status = err == nil && res.ExitCode == 0
	command.Output = res.Stdout
	command.Error = res.Stderr

	// Store if requested
	if cc.opts.store {
		if err := cc.wrapper.FinalizeCommand(command); err != nil {
			cc.logger.Warn("Failed to store captured command execution",
				zap.Error(err),
				zap.String("commandId", command.ID))
		} else {
			duration := command.TerminatedAt.Sub(command.CreatedAt)
			if command.Status {
				color.Green("[%s] ✅ Success (%v)", command.ID, duration.Round(time.Millisecond))
			} else {
				color.Red("[%s] ❌ Failed (%v)", command.ID, duration.Round(time.Millisecond))
			}
		}
	}

	if res.ExitCode != 0 {
		cc.exitFunc(res.ExitCode)
	}

	return nil
}

func (cc *CaptureCommand) generateCommandID() string {
	return fmt.Sprintf("CAP-%d", time.Now().UnixNano())
}

func (cc *CaptureCommand) Command() *cobra.Command {
	return cc.cmd
}
