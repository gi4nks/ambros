package repos

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
	models "github.com/gi4nks/ambros/internal/models"
	utils "github.com/gi4nks/ambros/internal/utils"
	"github.com/gi4nks/quant"
)

type Repository struct {
	parrot        *quant.Parrot
	configuration *utils.Configuration

	DB *bolt.DB
}

func NewRepository(p quant.Parrot, c utils.Configuration) *Repository {
	return &Repository{parrot: &p, configuration: &c}
}

func (r *Repository) InitDB() error {
	var err error

	b, err := quant.ExistsPath(r.configuration.RepositoryDirectory)
	if err != nil {
		return errors.New("Ambros repository path does not exist, please ckeck if following path exists: " + r.configuration.RepositoryDirectory)
	}

	if !b {
		quant.CreatePath(r.configuration.RepositoryDirectory)
	}

	r.DB, err = bolt.Open(r.configuration.RepositoryFullName(), 0600, nil)
	if err != nil {
		return errors.New("Ambros was not able to open db: please check if following path exists: " + r.configuration.RepositoryFullName())
	}

	//r.parrot.Println(r.DB)
	return nil
}

func (r *Repository) InitSchema() error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Commands"))
		if err != nil {
			//r.parrot.Println(">err", err)
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("CommandsStored"))
		if err != nil {
			//r.parrot.Println(">err", err)
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("CommandsIndex"))
		if err != nil {
			//r.parrot.Println(">err", err)
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
			return err
		}

		if complete {

			err = tx.DeleteBucket([]byte("CommandsStored"))
			if err != nil {
				return err
			}
		}

		err = tx.DeleteBucket([]byte("CommandsIndex"))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *Repository) CloseDB() error {
	if err := r.DB.Close(); err != nil {
		return errors.New("Error closing DB")
	}
	return nil
}

func (r *Repository) BackupSchema() error {
	b, _ := quant.ExistsPath(r.configuration.RepositoryDirectory)
	if !b {
		return errors.New("Ambros repository path does not exist")
	}

	err := r.DB.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(r.configuration.RepositoryFullName()+".bkp", 0600)
	})

	return err
}

// functionalities

func (r *Repository) Push(c models.Command) error {
	return r.DB.Update(func(tx *bolt.Tx) error {
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
}

func (r *Repository) Put(c models.Command) error {
	return r.DB.Update(func(tx *bolt.Tx) error {
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
}

func (r *Repository) findById(id string, collection string) (models.Command, error) {
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

	return command, err
}

func (r *Repository) deleteById(id string, collection string) error {
	return r.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		return b.Delete([]byte(id))
	})
}

func (r *Repository) FindById(id string) (models.Command, error) {
	return r.findById(id, "Commands")
}

func (r *Repository) FindInStoreById(id string) (models.Command, error) {
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

func (r *Repository) getAllCommands(collection string) ([]models.Command, error) {
	commands := []models.Command{}

	err := r.DB.View(func(tx *bolt.Tx) error {
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

	return commands, err
}

func (r *Repository) GetAllStoredCommands() ([]models.Command, error) {
	return r.getAllCommands("CommandsStored")
}

func (r *Repository) GetAllCommands() ([]models.Command, error) {
	return r.getAllCommands("Commands")
}

func (r *Repository) GetLimitCommands(limit int) ([]models.Command, error) {
	commands := []models.Command{}

	err := r.DB.View(func(tx *bolt.Tx) error {
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

	return commands, err
}

func (r *Repository) GetExecutedCommands(count int) ([]models.ExecutedCommand, error) {
	commands, err := r.GetLimitCommands(count)

	executedCommands := make([]models.ExecutedCommand, len(commands))

	for i := 0; i < len(commands); i++ {
		executedCommands[i] = commands[i].AsExecutedCommand(i)
	}

	return executedCommands, err
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
