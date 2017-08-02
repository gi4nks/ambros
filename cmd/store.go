package cmd

import (
	"strings"

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

			var cc = cmd.Flag("push").Value.String()

			Parrot.Println("> flag: ", cmd.Flag("push").Value.String())

			if cc != "" {

				c, as, err := commandFromArguments(strings.Split(cc, " "))

				if err != nil {
					Parrot.Println("Please provide a valid command")
					return
				}

				var command = initializeCommand(c, as)
				pushCommand(&command, true)
			}

			var sh = cmd.Flag("show").Changed
			if sh {
				var commands, err = Repository.GetAllStoredCommands()

				if err != nil {
					Parrot.Println("Commands not available in the store", err)
					return
				}

				for _, c := range commands {
					Parrot.Println(c.AsStoredCommand())
				}
			}

		})
	},
}

func init() {
	RootCmd.AddCommand(storeCmd)

	storeCmd.Flags().StringP("push", "p", "", "Pushed the given command to the store")
	storeCmd.Flags().BoolP("show", "s", false, "Shows all the commands in the store")

}
