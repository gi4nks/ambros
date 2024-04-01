package commands

import (
	"github.com/spf13/cobra"
)

// reviveCmd represents the output command
var reviveCmd = &cobra.Command{
	Use:   "revive",
	Short: "Revive",
	Long:  `Revive command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Revive command invoked")

			if cmd.Flag("complete").Changed == true {
				Parrot.Println("ambros will reinitialize all data.")

				Repository.BackupSchema()
				Repository.DeleteSchema(true)
				Repository.InitSchema()
			} else {
				Parrot.Println("ambros will reinitialize some data.")

				Repository.BackupSchema()
				Repository.DeleteSchema(false)
				Repository.InitSchema()
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(reviveCmd)
	reviveCmd.Flags().BoolP("complete", "c", false, "Complete revival of ambros")
}
