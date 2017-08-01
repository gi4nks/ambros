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

			id, err1 := stringFromArguments(args)
			if err1 != nil {
				Parrot.Println("Please provide a valid command id")
				return
			}

			var stored models.Command
			var err error

			if cmd.Flag("history").Changed == true {
				stored, err = Repository.FindInStoreById(id)
			} else {
				stored, err = Repository.FindById(id)
			}

			if err != nil {
				Parrot.Println("Id not available in the store (" + id + ")")
				return
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
