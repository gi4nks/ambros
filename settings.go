package main

import (
	"encoding/json"
	"io/ioutil"
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
	folder, err := pathUtils.ExecutableFolder()

	if err != nil {
		parrot.Error("Executable forlder error", err)
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

		parrot.Debug("folder: " + folder)
		parrot.Debug("file: " + asJson(sts.configs))

	}
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
		parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}
