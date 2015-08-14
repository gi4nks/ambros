package main

import (
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
)

func existsPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func createPath(path string) {
	os.Mkdir(executableFolder()+string(filepath.Separator)+path, 0777)
}

func executableFolder() string {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		parrot.Error("Warning", err)

		return ""
	}

	return folder
}

func asJson(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}

func random() string {

	var dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, 12)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func tail(a []string) []string {
	if len(a) >= 2 {
		return []string(a)[1:]
	}
	return []string{}
}

func check(e error) {
	if e != nil {
		parrot.Error("Error...", e)
		return
	}
}

func fatal(e error) {
	if e != nil {
		parrot.Error("Fatal...", e)
		panic(e)
	}
}
