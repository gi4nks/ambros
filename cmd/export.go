package cmd

import (
	"github.com/spf13/cobra"

	"bufio"
	"fmt"
	"os"

	models "github.com/gi4nks/ambros/models"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export",
	Long:  `Export command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Output command invoked")

			if len(args) != 2 {
				Parrot.Println("Please provide a valid command id stored and valid output file")
				return
			}

			id := args[0]
			fl := args[1]

			var stored models.Command
			var err error
			if cmd.Flag("history").Changed == true {
				stored, err = Repository.FindInStoreById(id)
			} else {
				stored, err = Repository.FindById(id)
			}

			if err != nil {
				Parrot.Println("Id not available in the store (" + id + ")")
				return
			}

			fileHandle, err := os.Create(fl)

			if err != nil {
				Parrot.Println("Impossible to create the required file (" + fl + ")")
				return
			}

			writer := bufio.NewWriter(fileHandle)
			defer fileHandle.Close()

			if stored.Output != "" {
				fmt.Fprintln(writer, stored.Output)
			}

			if stored.Error != "" {
				fmt.Fprintln(writer, stored.Error)
			}

			writer.Flush()

			Parrot.Println("Done!")
		})
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)
	exportCmd.Flags().BoolP("history", "y", false, "Recalls a command from history")

}
