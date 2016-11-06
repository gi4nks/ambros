package main

import (
	"encoding/json"
	"errors"
	//"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
)

type Repository struct {
	DB *bolt.DB
}

// HELPER FUNCTIONS
func repositoryFullName() string {
	return settings.RepositoryDirectory() + string(filepath.Separator) + settings.RepositoryFile()
}

//
func (r *Repository) InitDB() {
	var err error

	b, err := pathUtils.ExistsPath(settings.RepositoryDirectory())
	if err != nil {
		parrot.Println("Got error when reading repository directory", err)
	}

	if !b {
		pathUtils.CreatePath(settings.RepositoryDirectory())
	}

	r.DB, err = bolt.Open(repositoryFullName(), 0600, nil)
	if err != nil {
		parrot.Println("Got error creating repository directory", err)
	}
}

func (r *Repository) InitSchema() error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Commands"))
		if err != nil {
			parrot.Println("create bucket: ", err)
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("CommandsStored"))
		if err != nil {
			parrot.Println("create bucket: ", err)
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("CommandsIndex"))
		if err != nil {
			parrot.Println("create bucket: ", err)
			return err
		}

		return nil
	})

	return err
}

func (r *Repository) DeleteSchema(complete bool) error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte("Commands"))
		if err != nil {
			parrot.Println("delete bucket: ", err)
			return err
		}

		if complete {

			err = tx.DeleteBucket([]byte("CommandsStored"))
			if err != nil {
				parrot.Println("delete bucket: ", err)
				return err
			}
		}

		err = tx.DeleteBucket([]byte("CommandsIndex"))
		if err != nil {
			parrot.Println("delete bucket: ", err)
			return err
		}

		return nil
	})

	return err
}

func (r *Repository) CloseDB() {
	if err := r.DB.Close(); err != nil {
		parrot.Error("Error", err)
	}
}

func (r *Repository) BackupSchema() error {
	b, _ := pathUtils.ExistsPath(settings.RepositoryDirectory())
	if !b {
		return errors.New("Gadget repository path does not exist")
	}

	err := r.DB.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(repositoryFullName()+".bkp", 0600)
	})

	return err
}

// functionalities

func (r *Repository) Push(c Command) {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		cc, err := tx.CreateBucketIfNotExists([]byte("CommandsStored"))

		if err != nil {
			return err
		}

		encoded1, err := json.Marshal(c)
		if err != nil {
			return err
		}

		return cc.Put([]byte(c.ID), encoded1)
	})

	if err != nil {
		parrot.Println("Error inserting data:", err)
	}
}

func (r *Repository) Put(c Command) {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		cc, err := tx.CreateBucketIfNotExists([]byte("Commands"))

		if err != nil {
			return err
		}

		encoded1, err := json.Marshal(c)
		if err != nil {
			return err
		}

		err = cc.Put([]byte(c.ID), encoded1)

		if err != nil {
			return err
		}

		ii, err := tx.CreateBucketIfNotExists([]byte("CommandsIndex"))
		if err != nil {
			return err
		}

		return ii.Put([]byte(c.CreatedAt.Format(time.RFC3339)), []byte(c.ID))
	})

	if err != nil {
		parrot.Println("Error inserting data: ", err)
	}
}

func (r *Repository) findById(id string, collection string) Command {
	var command = Command{}

	err := r.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		v := b.Get([]byte(id))

		err := json.Unmarshal(v, &command)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		parrot.Println("No command found:", err)
	}

	return command
}

func (r *Repository) deleteById(id string, collection string) error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		return b.Delete([]byte(id))
	})

	return err
}

func (r *Repository) FindById(id string) Command {
	return r.findById(id, "Commands")
}

func (r *Repository) FindInStoreById(id string) Command {
	return r.findById(id, "CommandsStored")
}

func (r *Repository) DeleteStoredCommand(id string) error {
	return r.deleteById(id, "CommandsStored")
}

func (r *Repository) DeleteAllStoredCommands() error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte("CommandsStored"))
		if err != nil {
			parrot.Error("delete bucket: ", err)
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("CommandsStored"))
		if err != nil {
			parrot.Error("create bucket: ", err)
			return err
		}

		return nil
	})

	return err
}

func (r *Repository) getAllCommands(collection string) []Command {
	commands := []Command{}

	r.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var command = Command{}
			err := json.Unmarshal(v, &command)
			if err != nil {
				return err
			}

			commands = append(commands, command)
		}

		return nil
	})

	return commands
}

func (r *Repository) GetAllStoredCommands() []Command {
	return r.getAllCommands("CommandsStored")
}

func (r *Repository) GetAllCommands() []Command {
	return r.getAllCommands("Commands")
}

func (r *Repository) GetLimitCommands(limit int) []Command {
	commands := []Command{}

	r.DB.View(func(tx *bolt.Tx) error {
		cc := tx.Bucket([]byte("Commands"))
		ii := tx.Bucket([]byte("CommandsIndex")).Cursor()

		var i = limit

		for k, v := ii.Last(); k != nil && i > 0; k, v = ii.Prev() {
			var command = Command{}

			vv := cc.Get(v)

			err := json.Unmarshal(vv, &command)
			if err != nil {
				return err
			}
			commands = append(commands, command)

			i--
		}

		return nil
	})

	return commands
}

func (r *Repository) GetExecutedCommands(count int) []ExecutedCommand {
	commands := r.GetLimitCommands(count)

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
