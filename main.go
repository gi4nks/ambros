package main

import (
	"bufio"
	"bytes"
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
	app.Version = "0.2.2"
	app.Copyright = "gi4nks - 2016"

	//app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:    "settings",
			Aliases: []string{"st"},
			Usage:   "settings sub menu",
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage:   "list current configuration settings",
					Action:  CmdSettingsList,
				},
				{
					Name:    "revive",
					Aliases: []string{"re"},
					Usage:   "revive ambros",
					Subcommands: []cli.Command{
						{
							Name:    "complete",
							Aliases: []string{"com"},
							Usage:   "revive ambros deleting completely the database",
							Action:  CmdSettingsReviveComplete,
						},
						{
							Name:    "partial",
							Aliases: []string{"par"},
							Usage:   "revive ambros preserving the stored commands",
							Action:  CmdSettingsRevivePartial,
						},
					},
				},
				{
					Name:    "logs",
					Aliases: []string{"lo"},
					Usage:   "show me the logs of ambros",
					Subcommands: []cli.Command{
						{
							Name:    "id",
							Aliases: []string{"id"},
							Usage:   "Get the log of specific id",
							Action:  CmdSettingsLogsById,
						},
						{
							Name:    "all",
							Aliases: []string{"all"},
							Usage:   "Get all the logs",
							Action:  CmdSettingsLogs,
						},
					},
				},
			},
		},
		{
			Name:    "run",
			Aliases: []string{"ru"},
			Usage:   "run a command, remember to run the command with -- before",
			Action:  CmdRun,
			Flags: []cli.Flag{
				StoreFlag,
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
			Aliases: []string{"so"},
			Usage:   "store functionalities sub commands",
			Subcommands: []cli.Command{
				{
					Name:    "id",
					Aliases: []string{"id"},
					Usage:   "stores an executed command ",
					Action:  CmdStoreRunId,
				},
				{
					Name:    "push",
					Aliases: []string{"pu"},
					Usage:   "push a not executed command in the stack",
					Action:  CmdStorePush,
				},
				{
					Name:    "show",
					Aliases: []string{"sh"},
					Usage:   "Show all the stored commands",
					Action:  CmdStoreShow,
				},
				{
					Name:  "delete",
					Usage: "Deletes the stored commands",
					Subcommands: []cli.Command{
						{
							Name:    "id",
							Aliases: []string{"id"},
							Usage:   "deletes specific command from the store",
							Action:  CmdStoreDeleteById,
						},
						{
							Name:    "all",
							Aliases: []string{"all"},
							Usage:   "deletes all stored commands",
							Action:  CmdStoreDeleteAll,
						},
					},
				},
				{
					Name:    "run",
					Aliases: []string{"ru"},
					Usage:   "pops a command from the list and executes it",
					Action:  CmdStoreRunId,
					Flags: []cli.Flag{
						StoreFlag,
					},
				},
			},
		},
		{
			Name:    "recall",
			Aliases: []string{"rc"},
			Usage:   "recall a command and executes it again",
			Action:  CmdRecall,
			Flags: []cli.Flag{
				HistoryFlag,
			},
		},
		{
			Name:    "export",
			Aliases: []string{"ex"},
			Usage:   "exports the output of a command to a file",
			Action:  CmdExport,
			Flags: []cli.Flag{
				HistoryFlag,
			},
		},
	}

	app.Run(os.Args)

	defer closeDB()
}

// List of functions
func CmdExport(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		args, err := stringsFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id stored and valid output file")
			return
		}

		if len(args) != 2 {
			parrot.Println("Wrong number of arguments")
			return
		}

		id := args[0]
		fl := args[1]

		var stored Command
		if ctx.Bool("history") {
			stored = repository.FindInStoreById(id)
		} else {
			stored = repository.FindById(id)
		}

		fileHandle, _ := os.Create(fl)
		writer := bufio.NewWriter(fileHandle)
		defer fileHandle.Close()

		if stored.Output != "" {
			fmt.Fprintln(writer, stored.Output)
		}

		if stored.Error != "" {
			fmt.Fprintln(writer, stored.Error)
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
func stringsTailFromArguments(ctx *cli.Context) ([]string, error) {
	if !ctx.Args().Present() {
		return nil, errors.New("Value must be provided!")
	}

	str := ctx.Args().Tail()

	return str, nil
}

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

func pushCommand(command *Command, showid bool) {
	command.TerminatedAt = time.Now()
	repository.Push(*command)

	if showid {
		parrot.Println("[" + command.ID + "]")
	}
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
