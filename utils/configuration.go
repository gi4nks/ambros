package utils

import (
	"encoding/json"
	"path/filepath"

	"github.com/gi4nks/quant"
)

type Configuration struct {
	parrot *quant.Parrot

	RepositoryDirectory string
	RepositoryFile      string
	LastCountDefault    int
	DebugMode           bool
}

func NewConfiguration(p quant.Parrot) *Configuration {
	var c = Configuration{}
	c.parrot = &p

	c.RepositoryDirectory = ConstRepositoryDirectory
	c.RepositoryFile = ConstRepositoryFile
	c.LastCountDefault = ConstLastCountDefault
	c.DebugMode = ConstDebugMode

	return &c
}

func (c Configuration) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		c.parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}

func (c Configuration) RepositoryFullName() string {

	/*c.parrot.Println("1", c.RepositoryDirectory)
	c.parrot.Println("2", string(filepath.Separator))
	c.parrot.Println("3", c.RepositoryFile)*/
	return c.RepositoryDirectory + string(filepath.Separator) + c.RepositoryFile
}
