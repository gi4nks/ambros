package main

import (
	"bufio"
	"bytes"
	"github.com/codegangsta/cli"
	"github.com/gi4nks/quant"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var parrot = quant.NewParrot("ambros")

var settings = Settings{}
var repository = Repository{}

func configureDB() {
	repository.InitDB()
}

func readSettings() {
	settings.LoadSettings()
}

func main() {
	readSettings()
	configureDB()

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
			Aliases: []string{"e"},
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
	if !ctx.Args().Present() {

		parrot.Info("Id must be provided!")
		return
	}

	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		// handle error
		parrot.Info("Id provided is not valid!\n")
		parrot.Error("Error...", err)
		return
	}

	var command = repository.FindById(id)

	parrot.Info(command.String())
}

func CmdHistory(ctx *cli.Context) {
	var limit = settings.HistoryLimitDefault()
	var err error
	if ctx.Args().Present() {

		limit, err = strconv.Atoi(ctx.Args()[0])
		if err != nil {
			// handle error
			parrot.Info(err.Error())
			limit = settings.HistoryLimitDefault()
		}
	}

	var commands = repository.GetHistory(limit)

	for _, c := range commands {
		parrot.Info(c.AsHistory())
	}
}

func CmdLast(ctx *cli.Context) {
	var limit = settings.LastLimitDefault()
	var err error
	if ctx.Args().Present() {

		limit, err = strconv.Atoi(ctx.Args()[0])
		if err != nil {
			// handle error
			parrot.Info(err.Error())
			limit = settings.LastLimitDefault()
		}
	}

	var commands = repository.GetExecutedCommands(limit)

	for _, c := range commands {
		parrot.Info(c.AsFlatCommand())
	}
}

func CmdRecall(ctx *cli.Context) {
	if !ctx.Args().Present() {

		parrot.Info("Id must be provided!")
		return
	}

	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		// handle error
		parrot.Info("Id provided is not valid!\n")
		parrot.Error("Error...", err)
		return
	}

	var stored = repository.FindById(id)
	var command = Command{}
	command.Name = stored.Name
	command.Arguments = stored.Arguments
	command.CreatedAt = time.Now()

	execute(command)
}

func CmdOutput(ctx *cli.Context) {
	if !ctx.Args().Present() {

		parrot.Info("Id must be provided!")
		return
	}

	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		// handle error
		parrot.Info("Id provided is not valid!\n")
		parrot.Error("Error...", err)
		return
	}

	var command = repository.FindById(id)

	parrot.Info(command.Output)
}

func CmdRun(ctx *cli.Context) {

	var command = Command{}
	command.Name = ctx.Args()[0]
	command.Arguments = strings.Join(ctx.Args().Tail(), " ")
	command.CreatedAt = time.Now()

	execute(command)
}

func CmdListSettings(ctx *cli.Context) {
	parrot.Info(settings.String())
}

func CmdVerbose(ctx *cli.Context) {
	parrot.Info("Functionality not implemented yet!")
}

// ----------------

func finalizeCommand(command Command, output string, status bool) {
	command.TerminatedAt = time.Now()
	command.Output = output
	command.Status = status
	repository.Put(command)
}

func execute(command Command) {
	var buffer bytes.Buffer

	cmd := exec.Command(command.Name, command.Arguments)
	outputReader, err := cmd.StdoutPipe()
	if err != nil {
		parrot.Error("Error creating StdoutPipe for Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	errorReader, err := cmd.StderrPipe()
	if err != nil {
		parrot.Error("Error creating StderrPipe for Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	scannerOutput := bufio.NewScanner(outputReader)
	go func() {
		for scannerOutput.Scan() {
			parrot.Info(scannerOutput.Text())
			buffer.WriteString(scannerOutput.Text() + "\n")
		}
	}()

	scannerError := bufio.NewScanner(errorReader)
	go func() {
		for scannerError.Scan() {
			parrot.Info(scannerError.Text())
			buffer.WriteString(scannerError.Text() + "\n")
		}
	}()

	err = cmd.Start()
	if err != nil {
		parrot.Error("Error starting Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	err = cmd.Wait()
	if err != nil {
		parrot.Error("Error waiting for Cmd", err)
		finalizeCommand(command, err.Error(), false)
		return
	}

	finalizeCommand(command, buffer.String(), true)
}
