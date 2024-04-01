package commands

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

			if cc != "" {

				c, as, err := commandFromArguments(strings.Split(cc, " "))

				if err != nil {
					Parrot.Println("Please provide a valid command")
					return
				}

				var command = initializeCommand(c, as)
				pushCommand(&command, true)
				return
			}

			var sh = cmd.Flag("show").Changed
			if sh {
				var commands, err = Repository.GetAllStoredCommands()

				if err != nil {
					Parrot.Println("Commands not available in the store", err)
					return
				}

				if len(commands) > 0 {

					for _, c := range commands {
						Parrot.Println(c.AsStoredCommand())
					}
				} else {
					Parrot.Println("No commands available in the store!")
				}
				return
			}

			var rid = cmd.Flag("run").Value.String()

			if rid != "" {
				var stored, err = Repository.FindInStoreById(rid)
				if err != nil {
					Parrot.Println("Command ("+rid+") not available in the store", err)
					return
				}

				var command = initializeCommand(stored.Name, stored.Arguments)

				executeCommand(&command)
				finalizeCommand(&command)

				return
			}

			var did = cmd.Flag("delete").Value.String()

			if did != "" {
				var err = Repository.DeleteStoredCommand(did)
				if err != nil {
					Parrot.Println("Command ("+did+") not available in the store", err)
					return
				}
				Parrot.Println("Done!")
				return
			}

			var cl = cmd.Flag("clear").Changed
			if cl {
				err := Repository.DeleteAllStoredCommands()

				if err != nil {
					Parrot.Println("Deletion of stored commands failed")
					return
				}

				Parrot.Println("Done!")
				return
			}

		})
	},
}

func init() {
	RootCmd.AddCommand(storeCmd)

	storeCmd.Flags().StringP("push", "p", "", "Pushed the given command to the store")
	storeCmd.Flags().StringP("run", "r", "", "Run a command stored in the store")
	storeCmd.Flags().StringP("delete", "d", "", "Delete a command stored from the store")

	storeCmd.Flags().BoolP("show", "s", false, "Shows all the commands in the store")
	storeCmd.Flags().BoolP("clear", "c", false, "Removes all the commands in the store")

}
