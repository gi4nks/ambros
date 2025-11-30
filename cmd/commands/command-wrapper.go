package commands

import (
	"bufio"
	"bytes"
	"context"
	"errors" // Standard errors package
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	ambrosErrors "github.com/gi4nks/ambros/v3/internal/errors" // Alias for ambros's custom errors
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/utils"
)

// CommandWrapper provides utilities for command execution and management
type CommandWrapper struct {
	logger     *zap.Logger
	repository RepositoryInterface
	utilities  *utils.Utilities
	executor   *Executor // Add this field
}

// NewCommandWrapper creates a new command wrapper
func NewCommandWrapper(logger *zap.Logger, repo RepositoryInterface) *CommandWrapper {
	return &CommandWrapper{
		logger:     logger,
		repository: repo,
		utilities:  utils.NewUtilities(logger),
		executor:   NewExecutor(logger), // Initialize the executor
	}
}

// InitializeCommand creates a new command with basic metadata
func (cw *CommandWrapper) InitializeCommand(name string, arguments []string) models.Command {
	return models.Command{
		Entity: models.Entity{
			ID:        cw.utilities.Random(),
			CreatedAt: time.Now(),
		},
		Name:      name,
		Arguments: arguments,
	}
}

// InitializeCommands creates multiple commands from command parts
func (cw *CommandWrapper) InitializeCommands(cmds [][]string) []models.Command {
	commands := make([]models.Command, 0, len(cmds))

	for _, cmdParts := range cmds {
		if len(cmdParts) == 0 {
			continue
		}

		command := models.Command{
			Entity: models.Entity{
				ID:        cw.utilities.Random(),
				CreatedAt: time.Now(),
			},
			Name:      cmdParts[0],
			Arguments: cmdParts[1:],
		}

		commands = append(commands, command)
	}
	return commands
}

// FinalizeCommand completes and stores a command
func (cw *CommandWrapper) FinalizeCommand(command *models.Command) error {
	command.TerminatedAt = time.Now()

	if err := cw.repository.Put(context.Background(), *command); err != nil {
		cw.logger.Error("Failed to store command",
			zap.String("id", command.ID),
			zap.Error(err))
		return err
	}

	cw.logger.Info("Command finalized", zap.String("id", command.ID))
	return nil
}

// FinalizeCommands completes and stores multiple commands
func (cw *CommandWrapper) FinalizeCommands(commands []*models.Command) error {
	var lastErr error

	for _, command := range commands {
		command.TerminatedAt = time.Now()

		if err := cw.repository.Put(context.Background(), *command); err != nil {
			cw.logger.Error("Failed to store command",
				zap.String("id", command.ID),
				zap.Error(err))
			lastErr = err
			continue
		}

		cw.logger.Info("Command finalized", zap.String("id", command.ID))
	}

	return lastErr
}

// PushCommand stores a command using the Push method
func (cw *CommandWrapper) PushCommand(command *models.Command, showid bool) error {
	command.TerminatedAt = time.Now()

	if err := cw.repository.Push(*command); err != nil {
		cw.logger.Error("Failed to push command",
			zap.String("id", command.ID),
			zap.Error(err))
		return err
	}

	if showid {
		cw.logger.Info("Command pushed", zap.String("id", command.ID))
	}

	return nil
}

// PushCommands stores multiple commands using the Push method
func (cw *CommandWrapper) PushCommands(commands []*models.Command, showid bool) error {
	var lastErr error

	for _, command := range commands {
		cw.logger.Info("Storing command",
			zap.String("command", command.AsStoredCommand()))

		command.TerminatedAt = time.Now()

		if err := cw.repository.Push(*command); err != nil {
			cw.logger.Error("Failed to push command",
				zap.String("id", command.ID),
				zap.Error(err))
			lastErr = err
			continue
		}

		if showid {
			cw.logger.Info("Command pushed", zap.String("id", command.ID))
		}
	}

	return lastErr
}

// ExecuteCommand executes a single command and captures output
func (cw *CommandWrapper) ExecuteCommand(command *models.Command) {
	var bufferOutput bytes.Buffer
	var bufferError bytes.Buffer

	// Resolve and validate executable path before executing
	resolved, err := cw.executor.ResolveCommandPath(command.Name)
	if err != nil {
		cw.logger.Error("Invalid command name", zap.String("name", command.Name), zap.Error(err))
		command.Error = err.Error()
		command.Status = false
		return
	}
	cmd := exec.Command(resolved, command.Arguments...)

	cw.logger.Debug("Executing command",
		zap.String("name", command.Name),
		zap.Strings("arguments", command.Arguments))

	// Emit a warning if the command looks interactive; the CommandWrapper is
	// frequently used by scripts or server-side code so we avoid printing to
	// stdout/stderr and rely on structured logging.
	if cw.executor.isLikelyInteractive(command.Name, command.Arguments) {
		cw.logger.Warn("Likely interactive command executed via CommandWrapper",
			zap.String("command", command.Name), zap.Strings("args", command.Arguments))
	}

	outputReader, err := cmd.StdoutPipe()
	if err != nil {
		cw.logger.Error("Error creating StdoutPipe for Cmd", zap.Error(err))
		command.Error = err.Error()
		command.Status = false
		return
	}

	errorReader, err := cmd.StderrPipe()
	if err != nil {
		cw.logger.Error("Error creating StderrPipe for Cmd", zap.Error(err))
		command.Error = err.Error()
		command.Status = false
		return
	}

	err = cmd.Start()
	if err != nil {
		cw.logger.Error("Error starting Cmd", zap.Error(err))
		command.Error = err.Error()
		command.Status = false
		return
	}

	stopOut := make(chan bool)
	stopErr := make(chan bool)

	scannerOutput := bufio.NewScanner(outputReader)
	go func(stop chan bool) {
		for scannerOutput.Scan() {
			text := scannerOutput.Text()
			cw.logger.Debug("Command output", zap.String("output", text))
			bufferOutput.WriteString(text + "\n")
		}
		stop <- true
	}(stopOut)

	scannerError := bufio.NewScanner(errorReader)
	go func(stop chan bool) {
		for scannerError.Scan() {
			text := scannerError.Text()
			cw.logger.Debug("Command error output", zap.String("error", text))
			bufferError.WriteString(text + "\n")
		}
		stop <- true
	}(stopErr)

	<-stopOut
	<-stopErr

	err = cmd.Wait()
	if err != nil {
		cw.logger.Error("Error waiting for Cmd", zap.Error(err))
		command.Error = err.Error()
		command.Status = false
		return
	}

	command.Output = bufferOutput.String()
	if bufferError.Len() > 0 {
		command.Error = bufferError.String()
		command.Status = false
	} else {
		command.Status = true
	}
}

