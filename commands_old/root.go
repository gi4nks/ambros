package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version will be set by build flags
	Version = "dev"

	// Root command
	RootCmd = &cobra.Command{
		Use:   "ambros",
		Short: "Ambros - A powerful CLI tool",
		Long:  `Ambros is a CLI application built with Go and Cobra.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to Ambros!")
			fmt.Println("Use 'ambros --help' for more information.")
		},
	}

	cfgFile string
	verbose bool
)

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ambros.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Local flags
	RootCmd.Flags().BoolP("version", "", false, "Show version information")
}

func initConfig() {
	if verbose {
		fmt.Println("Verbose mode enabled")
	}

	if cfgFile != "" {
		fmt.Printf("Using config file: %s\n", cfgFile)
	}
}
