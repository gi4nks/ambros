package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type Configuration struct {
	logger              *zap.Logger
	RepositoryDirectory string
	RepositoryFile      string
	LastCountDefault    int
	DebugMode           bool
}

func NewConfiguration(logger *zap.Logger) *Configuration {
	c := &Configuration{
		logger: logger,
	}

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get home directory", zap.Error(err))
		c.RepositoryDirectory = filepath.Join(os.TempDir(), ".ambros")
	} else {
		c.RepositoryDirectory = filepath.Join(home, ".ambros")
	}

	c.RepositoryFile = "ambros.db"
	return c
}

func (c *Configuration) RepositoryFullName() string {
	return filepath.Join(c.RepositoryDirectory, c.RepositoryFile)
}

func (c *Configuration) String() string {
	return fmt.Sprintf(`{
	"repositoryDirectory": "%s",
	"repositoryFile": "%s",
	"lastCountDefault": %d,
	"debugMode": %t
}`, c.RepositoryDirectory, c.RepositoryFile, c.LastCountDefault, c.DebugMode)
}
