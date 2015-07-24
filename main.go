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
			Action:  runCommand,
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list current configuration settings",
			Action:  listSettings,
		},
		{
			Name:    "revive",
			Aliases: []string{"e"},
			Usage:   "revive ambros",
			Action:  revive,
		},
		{
			Name:    "logs",
			Aliases: []string{"lo"},
			Usage:   "show me the logs of ambros",
			Action:  logs,
		},
		{
			Name:    "history",
			Aliases: []string{"y"},
			Usage:   "show the history of ambros. followed with a valid number shows nummber of history ",
			Action:  history,
		},
		{
			Name:    "last",
			Aliases: []string{"ll"},
			Usage:   "show the last executed commands ",
			Action:  last,
		},
		{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "show me the output of a command executed with ambros",
			Action:  output,
		},
		{
			Name:    "recall",
			Aliases: []string{"re"},
			Usage:   "recall a command and execute again",
			Action:  recall,
		},
		{
			Name:    "verbose",
			Aliases: []string{"ve"},
			Usage:   "set verbose level to ambros",
			Action:  foo,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "debug", Usage: "set verbosity to debug"},
			},
			/*
				Subcommands: []cli.Command{
					{
						Name:  "add",
						Usage: "add a new template",
						Action: func(c *cli.Context) {
							println("new task template: ", c.Args().First())
						},
					},
					{
						Name:  "remove",
						Usage: "remove an existing template",
						Action: func(c *cli.Context) {
							println("removed task template: ", c.Args().First())
						},
					},
				},
			*/
		},
	}

	app.Run(os.Args)
}

// List of functions
func revive(ctx *cli.Context) {
	parrot.Info("==> Reviving ambros will reinitialize all the statistics.")

	repository.BackupSchema()

	repository.InitSchema()
}

func logs(ctx *cli.Context) {
	var commands = repository.GetAllCommands()

	for _, c := range commands {
		//TODO structure the output n a readable way
		parrot.Info(c.String())
	}
}

func history(ctx *cli.Context) {
	var count = settings.HistoryCountDefault()
	var err error
	if ctx.Args().Present() {

		count, err = strconv.Atoi(ctx.Args()[0])
		if err != nil {
			// handle error
			parrot.Info(err.Error())
			count = settings.HistoryCountDefault()
		}
	}

	var commands = repository.GetHistory(count)

	for _, c := range commands {
		//TODO structure the output n a readable way
		parrot.Info(c.AsHistory())
	}
}

func last(ctx *cli.Context) {
	var count = settings.LastCountDefault()
	var err error
	if ctx.Args().Present() {

		count, err = strconv.Atoi(ctx.Args()[0])
		if err != nil {
			// handle error
			parrot.Info(err.Error())
			count = settings.LastCountDefault()
		}
	}

	var commands = repository.GetExecutedCommands(count)

	for _, c := range commands {
		parrot.Info(c.AsFlatCommand())
	}
}

func prepareEnvironment(ctx *cli.Context) {
	parrot.Info("Prepare the environment!!!!")
}

func recall(ctx *cli.Context) {
	if !ctx.Args().Present() {

		parrot.Info("Id must be provided!")
		return
	}

	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		// handle error
		parrot.Info("Id provided is not valid!")
		parrot.Info(err.Error())
		return
	}

	var stored = repository.FindById(id)
	var command = Command{}
	command.Name = stored.Name
	command.Arguments = stored.Arguments
	command.CreatedAt = time.Now()

	execute(command)
}

func output(ctx *cli.Context) {
	if !ctx.Args().Present() {

		parrot.Info("Id must be provided!")
		return
	}

	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		// handle error
		parrot.Info("Id provided is not valid!")
		parrot.Info(err.Error())
		return
	}

	var command = repository.FindById(id)

	parrot.Info(command.Output)
}

func runCommand(ctx *cli.Context) {

	var command = Command{}
	command.Name = ctx.Args()[0]
	command.Arguments = strings.Join(ctx.Args().Tail(), " ")
	command.CreatedAt = time.Now()

	execute(command)
}

func listSettings(ctx *cli.Context) {
	parrot.Info("List of all the settings")
	parrot.Info("Command finished")
}

func foo(ctx *cli.Context) {
	parrot.Info("foooo")
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
