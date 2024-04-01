package commands

import (
	"github.com/spf13/cobra"
)

// lastCmd represents the output command
var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Last",
	Long:  `Last command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Last command invoked")

			limit, err1 := intFromArguments(args)

			if err1 != nil {
				limit = Configuration.LastCountDefault
			}

			var commands, err = Repository.GetExecutedCommands(limit)

			if err != nil {
				Parrot.Println("Error retrieving commands in the store", err)
				return
			}

			for _, c := range commands {
				c.Print(Parrot)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(lastCmd)
}
