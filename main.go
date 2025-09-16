package main

import (
	"log"

	"github.com/gi4nks/ambros/v3/cmd/commands"
)

func main() {
	if err := commands.InitializeRepository(); err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	commands.Execute()
}
