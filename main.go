package main

import (
	"bufio"
	"bytes"
	"github.com/codegangsta/cli"
	"github.com/gi4nks/quant"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var tracer = quant.NewTrace("ambros")

var settings = Settings{}
var repository = Repository{}

func configureTracer() {
	// Configuring logger
	tracer.Light()
}

func configureDB() {
	repository.InitDB()
}

func readSettings() {
	settings.LoadSettings()
}

func main() {
	readSettings()
	configureTracer()
	configureDB()

	// -------------------
	app := cli.NewApp()
	app.Name = "ambros"
	app.Usage = "the personal command butler!!!!"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "run a command, remember to run the command with -- before. (./butler r -- ls -la)",
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
			Usage:   "revive butler",
			Action:  revive,
		},
		{
			Name:    "logs",
			Aliases: []string{"o"},
			Usage:   "show me the logs of butler",
			Action:  logs,
		},
		{
			Name:    "history",
			Aliases: []string{"y"},
			Usage:   "show the history of butler. followed with a valid number shows nummber of history ",
			Action:  history,
		},
		{
			Name:    "last",
			Aliases: []string{"ll"},
			Usage:   "show the last executed commands ",
			Action:  last,
		},
	}

	app.Run(os.Args)
}

// List of functions
func revive(ctx *cli.Context) {
	tracer.Notice("Revive butler!!!!")
	tracer.Warning("Reviving butler will reinitialize all the statistics.")

	repository.BackupSchema()

	repository.InitSchema()
}

func logs(ctx *cli.Context) {
	tracer.Notice("Butler Logs")

	var commands = repository.GetAllCommands()

	for _, c := range commands {
		//TODO structure the output n a readable way
		tracer.News(c.String())
	}
}

func history(ctx *cli.Context) {
	tracer.Notice("Butler History")

	var count = settings.Configs.HistoryCountDefault
	var err error
	if ctx.Args().Present() {

		count, err = strconv.Atoi(ctx.Args()[0])
		if err != nil {
			// handle error
			tracer.Warning(err.Error())
			count = settings.Configs.HistoryCountDefault
		}
	}

	var commands = repository.GetHistory(count)

	for _, c := range commands {
		//TODO structure the output n a readable way
		tracer.News(c.AsHistory())
	}
}

func last(ctx *cli.Context) {
	tracer.Notice("Amdros Last")

	var count = settings.Configs.LastCountDefault
	var err error
	if ctx.Args().Present() {

		count, err = strconv.Atoi(ctx.Args()[0])
		if err != nil {
			// handle error
			tracer.Warning(err.Error())
			count = settings.Configs.LastCountDefault
		}
	}

	var commands = repository.GetExecutedCommands(count)

	for _, c := range commands {
		//TODO structure the output n a readable way
		tracer.News(c.String())
	}
}

func prepareEnvironment(ctx *cli.Context) {
	tracer.Notice("Prepare the environment!!!!")
}

func runCommand(ctx *cli.Context) {
	tracer.Notice("Run a command in the environment")

	var buffer bytes.Buffer

	var command = Command{}
	command.Name = ctx.Args()[0]
	command.Arguments = asJson(ctx.Args().Tail())
	command.CreatedAt = time.Now()

	cmd := exec.Command(ctx.Args()[0], ctx.Args().Tail()...)

	// Logging configuration
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		tracer.Error(err.Error())
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		tracer.Error(err.Error())
	}

	// read command's stdout line by line
	in := bufio.NewScanner(stdout)

	for in.Scan() {
		var ss = in.Text()
		tracer.News(ss) // write each line to your log, or anything you need
		buffer.WriteString(ss)
	}
	if err := in.Err(); err != nil {
		tracer.Error(err.Error())
	}

	command.TerminatedAt = time.Now()
	command.Output = buffer.String()
	command.Status = "Completed with SUCCESS"

	repository.Put(command)
	// ----------------------
	tracer.Notice("Command finished with error: " + err.Error())
}

func listSettings(ctx *cli.Context) {
	tracer.Notice("List of all the settings")
	tracer.News(asJson(settings.Configs))
	tracer.Notice("Command finished")
}
