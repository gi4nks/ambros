package main

import (
	"gopkg.in/urfave/cli.v1"
)

// List of functions
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
