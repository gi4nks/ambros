package main

/*
import (
	"bytes"
	"encoding/json"

	"gopkg.in/urfave/cli.v1"
)

func CmdSettingsReviveComplete(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		parrot.Println("ambros will reinitialize all statistics.")

		repository.BackupSchema()
		repository.DeleteSchema(true)
		repository.InitSchema()
	})
	return nil
}

func CmdSettingsRevivePartial(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		parrot.Println("ambros will reinitialize some statistics.")

		repository.BackupSchema()
		repository.DeleteSchema(false)
		repository.InitSchema()
	})
	return nil
}

func CmdSettingsList(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		buf := new(bytes.Buffer)
		json.Indent(buf, []byte(settings.String()), "", "  ")
		parrot.Println(buf)
	})
	return nil
}

func CmdSettingsLogs(ctx *cli.Context) error {
	commandWrapper(ctx, func() {
		var commands = repository.GetAllCommands()

		for _, c := range commands {
			parrot.Println(c.String())
		}
	})
	return nil
}

func CmdSettingsLogsById(ctx *cli.Context) error {
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
*/
