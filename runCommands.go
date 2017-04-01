package main

/*
import (
	"gopkg.in/urfave/cli.v1"
)

// StoreFlag prints the version for the application
var StoreFlag = cli.BoolFlag{
	Name:  "store, s",
	Usage: "store the command after execution",
}

var HistoryFlag = cli.BoolFlag{
	Name:  "history, y",
	Usage: "select the command from store",
}

func CmdRun(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		cmd, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command")
			return
		}

		args, _ := stringsTailFromArguments(ctx)

		var command = initializeCommand(cmd, args)
		executeCommand(&command)
		finalizeCommand(&command)

		if ctx.Bool("store") {
			pushCommand(&command, false)
		}
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

		var stored Command

		if ctx.Bool("history") {
			stored = repository.FindInStoreById(id)
		} else {
			stored = repository.FindById(id)
		}

		var command = initializeCommand(stored.Name, stored.Arguments)

		executeCommand(&command)
		finalizeCommand(&command)
	})
	return nil
}
*/
