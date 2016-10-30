package main

import (
	"gopkg.in/urfave/cli.v1"
)

func CmdStorePush(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		cmd, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command")
			return
		}

		args, _ := stringsTailFromArguments(ctx)

		var command = initializeCommand(cmd, args)
		pushCommand(&command, true)
	})
	return nil
}

func CmdStoreRunId(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id stored")
			return
		}

		var stored = repository.FindInStoreById(id)

		var command = initializeCommand(stored.Name, stored.Arguments)

		executeCommand(&command)
		finalizeCommand(&command)

		if ctx.Bool("store") {
			pushCommand(&command, false)
		}
	})
	return nil
}

func CmdStoreShow(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		var commands = repository.GetAllStoredCommands()

		for _, c := range commands {
			parrot.Println(c.AsStoredCommand())
		}
	})
	return nil
}

func CmdStoreDeleteById(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		id, err := stringFromArguments(ctx)
		if err != nil {
			parrot.Println("Please provide a valid command id stored in the store")
			return
		}

		err = repository.DeleteStoredCommand(id)

		if err != nil {
			parrot.Println("Deletion of stored commands failed")
			return
		}
	})
	return nil
}

func CmdStoreDeleteAll(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		err := repository.DeleteAllStoredCommands()

		if err != nil {
			parrot.Println("Deletion of stored commands failed")
			return
		}
	})
	return nil
}
