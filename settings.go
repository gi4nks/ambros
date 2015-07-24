package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Configuration struct {
	RepositoryDirectory string
	RepositoryFile      string
	RepositoryLogMode   bool

	HistoryCountDefault int
	LastCountDefault    int
}

type Settings struct {
	Configs Configuration
}

func (sts *Settings) LoadSettings() {
	parrot.Info("Load settings if provided")

	file, err := ioutil.ReadFile("/home/gianluca/Projects/golang/bin/conf.json")
	if err != nil {
		parrot.Error("Warning", err)
		os.Exit(1)
	}
	//m := new(Dispatch)
	//var m interface{}
	json.Unmarshal(file, &sts.Configs)
}

func RepositoryDirectory(sts *Settings) string {
	if sts.Configs == (Configuration{}) {
		return ConstRepositoryDirectory
	}

	if sts.Configs.RepositoryDirectory == "" {
		return ConstRepositoryDirectory
	}

	return sts.Configs.RepositoryDirectory
}

func RepositoryFile(sts *Settings) string {
	if sts.Configs == (Configuration{}) {
		return ConstRepositoryFile
	}

	if sts.Configs.RepositoryFile == "" {
		return ConstRepositoryFile
	}

	return sts.Configs.RepositoryFile
}

func RepositoryLogMode(sts *Settings) bool {
	if sts.Configs == (Configuration{}) {
		return ConstRepositoryLogMode
	}

	return sts.Configs.RepositoryLogMode
}

func HistoryCountDefault(sts *Settings) int {
	if sts.Configs == (Configuration{}) {
		return ConstHistoryCountDefault
	}

	return sts.Configs.HistoryCountDefault
}

func LastCountDefault(sts *Settings) int {
	if sts.Configs == (Configuration{}) {
		return ConstLastCountDefault
	}

	return sts.Configs.LastCountDefault
}
