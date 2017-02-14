package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"os"

	homedir "github.com/mitchellh/go-homedir"
)

type Configuration struct {
	RepositoryDirectory string
	RepositoryFile      string
	LastCountDefault    int
	DebugMode           bool
}

type Settings struct {
	configs Configuration
}

func (sts *Settings) LoadSettings() {
	folder, err := appDataFolder()
	if err == nil {
		err = os.MkdirAll(folder, 0700)
	}

	if err != nil {
		parrot.Error("Application data folder error", err)
	}

	file, err := ioutil.ReadFile(filepath.Join(folder, "conf.json"))

	if err != nil {
		sts.configs = Configuration{}
		sts.configs.RepositoryDirectory = ConstRepositoryDirectory
		sts.configs.RepositoryFile = ConstRepositoryFile
		sts.configs.LastCountDefault = ConstLastCountDefault
		sts.configs.DebugMode = ConstDebugMode

	} else {
		json.Unmarshal(file, &sts.configs)

		parrot.Debug("folder: " + folder)
		parrot.Debug("file: " + asJson(sts.configs))
	}
}

func (sts Settings) RepositoryDirectory() string {
	val, _ := homedir.Expand(sts.configs.RepositoryDirectory)
	return val
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
		parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}
