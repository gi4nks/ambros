package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/codegangsta/cli"
	"github.com/gi4nks/quant"
	"github.com/bradhe/stopwatch"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var parrot = quant.NewParrot("ambros")

var settings = Settings{}
var repository = Repository{}

func initDB() {
	repository.InitDB()
	repository.InitSchema()
}

func closeDB() {
	repository.CloseDB()
}

func readSettings() {
	settings.LoadSettings()
}

func main() {
	readSettings()
	initDB()

	// -------------------
	app := cli.NewApp()
	app.Name = "ambros"
	app.Usage = "the personal command butler!!!!"
	app.Version = "0.0.1-alpha"
	app.Copyright = "gi4nks - 2015"

	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{"ru"},
			Usage:   "run a command, remember to run the command with -- before. (./ambros r -- ls -la)",
			Action:  CmdRun,
		},
		{
			Name:    "list",
			Aliases: []string{"li"},
			Usage:   "list current configuration settings",
			Action:  CmdListSettings,
		},
		{
			Name:    "revive",
			Aliases: []string{"re"},
			Usage:   "revive ambros",
			Action:  CmdRevive,
		},
		{
			Name:    "logs",
			Aliases: []string{"lo"},
			Usage:   "show me the logs of ambros",
			Subcommands: []cli.Command{
				{
					Name:   "id",
					Usage:  "Get the log of specific id",
					Action: CmdLogsById,
				},
				{
					Name:   "all",
					Usage:  "Get all the logs",
					Action: CmdLogs,
				},
			},
		},
		/*
			{
				Name:    "history",
				Aliases: []string{"hi"},
				Usage:   "show the history of ambros. followed with a valid number shows # of history ",
				Action:  CmdHistory,
			},
		*/
		{
			Name:    "last",
			Aliases: []string{"ll"},
			Usage:   "show the last executed commands ",
			Action:  CmdLast,
		},
		{
			Name:    "output",
			Aliases: []string{"ou"},
			Usage:   "show me the output of a command executed with ambros",
			Action:  CmdOutput,
		},
		{
			Name:    "recall",
			Aliases: []string{"re"},
			Usage:   "recall a command and execute again",
			Action:  CmdRecall,
		},
		{
			Name:    "verbose",
			Aliases: []string{"ve"},
			Usage:   "set verbose level to ambros",
			Action:  CmdVerbose,
		},
	}

	app.Run(os.Args)

	defer closeDB()
}

// List of functions
func CmdRevive(ctx *cli.Context) {
	parrot.Info("==> Reviving ambros will reinitialize all the statistics.")

	repository.BackupSchema()
	repository.InitSchema()

}

func CmdLogs(ctx *cli.Context) {
	var commands = repository.GetAllCommands()

	for _, c := range commands {
		parrot.Info(c.String())
	}
}

func CmdLogsById(ctx *cli.Context) {
	id, err := stringFromArguments(ctx)
	if err != nil {
		parrot.Error("Error...", err)
		return
	}

	var command = repository.FindById(id)

	parrot.Info(command.String())
}

func CmdHistory(ctx *cli.Context) {
	var limit = settings.HistoryLimitDefault()
	limit, _ = intFromArguments(ctx)

	var commands = repository.GetHistory(limit)

	for _, c := range commands {
		parrot.Info(c.AsHistory())
	}
}

func CmdLast(ctx *cli.Context) {
	var limit = settings.LastLimitDefault()

	limit, _ = intFromArguments(ctx)
	var commands = repository.GetExecutedCommands(limit)

	for _, c := range commands {
		parrot.Info(c.AsFlatCommand())
	}
}

func CmdRecall(ctx *cli.Context) {
	id, err := stringFromArguments(ctx)
	if err != nil {
		parrot.Error("Error...", err)
		return
	}

	var stored = repository.FindById(id)
	var command = Command{}
	command.Name = stored.Name
	command.Arguments = stored.Arguments
	command.CreatedAt = time.Now()

	executeCommand(command)
}

func CmdOutput(ctx *cli.Context) {
	id, err := stringFromArguments(ctx)
	if err != nil {
		parrot.Error("Error...", err)
		return
	}

	var command = repository.FindById(id)

	parrot.Info(command.Output)
}

func CmdRun(ctx *cli.Context) {
	var command = initializeCommand(ctx)
	
	executeCommand(command)
}

func CmdListSettings(ctx *cli.Context) {
	parrot.Info(settings.String())
}

func CmdVerbose(ctx *cli.Context) {
	parrot.Info("Functionality not implemented yet!")
}

// ----------------
func stringFromArguments(ctx *cli.Context) (string, error) {
	if !ctx.Args().Present() {
		return "", errors.New("Value must be provided!")
	}

	str := ctx.Args()[0]

	return str, nil
}

func intFromArguments(ctx *cli.Context) (int, error) {
	if !ctx.Args().Present() {
		return -1, errors.New("Value must be provided!")
	}

	i, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return -1, err
	}

	return i, nil
}

func initializeCommand(ctx *cli.Context) Command {
	start := stopwatch.Start()

	var command = Command{}
	command.ID = Random()
	command.Name = ctx.Args()[0]
	command.Arguments = strings.Join(ctx.Args().Tail(), " ")
	command.CreatedAt = time.Now()

	watch := stopwatch.Stop(start)
    fmt.Printf("initializeCommand - Milliseconds elappsed: %v\n", watch.Milliseconds())
	return command
}

func finalizeCommand(command Command, output string, status bool) {
	start := stopwatch.Start()
	command.TerminatedAt = time.Now()
	command.Output = output
	command.Status = status
	repository.Put(command)
	watch := stopwatch.Stop(start)
    fmt.Printf("finalizeCommand - Milliseconds elappsed: %v\n", watch.Milliseconds())
	
}

func executeCommand(command Command) {
	start := stopwatch.Start()
	
	var buffer bytes.Buffer
	parrot.Info("sono qui - 0")

	cmd := exec.Command(command.Name, command.Arguments)
	outputReader, err := cmd.StdoutPipe()
	if err != nil {
		parrot.Error("Error creating StdoutPipe for Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	parrot.Info("sono qui - 1")

	errorReader, err := cmd.StderrPipe()
	if err != nil {
		parrot.Error("Error creating StderrPipe for Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	parrot.Info("sono qui - 2")

	scannerOutput := bufio.NewScanner(outputReader)
	go func() {
		for scannerOutput.Scan() {
			parrot.Info(scannerOutput.Text())
			buffer.WriteString(scannerOutput.Text() + "\n")
		}
	}()

	parrot.Info("sono qui - 3")

	scannerError := bufio.NewScanner(errorReader)
	go func() {
		for scannerError.Scan() {
			parrot.Info(scannerError.Text())
			buffer.WriteString(scannerError.Text() + "\n")
		}
	}()

	parrot.Info("sono qui - 4")

	err = cmd.Start()
	if err != nil {
		parrot.Error("Error starting Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	parrot.Info("sono qui - 5")

	err = cmd.Wait()
	if err != nil {
		parrot.Error("Error waiting for Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	parrot.Info("sono qui - 6")

	finalizeCommand(command, buffer.String(), true)
	
	watch := stopwatch.Stop(start)
    fmt.Printf("executeCommand - Milliseconds elappsed: %v\n", watch.Milliseconds())
	
}