// ExecuteCommands executes multiple commands in sequence (pipeline)
func (cw *CommandWrapper) ExecuteCommands(commands []*models.Command) error {
	var output []byte

	for _, command := range commands {
		command.CreatedAt = time.Now()
		// Validate command name and resolve path; continue to attempt storing even on error
		resolved, err := cw.executor.ResolveCommandPath(command.Name)
		if err != nil {
			cw.logger.Error("Invalid command name", zap.String("name", command.Name), zap.Error(err))
			command.Error = err.Error()
			command.Status = false
		}
		execName := command.Name
		if resolved != "" {
			execName = resolved
		}
		cmd := exec.Command(execName, command.Arguments...)
		var intermediate bytes.Buffer
		cmd.Stdout = &intermediate
		cmd.Stderr = &intermediate

		// Pipe previous command output to stdin of current command
		if len(output) > 0 {
			cmd.Stdin = bytes.NewReader(output)
		}

		err = cmd.Run()
		output = intermediate.Bytes()

		cw.logger.Debug("Command output", zap.String("output", string(output)))
		command.Output = string(output)
		command.Error = ""

		if err != nil {
			cw.logger.Error("Error running the command", zap.Error(err))
			command.Error = err.Error()
			command.Status = false
		} else {
			cw.logger.Info("Command completed successfully")
			command.Status = true
		}

		command.TerminatedAt = time.Now()

		if err := cw.repository.Put(context.Background(), *command); err != nil {
			cw.logger.Error("Error storing the command", zap.Error(err))
			return err
		}

		cw.logger.Info("Command stored",
			zap.String("command", command.AsStoredCommand()))

		// Stop pipeline if command failed
		if !command.Status {
			return ambrosErrors.NewError(ambrosErrors.ErrExecutionFailed, "command pipeline failed", nil)
		}
	}

	return nil
}

// Utility functions for parsing arguments

// CommandsFromArguments parses a command string into multiple commands split by pipes
func (cw *CommandWrapper) CommandsFromArguments(args []string) ([][]string, error) {
	if len(args) <= 0 {
		return nil, ambrosErrors.NewError(ambrosErrors.ErrInvalidCommand, "arguments must be provided", nil)
	}
	command := strings.Join(args, " ")
	pipeCommands, err := splitByUnescapedPipe(command)
	if err != nil {
		return nil, err
	}

	result := make([][]string, 0, len(pipeCommands))
	for _, cmd := range pipeCommands {
		parts, err := shellFields(strings.TrimSpace(cmd))
		if err != nil {
			return nil, err
		}
		if len(parts) > 0 {
			result = append(result, parts)
		}
	}

	return result, nil
}

// CommandFromArguments extracts command name and arguments
func (cw *CommandWrapper) CommandFromArguments(args []string) (string, []string, error) {
	if len(args) <= 0 {
		return "", nil, errors.New("arguments must be provided")
	}

	return args[0], cw.utilities.Tail(args), nil
}

// StringsFromArguments validates and returns string arguments
func (cw *CommandWrapper) StringsFromArguments(args []string) ([]string, error) {
	if len(args) <= 0 {
		return nil, ambrosErrors.NewError(ambrosErrors.ErrInvalidCommand, "arguments must be provided", nil)
	}

	return args, nil
}

// StringFromArguments extracts a single string argument
func (cw *CommandWrapper) StringFromArguments(args []string) (string, error) {
	if len(args) < 1 {
		return "", ambrosErrors.NewError(ambrosErrors.ErrInvalidCommand, "argument must be provided", nil)
	}

	return args[0], nil
}

// IntFromArguments extracts a single integer argument
func (cw *CommandWrapper) IntFromArguments(args []string) (int, error) {
	if len(args) != 1 {
		return -1, ambrosErrors.NewError(ambrosErrors.ErrInvalidCommand, "exactly one argument must be provided", nil)
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		return -1, ambrosErrors.NewError(ambrosErrors.ErrInvalidCommand, "argument must be a valid integer", err)
	}

	return i, nil
}
