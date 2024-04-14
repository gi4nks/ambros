package commands

import (
	models "github.com/gi4nks/ambros/internal/models"
	"github.com/spf13/cobra"
)

// runCmd represents the output command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run",
	Long:  `Run command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Run command invoked")

			cmds, err := commandsFromArguments(args)

			if err != nil {
				Parrot.Println("Please provide a valid command")
				return
			}

			var commands = initializeCommands(cmds)

			var commandPointers []*models.Command
			for i := range commands {
				commandPointers = append(commandPointers, &commands[i])
			}

			// Now call executeCommands with []*models.Command
			executeCommands(commandPointers)

			/*
				var command = initializeCommand(c, as)
				executeCommand(&command)
				finalizeCommand(&command)
			*/

			//Parrot.Println("> flag: ", cmd.Flag("store").Changed)
			if cmd.Flag("store").Changed {
				Parrot.Println("Storing command")
				pushCommands(commandPointers, false)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolP("store", "s", false, "Store the results")

}
