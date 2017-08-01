package cmd

import (
	"github.com/spf13/cobra"
)

// storeCmd represents the output command
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store",
	Long:  `Store command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Store command invoked")

			c, as, err := commandFromArguments(args)

			if err != nil {
				Parrot.Println("Please provide a valid command")
				return
			}

			var command = initializeCommand(c, as)
			executeCommand(&command)
			finalizeCommand(&command)

			//Parrot.Println("> flag: ", cmd.Flag("store").Changed)
			if cmd.Flag("store").Changed == true {
				//Parrot.Println("Storing command")
				pushCommand(&command, false)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(storeCmd)

	storeCmd.Flags().BoolP("store", "s", false, "Store the results")

}
