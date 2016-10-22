package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gi4nks/quant"
	"gopkg.in/urfave/cli.v1"
)

var parrot = quant.NewParrot("ambros")

var settings = Settings{}
var repository = Repository{}

var pathUtils = quant.NewPathUtils()

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
}

func main() {
	readSettings()
	initDB()

	// -------------------
	app := cli.NewApp()
	app.Name = "ambros"
	app.Usage = "the personal command butler!!!!"
	app.Version = "0.2.0"
	app.Copyright = "gi4nks - 2016"

	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{"ru"},
			Usage:   "run a command, remember to run the command with -- before. (./ambros r -- ls -la)",
			Action:  CmdRun,
		},
		{
			Name:    "config",
			Aliases: []string{"cf"},
			Usage:   "list current configuration settings",
			Action:  CmdListSettings,
		},
		{
			Name:    "revive",
			Aliases: []string{"re"},
			Usage:   "revive ambros",
			Action:  CmdRevive,
		},
		/*{
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
		},*/
		{
			Name:    "stack",
			Aliases: []string{"st"},
			Usage:   "stack functionalities sub commands",
			Subcommands: []cli.Command{
				{
					Name:    "push",
					Aliases: []string{"pu"},
					Usage:   "pushs a command to the list of executed command withou executing it",
					Action:  CmdStackPush,
				},
				{
					Name:   "all",
					Usage:  "Get all the stacked commands",
					Action: CmdStackAll,
				},
				{
					Name:    "run",
					Aliases: []string{"sr"},
					Usage:   "pops a command from the list and executes it",
					Action:  CmdStackRunId,
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
			Name:    "store",
			Aliases: []string{"sr"},
			Usage:   "store functionalities sub commands",
			Subcommands: []cli.Command{
				{
					Name:    "id",
					Aliases: []string{"id"},
					Usage:   "stres an executed command ",
					Action:  CmdStoreId,
				},
				{
					Name:   "all",
					Usage:  "Get all the stacked commands",
					Action: CmdStackAll,
				},
				{
					Name:    "run",
					Aliases: []string{"sr"},
					Usage:   "pops a command from the list and executes it",
					Action:  CmdStackRunId,
				},
			},
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
func CmdRevive(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		parrot.Println("Ambros will reinitialize all statistics.")

		repository.BackupSchema()
		repository.DeleteSchema()
		repository.InitSchema()
	})
	return nil
}

func CmdLogs(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		var commands = repository.GetAllCommands()

		for _, c := range commands {
			parrot.Println(c.String())
		}
	})
	return nil
}

func CmdLogsById(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id")
			return
		}

		var command = repository.FindById(id)

		parrot.Println(command.String())
	})
	return nil
}

func CmdLast(ctx *cli.Context) error {
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
	return nil
}

func CmdRecall(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id")
			return
		}

		var stored = repository.FindById(id)

		var command = initializeCommand(stored.Name, stored.Arguments)

		executeCommand(&command)
		finalizeCommand(&command)
	})
	return nil
}

func CmdOutput(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		parrot.Debug("Output command invoked")

		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id")
			return
		}
		var command = repository.FindById(id)

		if command.Output != "" {
			parrot.Println(command.Output)
		}

		if command.Error != "" {
			parrot.Println(command.Error)
		}
	})
	return nil
}

func CmdStoreId(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		parrot.Debug("Output command invoked")

		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id")
			return
		}
		var command = repository.FindById(id)

		if command.Output != "" {
			parrot.Println(command.Output)
		}

		if command.Error != "" {
			parrot.Println(command.Error)
		}
	})
	return nil
}

func CmdStoreAll(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		var commands = repository.GetAllStoredCommands()

		for _, c := range commands {
			parrot.Println(c.String())
		}
	})
	return nil
}

func CmdStackPush(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		var command = initializeCommand(ctx.Args()[0], ctx.Args().Tail())
		pushCommand(&command)
	})
	return nil
}

func CmdStackAll(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		var commands = repository.GetAllStackedCommands()

		for _, c := range commands {
			parrot.Println(c.AsStackedCommand())
		}
	})
	return nil
}

func CmdStackRunId(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id stored in the stack")
			return
		}

		var stored = repository.FindInStackById(id)

		var command = initializeCommand(stored.Name, stored.Arguments)

		executeCommand(&command)
		finalizeCommand(&command)
	})
	return nil
}

func CmdRun(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		var command = initializeCommand(ctx.Args()[0], ctx.Args().Tail())
		executeCommand(&command)
		finalizeCommand(&command)
	})
	return nil
}

func CmdListSettings(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		buf := new(bytes.Buffer)
		json.Indent(buf, []byte(settings.String()), "", "  ")
		parrot.Println(buf)
	})
	return nil
}

func CmdExport(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		args, err := stringsFromArguments(ctx)
		check(err)

		var command = repository.FindById(args[0])

		fileHandle, _ := os.Create(args[1])
		writer := bufio.NewWriter(fileHandle)
		defer fileHandle.Close()

		if command.Output != "" {
			fmt.Fprintln(writer, command.Output)
		}

		if command.Error != "" {
			fmt.Fprintln(writer, command.Error)
		}

		writer.Flush()

		parrot.Println("Done!")
	})
	return nil
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

	parrot.Println("[" + command.ID + "]")
}

func pushCommand(command *Command) {
	command.TerminatedAt = time.Now()
	repository.Push(*command)

	parrot.Println("[" + command.ID + "]")
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

	stopOut := make(chan bool)
	stopErr := make(chan bool)

	scannerOutput := bufio.NewScanner(outputReader)
	go func(stop chan bool) {
		for scannerOutput.Scan() {
			parrot.Println(scannerOutput.Text())
			bufferOutput.WriteString(scannerOutput.Text() + "\n")
		}

		stop <- true
	}(stopOut)

	scannerError := bufio.NewScanner(errorReader)
	go func(stop chan bool) {
		for scannerError.Scan() {
			parrot.Println(scannerError.Text())
			bufferError.WriteString(scannerError.Text() + "\n")
		}

		stop <- true
	}(stopErr)

	<-stopOut
	<-stopErr

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
