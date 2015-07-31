package main

import (
	"encoding/json"
	"time"
	"fmt"
	"github.com/boltdb/bolt"
	"path/filepath"
	"strconv"
	"github.com/bradhe/stopwatch"
)

type Repository struct {
	DB       *bolt.DB
}

// HELPER FUNCTIONS
func repositoryFullName() string {
	return settings.RepositoryDirectory() + string(filepath.Separator) + settings.RepositoryFile()
}

//

func (r *Repository) InitDB() {
	start := stopwatch.Start()

	var err error

	b, err := ExistsPath(settings.RepositoryDirectory())
	if err != nil {
		parrot.Error("Got error when reading repository directory", err)
	}

	if !b {
		CreatePath(settings.RepositoryDirectory())
	}

	r.DB, err = bolt.Open(repositoryFullName(), 0600, nil)
    if err != nil {
        parrot.Error("Got error creating repository directory", err)
    }
    defer r.DB.Close()

	watch := stopwatch.Stop(start)
    fmt.Printf("Milliseconds elappsed: %v\n", watch.Milliseconds())
}

func (r *Repository) InitSchema() error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
	    b, err := tx.CreateBucket([]byte("Commands"))
	    if err != nil {
	        return fmt.Errorf("create bucket: %s", err)
	    }
	    return nil
	})
	
	return err
}
	
	//defer r.DB.Close()

func (r *Repository) CloseDB() {
	if err := r.DB.Close(); err != nil {
		parrot.Error("Error", err)
	}
}

func (r *Repository) BackupSchema() {
	/*
	// Drop (delete) collection "Commands"
	if err := r.DB.Drop("Commands.bkp"); err != nil {
		parrot.Error("Commands.bkp collection cannot be deleted", err)
	}

	if err := r.DB.Rename("Commands", "Commands.bkp"); err != nil {
		parrot.Error("Commands.bkp collection cannot be created", err)
	}
	*/
}

// functionalities

func (r *Repository) Put(c Command) error {
	err := db.Update(func(tx *bolt.Tx) error {
	    b, err := tx.CreateBucketIfNotExists([]byte("Commands"))
	    if err != nil {
	        return err
	    }
	    encoded, err := json.Marshal(c)
	    if err != nil {
	        return err
	    }
	    return b.Put([]byte(c.CreatedAt.Format(time.RFC3339)), encoded)
	})	
	return err
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
		parrot.Error("Error", err)
	}

	var command = Command{}
	for id := range queryResult {
		// To get query result document, simply read it
		readBack, err := r.commands.Read(id)
		if err != nil {
			parrot.Error("Error", err)
		}
		parrot.Info("Query returned document: ") //, readBack)

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
	commands := r.GetAllCommands()

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
