package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/boltdb/bolt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gi4nks/ambros/internal/models"
)

func main() {
	src := flag.String("src", "", "Source BoltDB file path")
	dst := flag.String("dst", "", "Destination BadgerDB directory")
	flag.Parse()

	if *src == "" || *dst == "" {
		log.Fatal("Source and destination paths are required")
	}

	// Open source BoltDB
	srcDB, err := bolt.Open(*src, 0600, nil)
	if err != nil {
		log.Fatalf("Failed to open source db: %v", err)
	}
	defer srcDB.Close()

	// Open destination BadgerDB
	opts := badger.DefaultOptions(*dst)
	dstDB, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to open destination db: %v", err)
	}
	defer dstDB.Close()

	// Migrate data
	err = migrateData(srcDB, dstDB)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully")
}

func migrateData(srcDB *bolt.DB, dstDB *badger.DB) error {
	wb := dstDB.NewWriteBatch()
	defer wb.Cancel()

	return srcDB.View(func(tx *bolt.Tx) error {
		// Migrate Commands bucket
		if b := tx.Bucket([]byte("Commands")); b != nil {
			if err := migrateBucket(b, wb, "cmd:"); err != nil {
				return err
			}
		}

		// Migrate CommandsStored bucket
		if b := tx.Bucket([]byte("CommandsStored")); b != nil {
			if err := migrateBucket(b, wb, "stored:"); err != nil {
				return err
			}
		}

		// Migrate CommandsIndex bucket
		if b := tx.Bucket([]byte("CommandsIndex")); b != nil {
			if err := migrateBucket(b, wb, "time:"); err != nil {
				return err
			}
		}

		return wb.Flush()
	})
}

func migrateBucket(b *bolt.Bucket, wb *badger.WriteBatch, prefix string) error {
	return b.ForEach(func(k, v []byte) error {
		// For Commands and CommandsStored, we need to validate the JSON
		if prefix == "cmd:" || prefix == "stored:" {
			var cmd models.Command
			if err := json.Unmarshal(v, &cmd); err != nil {
				return err
			}
		}

		return wb.Set(append([]byte(prefix), k...), v)
	})
}
