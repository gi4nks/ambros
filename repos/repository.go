package repos

import (
	"encoding/json"
	"errors"
	//"os"
	"time"

	"github.com/boltdb/bolt"
	models "github.com/gi4nks/ambros/models"
	utils "github.com/gi4nks/ambros/utils"
	"github.com/gi4nks/quant/parrot"
	"github.com/gi4nks/quant/paths"
)

type Repository struct {
	parrot   *parrot.Parrot
	settings *utils.Settings

	DB *bolt.DB
}

func NewRepository(p parrot.Parrot, sts utils.Settings) *Repository {
	return &Repository{parrot: &p, settings: &sts}
}

//
func (r *Repository) InitDB() {
	var err error

	r.parrot.Println("sts 1", r.settings)

	r.parrot.Println("sts 1", r.settings)

	b, err := paths.ExistsPath(r.settings.RepositoryDirectory())
	if err != nil {
		r.parrot.Println("Got error when reading repository directory", err)
	}

	if !b {
		paths.CreatePath(r.settings.RepositoryDirectory())
	}

	r.parrot.Println("--", r.settings.RepositoryFullName())

	r.DB, err = bolt.Open(r.settings.RepositoryFullName(), 0600, nil)
	if err != nil {
		r.parrot.Println("Got error creating repository directory", err)
	}

	r.parrot.Println(r.DB)
}

func (r *Repository) InitSchema() error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Commands"))
		if err != nil {
			r.parrot.Println("create bucket: ", err)
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("CommandsStored"))
		if err != nil {
			r.parrot.Println("create bucket: ", err)
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("CommandsIndex"))
		if err != nil {
			r.parrot.Println("create bucket: ", err)
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
			r.parrot.Println("delete bucket: ", err)
			return err
		}

		if complete {

			err = tx.DeleteBucket([]byte("CommandsStored"))
			if err != nil {
				r.parrot.Println("delete bucket: ", err)
				return err
			}
		}

		err = tx.DeleteBucket([]byte("CommandsIndex"))
		if err != nil {
			r.parrot.Println("delete bucket: ", err)
			return err
		}

		return nil
	})

	return err
}

func (r *Repository) CloseDB() {
	if err := r.DB.Close(); err != nil {
		r.parrot.Error("Error", err)
	}
}

func (r *Repository) BackupSchema() error {
	b, _ := paths.ExistsPath(r.settings.RepositoryDirectory())
	if !b {
		return errors.New("Gadget repository path does not exist")
	}

	err := r.DB.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(r.settings.RepositoryFullName()+".bkp", 0600)
	})

	return err
}

// functionalities

func (r *Repository) Push(c models.Command) {
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
		r.parrot.Println("Error inserting data:", err)
	}
}

func (r *Repository) Put(c models.Command) {
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
		r.parrot.Println("Error inserting data: ", err)
	}
}

func (r *Repository) findById(id string, collection string) models.Command {
	var command = models.Command{}

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
		r.parrot.Println("No command found:", err)
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

func (r *Repository) FindById(id string) models.Command {
	return r.findById(id, "Commands")
}

func (r *Repository) FindInStoreById(id string) models.Command {
	return r.findById(id, "CommandsStored")
}

func (r *Repository) DeleteStoredCommand(id string) error {
	return r.deleteById(id, "CommandsStored")
}

func (r *Repository) DeleteAllStoredCommands() error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte("CommandsStored"))
		if err != nil {
			r.parrot.Error("delete bucket: ", err)
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("CommandsStored"))
		if err != nil {
			r.parrot.Error("create bucket: ", err)
			return err
		}

		return nil
	})

	return err
}

func (r *Repository) getAllCommands(collection string) []models.Command {
	commands := []models.Command{}

	r.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var command = models.Command{}
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

func (r *Repository) GetAllStoredCommands() []models.Command {
	return r.getAllCommands("CommandsStored")
}

func (r *Repository) GetAllCommands() []models.Command {
	return r.getAllCommands("Commands")
}

func (r *Repository) GetLimitCommands(limit int) []models.Command {
	commands := []models.Command{}

	r.DB.View(func(tx *bolt.Tx) error {
		cc := tx.Bucket([]byte("Commands"))
		ii := tx.Bucket([]byte("CommandsIndex")).Cursor()

		var i = limit

		for k, v := ii.Last(); k != nil && i > 0; k, v = ii.Prev() {
			var command = models.Command{}

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

func (r *Repository) GetExecutedCommands(count int) []models.ExecutedCommand {
	commands := r.GetLimitCommands(count)

	executedCommands := make([]models.ExecutedCommand, len(commands))

	for i := 0; i < len(commands); i++ {
		executedCommands[i] = commands[i].AsExecutedCommand(i)
	}

	return executedCommands
}

func (r *Repository) extend(slice []models.Command, element models.Command) []models.Command {
	n := len(slice)
	if n == cap(slice) {
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]models.Command, len(slice), 2*len(slice)+1)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}

// Append appends the items to the slice.
// First version: just loop calling Extend.
func (r *Repository) append(slice []models.Command, items ...models.Command) []models.Command {
	for _, item := range items {
		slice = r.extend(slice, item)
	}
	return slice
}
