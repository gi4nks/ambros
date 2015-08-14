package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/codegangsta/cli"
	"github.com/gi4nks/quant"
	"os"
	"os/exec"
	"strconv"
	"time"
	"encoding/json"
	"io/ioutil"
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

	if settings.DebugMode() {
		parrot = quant.NewVerboseParrot("ambros")
	}

	parrot.Debug("Parrot is set to talk so much!")

}

func main() {
	readSettings()
	initDB()

	// -------------------
	app := cli.NewApp()
	app.Name = "ambros"
	app.Usage = "the personal command butler!!!!"
	app.Version = "0.0.1"
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
			Aliases: []string{"rc"},
			Usage:   "recall a command and execute again",
			Action:  CmdRecall,
		},
		{
			Name:    "export",
			Aliases: []string{"ex"},
			Usage:   "exports the output of a command to a file",
			Action:  CmdExport,
		},
	}

	app.Run(os.Args)

	defer closeDB()
}

// List of functions
func CmdRevive(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		parrot.Info("==> Reviving ambros will reinitialize all the statistics.")

		repository.BackupSchema()
		repository.InitSchema()
	})
}

func CmdLogs(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		var commands = repository.GetAllCommands()

		for _, c := range commands {
			parrot.Info(c.String())
		}
	})
}

func CmdLogsById(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Error("Error...", err)
			return
		}

		var command = repository.FindById(id)

		parrot.Info(command.String())
	})
}

func CmdLast(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		limit, err := intFromArguments(ctx)
		if err != nil {
			limit = settings.LastLimitDefault()
		}

		var commands = repository.GetExecutedCommands(limit)

		for _, c := range commands {
			c.Print()
		}
	})
}

func CmdRecall(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Error("Error...", err)
			return
		}

		var stored = repository.FindById(id)

		var command = initializeCommand(stored.Name, stored.Arguments)

		executeCommand(&command)
		finalizeCommand(&command)
	})
}

func CmdOutput(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		parrot.Debug("Output command invoked")

		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Error("Error...", err)
			return
		}

		parrot.Debug("==> id: " + id)

		var command = repository.FindById(id)

		if command.Output != "" {
			parrot.Info("==> Output:")
			parrot.Info(command.Output)
		}

		if command.Error != "" {
			parrot.Info("==> Error:")
			parrot.Info(command.Error)
		}
	})
}

func CmdRun(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		var command = initializeCommand(ctx.Args()[0], ctx.Args().Tail())
		executeCommand(&command)
		finalizeCommand(&command)
	})
}

func CmdListSettings(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		buf := new(bytes.Buffer)
		json.Indent(buf, []byte(settings.String()), "", "  ")
		parrot.Println(buf)	
	})
}

func CmdExport(ctx *cli.Context) {
	commandWrapper(ctx, func() {
		args, err := stringsFromArguments(ctx)
		check(err)
	
		var command = repository.FindById(args[0])
		
		d1 := []byte(command.Output)
	    err = ioutil.WriteFile(args[1], d1, 0644)
	    check(err)
		
		parrot.Info(command.String())
	})
}

func CmdWrapper(ctx *cli.Context) {
}

// ----------------
// Arguments from command string
// ----------------
func stringsFromArguments(ctx *cli.Context) ([]string, error) {
	if !ctx.Args().Present() {
		return nil, errors.New("Value must be provided!")
	}

	str := ctx.Args()

	return str, nil
}

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

// ----------------
// command management
// ----------------
func initializeCommand(name string, arguments []string) Command {
	var command = Command{}
	command.ID = random()

	command.Name = name
	command.Arguments = arguments

	command.CreatedAt = time.Now()
	return command
}

func finalizeCommand(command *Command) {
	command.TerminatedAt = time.Now()
	repository.Put(*command)

	parrot.Info("[" + command.ID + "]")

}

func executeCommand(command *Command) {
	var bufferOutput bytes.Buffer
	var bufferError bytes.Buffer

	cmd := exec.Command(command.Name, command.Arguments...)

	parrot.Debug("--> CommandName " + command.Name)
	parrot.Debug("--> Command Arguments " + asJson(command.Arguments))

	outputReader, err := cmd.StdoutPipe()
	if err != nil {
		parrot.Error("Error creating StdoutPipe for Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	errorReader, err := cmd.StderrPipe()
	if err != nil {
		parrot.Error("Error creating StderrPipe for Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	err = cmd.Start()
	if err != nil {
		parrot.Error("Error starting Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	scannerOutput := bufio.NewScanner(outputReader)
	go func() {
		for scannerOutput.Scan() {
			parrot.Info(scannerOutput.Text())
			bufferOutput.WriteString(scannerOutput.Text() + "\n")
		}
	}()

	scannerError := bufio.NewScanner(errorReader)
	go func() {
		for scannerError.Scan() {
			parrot.Info(scannerError.Text())
			bufferError.WriteString(scannerError.Text() + "\n")
		}
	}()

	err = cmd.Wait()
	if err != nil {
		parrot.Error("Error waiting for Cmd", err)
		command.Error = err.Error()
		command.Status = false
		return
	}

	command.Output = bufferOutput.String()
	command.Error = bufferError.String()

	command.Status = true
}

// -------------------------------
// Cli command wrapper
// -------------------------------
func commandWrapper(ctx *cli.Context, cmd quant.Action0) {
	CmdWrapper(ctx)

	cmd()
}
