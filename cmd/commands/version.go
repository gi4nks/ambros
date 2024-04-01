package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Ambros",
	Long:  `All software has versions. This is Ambros's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.5.0")
	},
}
