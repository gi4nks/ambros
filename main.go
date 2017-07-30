package main

import (
	"fmt"
	"os"

	"github.com/gi4nks/ambros/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	/*

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
	*/

}

/*
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


*/
