package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/models"
)

type Repository struct {
	logger *zap.Logger
	db     *badger.DB
	dbPath string
}

// NewRepository creates a new repository instance
func NewRepository(dbPath string, logger *zap.Logger) (*Repository, error) {
	repo := &Repository{
		logger: logger,
		dbPath: dbPath,
	}

	if err := repo.initDB(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) initDB() error {
	// Ensure directory exists
	dir := filepath.Dir(r.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	opts := badger.DefaultOptions(r.dbPath)
	opts.Logger = nil // Disable Badger's internal logger

	var err error
	r.db, err = badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	return nil
}

func (r *Repository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Put stores a command in the repository
func (r *Repository) Put(ctx context.Context, c models.Command) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		// Store command by ID
		if err := txn.Set([]byte("cmd:"+c.ID), data); err != nil {
			return err
		}

		// Store index by timestamp for ordering
		timeKey := fmt.Sprintf("time:%s:%s", c.CreatedAt.Format("20060102T150405.999999999Z0700"), c.ID)
		return txn.Set([]byte(timeKey), []byte(c.ID))
	})
}

// Get retrieves a command by ID
func (r *Repository) Get(id string) (*models.Command, error) {
	var command models.Command

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("cmd:" + id))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &command)
		})
	})

	if err != nil {
		return nil, err
	}
	return &command, nil
}

// FindById retrieves a command by ID (legacy method)
func (r *Repository) FindById(id string) (models.Command, error) {
	cmd, err := r.Get(id)
	if err != nil {
		return models.Command{}, err
	}
	return *cmd, nil
}

// GetLimitCommands retrieves the most recent commands up to the specified limit
func (r *Repository) GetLimitCommands(limit int) ([]models.Command, error) {
	var commands []models.Command

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		for it.Rewind(); it.Valid() && count < limit; it.Next() {
			item := it.Item()
			key := string(item.Key())

			// Skip non-time keys
			if !strings.HasPrefix(key, "time:") {
				continue
			}

			var cmdID []byte
			err := item.Value(func(val []byte) error {
				cmdID = append([]byte{}, val...)
				return nil
			})
			if err != nil {
				continue
			}

			cmdItem, err := txn.Get([]byte("cmd:" + string(cmdID)))
			if err != nil {
				continue
			}

			err = cmdItem.Value(func(val []byte) error {
				var cmd models.Command
				if err := json.Unmarshal(val, &cmd); err != nil {
					return err
				}
				commands = append(commands, cmd)
				return nil
			})
			if err != nil {
				continue
			}

			count++
		}
		return nil
	})

	return commands, err
}

// GetAllCommands retrieves all commands from the repository
func (r *Repository) GetAllCommands() ([]models.Command, error) {
	var commands []models.Command

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("cmd:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var command models.Command
				if err := json.Unmarshal(val, &command); err != nil {
					return err
				}
				commands = append(commands, command)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return commands, err
}

// SearchByTag searches for commands by tag
func (r *Repository) SearchByTag(tag string) ([]models.Command, error) {
	allCommands, err := r.GetAllCommands()
	if err != nil {
		return nil, err
	}

	var filtered []models.Command
	for _, cmd := range allCommands {
		for _, cmdTag := range cmd.Tags {
			if strings.EqualFold(cmdTag, tag) {
				filtered = append(filtered, cmd)
				break
			}
		}
	}
	return filtered, nil
}

// SearchByStatus searches for commands by status
func (r *Repository) SearchByStatus(success bool) ([]models.Command, error) {
	allCommands, err := r.GetAllCommands()
	if err != nil {
		return nil, err
	}

	var filtered []models.Command
	for _, cmd := range allCommands {
		if cmd.Status == success {
			filtered = append(filtered, cmd)
		}
	}
	return filtered, nil
}

// GetTemplate retrieves a command template by name
func (r *Repository) GetTemplate(name string) (*models.Template, error) {
	// This is a placeholder implementation
	// Templates could be stored in a separate key prefix
	return nil, fmt.Errorf("template not found: %s", name)
}

// Push stores a command for future use (bookmark/favorite)
func (r *Repository) Push(command models.Command) error {
	// Store as a "stored" command for bookmarking
	data, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("stored:"+command.ID), data)
	})
}

// Legacy/Additional methods for backward compatibility

// BackupSchema creates a backup of the database
func (r *Repository) BackupSchema() error {
	backupFile := r.dbPath + ".bkp"
	f, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer f.Close()

	_, err = r.db.Backup(f, 0)
	if err != nil {
		return fmt.Errorf("failed to backup database: %w", err)
	}

	return nil
}

// DeleteSchema deletes all data from the database
func (r *Repository) DeleteSchema(complete bool) error {
	if err := r.db.DropAll(); err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}
	return nil
}

// GetAllStoredCommands retrieves all stored/bookmarked commands
func (r *Repository) GetAllStoredCommands() ([]models.Command, error) {
	var commands []models.Command

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("stored:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				var cmd models.Command
				if err := json.Unmarshal(val, &cmd); err != nil {
					return err
				}
				commands = append(commands, cmd)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return commands, err
}

// FindInStoreById retrieves a stored command by ID
func (r *Repository) FindInStoreById(id string) (models.Command, error) {
	var command models.Command

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("stored:" + id))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &command)
		})
	})

	return command, err
}

// Delete removes a command by ID
func (r *Repository) Delete(id string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// First, get the command to get its timestamp for time index removal
		cmdKey := []byte("cmd:" + id)
		item, err := txn.Get(cmdKey)
		if err != nil {
			return err
		}

		var cmd models.Command
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &cmd)
		})
		if err != nil {
			return err
		}

		// Delete the main command entry
		if err := txn.Delete(cmdKey); err != nil {
			return err
		}

		// Delete the time index entry
		timeKey := []byte("time:" + cmd.CreatedAt.Format(time.RFC3339Nano))
		if err := txn.Delete(timeKey); err != nil {
			// Continue even if time index deletion fails
		}

		// Delete any tag entries for this command
		for _, tag := range cmd.Tags {
			tagKey := []byte("tag:" + tag + ":" + id)
			if err := txn.Delete(tagKey); err != nil {
				// Continue even if tag deletion fails
			}
		}

		return nil
	})
}

// DeleteStoredCommand removes a stored command
func (r *Repository) DeleteStoredCommand(id string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("stored:" + id))
	})
}

// DeleteAllStoredCommands removes all stored commands
func (r *Repository) DeleteAllStoredCommands() error {
	return r.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("stored:")
		it := txn.NewIterator(opts)
		defer it.Close()

		var keysToDelete [][]byte
		for it.Rewind(); it.Valid(); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysToDelete = append(keysToDelete, key)
		}

		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetExecutedCommands converts commands to executed command format
func (r *Repository) GetExecutedCommands(count int) ([]models.ExecutedCommand, error) {
	commands, err := r.GetLimitCommands(count)
	if err != nil {
		return nil, err
	}

	executedCommands := make([]models.ExecutedCommand, len(commands))
	for i := 0; i < len(commands); i++ {
		executedCommands[i] = commands[i].AsExecutedCommand(i)
	}

	return executedCommands, nil
}
