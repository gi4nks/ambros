package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"time"

	models "github.com/gi4nks/ambros/models"
	"github.com/gi4nks/quant"
)

// -------------------------------
// Cli command wrapper
// -------------------------------
func CmdWrapper(args []string) {
}

func commandWrapper(args []string, cmd quant.Action0) {
	Repository.InitDB()
	Repository.InitSchema()

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

func finalizeCommand(command *models.Command) {
	command.TerminatedAt = time.Now()
	Repository.Put(*command)

	Parrot.Println("[" + command.ID + "]")
}

func pushCommand(command *models.Command, showid bool) {
	command.TerminatedAt = time.Now()
	Repository.Push(*command)

	if showid {
		Parrot.Println("[" + command.ID + "]")
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

// ----------------
// Arguments from command string
// ----------------
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
