package cmd

import (
	"github.com/spf13/cobra"
)

// outputCmd represents the output command
var outputCmd = &cobra.Command{
	Use:   "output",
	Short: "Output",
	Long:  `Output command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Output command invoked")

			id, err1 := stringFromArguments(args)
			if err1 != nil {
				Parrot.Println("Please provide a valid command id")
				return
			}
			var command, err = Repository.FindById(id)

			if err != nil {
				Parrot.Println("Error retrieving command in the store ("+id+")", err)
				return
			}

			if command.Output != "" {
				Parrot.Println(command.Output)
			}

			if command.Error != "" {
				Parrot.Println(command.Error)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(outputCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// outputCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// outputCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
