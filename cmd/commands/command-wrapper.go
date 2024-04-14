package commands

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"

	models "github.com/gi4nks/ambros/internal/models"
	"github.com/gi4nks/quant"
)

// -------------------------------
// Cli command wrapper
// -------------------------------
func CmdWrapper(args []string) {
}

func commandWrapper(args []string, cmd quant.Action0) {
	err := Repository.InitDB()

	if err != nil {
		Parrot.Println(err)
		return
	}

	err = Repository.InitSchema()

	if err != nil {
		Parrot.Println(err)
		return
	}

	CmdWrapper(args)

	cmd()

	defer Repository.CloseDB()
}

// ----------------
// command management
// ----------------
func initializeCommand(name string, arguments []string) models.Command {
	var command = models.Command{}
	command.ID = Utilities.Random()

	command.Name = name
	command.Arguments = arguments

	command.CreatedAt = time.Now()
	return command
}

func initializeCommands(cmds [][]string) []models.Command {
	var commands = []models.Command{}

	for _, cmdParts := range cmds {
		var command = models.Command{}
		command.ID = Utilities.Random()

		command.Name = cmdParts[0]
		command.Arguments = cmdParts[1:]
		command.CreatedAt = time.Now()

		// Append the command to the commands slice
		commands = append(commands, command)
	}
	return commands
}

func finalizeCommand(command *models.Command) {
	command.TerminatedAt = time.Now()
	Repository.Put(*command)

	Parrot.Println("[" + command.ID + "]")
}

func finalizeCommands(commands []*models.Command) {
	for _, command := range commands {
		command.TerminatedAt = time.Now()
		Repository.Put(*command)
		Parrot.Println("[" + command.ID + "]")
	}
}

func pushCommand(command *models.Command, showid bool) {
	command.TerminatedAt = time.Now()
	Repository.Push(*command)

	if showid {
		Parrot.Println("[" + command.ID + "]")
	}
}

func pushCommands(commands []*models.Command, showid bool) {
	for _, command := range commands {

		Parrot.Println(command.AsStoredCommand())

		command.TerminatedAt = time.Now()
		Repository.Push(*command)

		if showid {
			Parrot.Println("[" + command.ID + "]")
		}
	}
}

func executeCommand(command *models.Command) {
	var bufferOutput bytes.Buffer
	var bufferError bytes.Buffer

	cmd := exec.Command(command.Name, command.Arguments...)

	Parrot.Debug("--> CommandName " + command.Name)
	Parrot.Debug("--> Command Arguments " + Utilities.AsJson(command.Arguments))

	outputReader, err := cmd.StdoutPipe()
	if err != nil {
		Parrot.Error("Error creating StdoutPipe for Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	errorReader, err := cmd.StderrPipe()
	if err != nil {
		Parrot.Error("Error creating StderrPipe for Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	err = cmd.Start()
	if err != nil {
		Parrot.Error("Error starting Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	stopOut := make(chan bool)
	stopErr := make(chan bool)

	scannerOutput := bufio.NewScanner(outputReader)
	go func(stop chan bool) {
		for scannerOutput.Scan() {
			Parrot.Println(scannerOutput.Text())
			bufferOutput.WriteString(scannerOutput.Text() + "\n")
		}

		stop <- true
	}(stopOut)

	scannerError := bufio.NewScanner(errorReader)
	go func(stop chan bool) {
		for scannerError.Scan() {
			Parrot.Println(scannerError.Text())
			bufferError.WriteString(scannerError.Text() + "\n")
		}

		stop <- true
	}(stopErr)

	<-stopOut
	<-stopErr

	err = cmd.Wait()
	if err != nil {
		Parrot.Error("Error waiting for Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	command.Output = bufferOutput.String()
	command.Error = bufferError.String()

	command.Status = true
}

func executeCommands(commands []*models.Command) {
	var output []byte

	// Execute commands sequentially, capturing intermediate output
	for _, cmdParts := range commands {
		cmdParts.CreatedAt = time.Now()
		cmd := exec.Command(cmdParts.Name, cmdParts.Arguments...)
		var intermediate bytes.Buffer
		cmd.Stdout = &intermediate
		cmd.Stderr = &intermediate // use stderr to capture combined output

		// Write previous command output to stdin of current command if needed
		if len(output) > 0 {
			cmd.Stdin = bytes.NewReader(output)
		}

		// Executing the command and managing the error and sthe status at the end
		err := cmd.Run()
		output = intermediate.Bytes()

		Parrot.Println(string(output))
		cmdParts.Output = string(output)
		cmdParts.Error = ""

		if err != nil {
			Parrot.Error("Error running the command", err)
			cmdParts.Error = err.Error()
			cmdParts.Status = false
		} else {
			Parrot.Println(string(output))
			cmdParts.Status = true
		}

		cmdParts.TerminatedAt = time.Now()

		if err1 := Repository.Put(*cmdParts); err1 != nil {
			Parrot.Error("Error storing the command", err1)
		}

		Parrot.Println(cmdParts.AsStoredCommand() + "\n")

		if !cmdParts.Status {
			return
		}
	}
}

// ----------------
// Arguments from command string
// ----------------

func commandsFromArguments(args []string) ([][]string, error) {
	if len(args) <= 0 {
		return nil, errors.New("Value must be provided!")
	}

	var command = strings.Join(args, " ")
	// Split the command string by pipe characters
	pipeCommands := strings.Split(command, "|")

	// Split each command by spaces
	var result [][]string
	for _, cmd := range pipeCommands {
		parts := strings.Fields(strings.TrimSpace(cmd))
		result = append(result, parts)
	}
	return result, nil
}

func commandFromArguments(args []string) (string, []string, error) {
	if len(args) <= 0 {
		return "", nil, errors.New("Value must be provided!")
	}

	return args[0], Utilities.Tail(args), nil
}

func stringsFromArguments(args []string) ([]string, error) {
	if len(args) <= 0 {
		return nil, errors.New("Value must be provided!")
	}

	return args, nil
}

func stringFromArguments(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("Value must be provided!")
	}

	str := args[0]

	return str, nil
}

func intFromArguments(args []string) (int, error) {
	if len(args) != 1 {
		return -1, errors.New("Value must be provided!")
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		return -1, err
	}

	return i, nil
}
