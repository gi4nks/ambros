package cmd

import (
	"github.com/spf13/cobra"

	models "github.com/gi4nks/ambros/models"
)

// recallCmd represents the output command
var recallCmd = &cobra.Command{
	Use:   "recall",
	Short: "Recall",
	Long:  `Recall command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Recall command invoked")

			id, err := stringFromArguments(args)
			if err != nil {
				Parrot.Println("Please provide a valid command id")
				return
			}

			var stored models.Command

			if cmd.Flag("history").Changed == true {
				stored = Repository.FindInStoreById(id)
			} else {
				stored = Repository.FindById(id)
			}

			var command = initializeCommand(stored.Name, stored.Arguments)

			executeCommand(&command)
			finalizeCommand(&command)

			if cmd.Flag("store").Changed == true {
				//Parrot.Println("Storing command")
				pushCommand(&command, false)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(recallCmd)
	recallCmd.Flags().BoolP("history", "y", false, "Recalls a command from history")
	recallCmd.Flags().BoolP("store", "s", false, "Store the results")
}
