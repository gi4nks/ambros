package main

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	RepositoryDirectory string
	RepositoryFile      string
	RepositoryLogMode   bool

	HistoryCountDefault int
	LastCountDefault    int
}

type Settings struct {
	configs Configuration
}

func (sts *Settings) LoadSettings() {
	folder := ExecutableFolder()

	file, err := ioutil.ReadFile(folder + "/conf.json")
	if err != nil {
		parrot.Error("Warning", err)
		parrot.Info("==> Using default values")
		sts.configs = Configuration{}
		return
	}
	json.Unmarshal(file, &sts.configs)
}

func (sts Settings) RepositoryDirectory() string {
	if sts.configs == (Configuration{}) {
		return ConstRepositoryDirectory
	}

	if sts.configs.RepositoryDirectory == "" {
		return ConstRepositoryDirectory
	}

	return sts.configs.RepositoryDirectory
}

func (sts Settings) RepositoryFile() string {
	if sts.configs == (Configuration{}) {
		return ConstRepositoryFile
	}

	if sts.configs.RepositoryFile == "" {
		return ConstRepositoryFile
	}

	return sts.configs.RepositoryFile
}

func (sts Settings) RepositoryLogMode() bool {
	if sts.configs == (Configuration{}) {
		return ConstRepositoryLogMode
	}

	return sts.configs.RepositoryLogMode
}

func (sts Settings) HistoryCountDefault() int {
	if sts.configs == (Configuration{}) {
		return ConstHistoryCountDefault
	}

	return sts.configs.HistoryCountDefault
}

func (sts Settings) LastCountDefault() int {
	if sts.configs == (Configuration{}) {
		return ConstLastCountDefault
	}

	return sts.configs.LastCountDefault
}
