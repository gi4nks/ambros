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
		})
	},
}

func init() {
	RootCmd.AddCommand(storeCmd)

	storeCmd.Flags().StringP("push", "p", "", "Pushed the given command to the store")

}
