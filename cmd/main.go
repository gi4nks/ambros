package main

import (
	"fmt"
	"os"

	"github.com/gi4nks/ambros/cmd/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
