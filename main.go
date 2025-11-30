package main

import (
	"log"
	"os"

	"github.com/gi4nks/ambros/v3/cmd/commands"
)

func main() {
	if err := commands.InitializeRepository(); err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	pc := commands.GetPluginCommand()
	if pc != nil {
		// Execute pre-run hooks
		commands.ExecuteHooks(pc, "pre-run", os.Args[1:])
	}

	commands.Execute()

	if pc != nil {
		// Execute post-run hooks
		commands.ExecuteHooks(pc, "post-run", os.Args[1:])
	}
}
