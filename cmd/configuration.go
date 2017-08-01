package cmd

import (
	"github.com/spf13/cobra"

	"bytes"
	"encoding/json"
)

// configurationCmd represents the configuration command
var configurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Configuration",
	Long:  `Configuration command`,
	Run: func(cmd *cobra.Command, args []string) {
		commandWrapper(args, func() {
			Parrot.Debug("Configuration command invoked")

			buf := new(bytes.Buffer)
			json.Indent(buf, []byte(Configuration.String()), "", "  ")
			Parrot.Println(buf)
		})
	},
}

func init() {
	RootCmd.AddCommand(configurationCmd)
}
