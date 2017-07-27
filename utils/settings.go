package utils

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	utils "github.com/gi4nks/ambros/utils"
	"github.com/gi4nks/quant/parrot"
	"github.com/gi4nks/quant/paths"
)

type Configuration struct {
	RepositoryDirectory string
	RepositoryFile      string
	LastCountDefault    int
	DebugMode           bool
}

type Settings struct {
	parrot    *parrot.Parrot
	utilities *Utilities
	configs   Configuration
}

func NewSettings(p parrot.Parrot, u Utilities) *Settings {
	return &Settings{parrot: &p, utilities: &u}
}

func (sts *Settings) LoadSettings() {
	folder, err := paths.ExecutableFolder()

	if err != nil {
		sts.parrot.Error("Executable forlder error", err)
	}

	file, err := ioutil.ReadFile(folder + "/conf.json")

	if err != nil {
		sts.configs = Configuration{}
		sts.configs.RepositoryDirectory = folder + "/" + ConstRepositoryDirectory
		sts.configs.RepositoryFile = ConstRepositoryFile
		sts.configs.LastCountDefault = ConstLastCountDefault
		sts.configs.DebugMode = ConstDebugMode

	} else {
		json.Unmarshal(file, &sts.configs)

		//sts.parrot.Println("> folder: " + folder)
		//sts.parrot.Println("> file: " + sts.utilities.AsJson(sts.configs))
	}

	//sts.parrot.Println("> config", sts.configs)
}

func (sts Settings) RepositoryDirectory() string {
	return sts.configs.RepositoryDirectory
}

func (sts Settings) RepositoryFile() string {
	return sts.configs.RepositoryFile
}

func (sts Settings) LastLimitDefault() int {
	return sts.configs.LastCountDefault
}

func (sts Settings) DebugMode() bool {
	return sts.configs.DebugMode
}

func (sts Settings) String() string {
	b, err := json.Marshal(sts.configs)
	if err != nil {
		sts.parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}

func (sts Settings) RepositoryFullName() string {

	sts.parrot.Println("1", sts.RepositoryDirectory())
	sts.parrot.Println("2", string(filepath.Separator))
	sts.parrot.Println("3", sts.RepositoryFile())
	return sts.RepositoryDirectory() + string(filepath.Separator) + sts.RepositoryFile()
}
