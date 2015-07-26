package main

import (
	"encoding/json"
	"fmt"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/structs"
	"path/filepath"
	"strconv"
)

type Repository struct {
	DB       *db.DB
	commands *db.Col
}

// HELPER FUNCTIONS
func repositoryFullName() string {
	return settings.RepositoryDirectory() + string(filepath.Separator) + settings.RepositoryFile()
}

//

func (r *Repository) InitDB() {
	var err error

	b, err := ExistsPath(settings.RepositoryDirectory())
	if err != nil {
		parrot.Error("Got error when reading repository directory", err)
	}

	if !b {
		CreatePath(settings.RepositoryDirectory())
	}

	// (Create if not exist) open a database
	r.DB, err = db.OpenDB(repositoryFullName())
	if err != nil {
		parrot.Error("Got error creating repository directory", err)
	}
}

func (r *Repository) InitSchema() {
	// Drop (delete) collection "Commands"
	if err := r.DB.Drop("Commands"); err != nil {
		parrot.Error("Commands collection cannot be deleted", err)
	}

	// Create two collections: Commands and Votes
	if err := r.DB.Create("Commands"); err != nil {
		parrot.Error("Commands collection already exists", err)
	}

	r.commands = r.DB.Use("Commands")

	if err := r.commands.Index([]string{"command_id"}); err != nil {
		parrot.Error("Error", err)
	}
}

func (r *Repository) CloseDB() {
	if err := r.DB.Close(); err != nil {
		parrot.Error("Error", err)
	}
}

func (r *Repository) BackupSchema() {
	parrot.Info("Not implemented")
}

// functionalities

func (r *Repository) Put(c Command) {
	_, err := r.commands.Insert(structs.Map(c))
	if err != nil {
		parrot.Error("Error", err)
	}

}

/*
func (r *Repository) GetOne() Command {
	command := Command{}
	r.DB.First(&command)
	return command
}
*/

func (r *Repository) FindById(id string) Command {
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": "`+id+`", "in": ["command_id"]}]`), &query)

	queryResult := make(map[int]struct{})

	if err := db.EvalQuery(query, r.commands, &queryResult); err != nil {
		panic(err)
	}

	var command = Command{}
	for id := range queryResult {
		// To get query result document, simply read it
		readBack, err := r.commands.Read(id)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Query returned document %v\n", readBack)

		command.FromMap(readBack)
	}

	return command
}

func (r *Repository) GetAllCommands() []Command {
	commands := []Command{}

	var query interface{}
	queryResult := make(map[int]struct{})

	if err := db.EvalQuery(query, r.commands, &queryResult); err != nil {
		panic(err)
	}

	var command = Command{}
	for id := range queryResult {
		// To get query result document, simply read it
		readBack, err := r.commands.Read(id)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Query returned document %v\n", readBack)
		command.FromMap(readBack)

		commands = append(commands, command)
	}

	return commands
}

func (r *Repository) GetHistory(count int) []Command {
	commands := []Command{}
	//r.DB.Order("terminated_at desc").Find(&commands).Count(&count)
	return commands
}

func (r *Repository) GetExecutedCommands(count int) []ExecutedCommand {
	commands := []Command{}

	parrot.Info("Count is: " + strconv.Itoa(count))

	//r.DB.Limit(count).Order("terminated_at desc").Find(&commands)

	executedCommands := make([]ExecutedCommand, len(commands))

	for i := 0; i < len(commands); i++ {
		executedCommands[i] = commands[i].AsExecutedCommand(i)
	}

	return executedCommands
}

func extend(slice []Command, element Command) []Command {
	n := len(slice)
	if n == cap(slice) {
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]Command, len(slice), 2*len(slice)+1)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}

// Append appends the items to the slice.
// First version: just loop calling Extend.
func append(slice []Command, items ...Command) []Command {
	for _, item := range items {
		slice = extend(slice, item)
	}
	return slice
}
