package cmd

import (
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Logs",
	Long:  `Logs command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Output command invoked")

			var id = cmd.Flag("id").Value.String()

			if id != "" {
				var command, err = Repository.FindById(id)

				if err != nil {
					Parrot.Println("Error retrieving command in the store ("+id+")", err)
					return
				}

				Parrot.Println(command.String())
			} else {
				var commands, err = Repository.GetAllCommands()

				if err != nil {
					Parrot.Println("Error retrieving commands in the store", err)
					return
				}

				for _, c := range commands {
					Parrot.Println(c.String())
				}
			}

		})
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringP("id", "i", "", "id to show the logs")
}
