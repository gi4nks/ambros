package cmd

import (
	"github.com/spf13/cobra"
)

// runCmd represents the output command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run",
	Long:  `Run command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Run command invoked")

			c, as, err := commandFromArguments(args)

			if err != nil {
				Parrot.Println("Please provide a valid command")
				return
			}

			var command = initializeCommand(c, as)
			executeCommand(&command)
			finalizeCommand(&command)

			if cmd.Flag("store").Changed == true {
				pushCommand(&command, false)
			}

			/*
				if ctx.Bool("store") {
					pushCommand(&command, false)
				}
			*/
		})
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// outputCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// outputCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().BoolP("store", "s", false, "HStore the results")

}
