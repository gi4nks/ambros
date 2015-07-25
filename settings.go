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
		sts.configs = Configuration{}
		sts.configs.RepositoryDirectory = ConstRepositoryDirectory
		sts.configs.RepositoryFile = ConstRepositoryFile
		sts.configs.RepositoryLogMode = ConstRepositoryLogMode
		sts.configs.HistoryCountDefault = ConstHistoryCountDefault
		sts.configs.LastCountDefault = ConstLastCountDefault

	} else {
		json.Unmarshal(file, &sts.configs)
	}
}

func (sts Settings) RepositoryDirectory() string {
	return sts.configs.RepositoryDirectory
}

func (sts Settings) RepositoryFile() string {
	return sts.configs.RepositoryFile
}

func (sts Settings) RepositoryLogMode() bool {
	return sts.configs.RepositoryLogMode
}

func (sts Settings) HistoryLimitDefault() int {
	return sts.configs.HistoryCountDefault
}

func (sts Settings) LastLimitDefault() int {
	return sts.configs.LastCountDefault
}

func (sts Settings) String() string {
	b, err := json.Marshal(sts.configs)
	if err != nil {
		parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}
