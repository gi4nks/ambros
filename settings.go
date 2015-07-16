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
	tracer.Notice("Load settings if provided")

	file, e := ioutil.ReadFile("./conf.json")
	if e != nil {
		tracer.Warning("File error: " + e.Error())
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
