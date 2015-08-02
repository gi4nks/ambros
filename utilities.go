package main

import (
	"crypto/rand"
	"encoding/json"
	"github.com/kardianos/osext"
	"os"
	"path/filepath"
)

func ExistsPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func CreatePath(path string) {
	os.Mkdir(ExecutableFolder()+string(filepath.Separator)+path, 0777)
}

func ExecutableFolder() string {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		parrot.Error("Warning", err)

		return ""
	}

	return folder
}

func AsJson(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}

func Random() string {

	var dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, 12)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}


func hTail(a []string) []string {
	if len(a) >= 2 {
		return []string(a)[1:]
	}
	return []string{}
}