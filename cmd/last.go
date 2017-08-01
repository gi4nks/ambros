package cmd

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

			limit, err := intFromArguments(args)

			if err != nil {
				limit = Configuration.LastCountDefault
			}

			var commands = Repository.GetExecutedCommands(limit)

			for _, c := range commands {
				c.Print(Parrot)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(lastCmd)
}
