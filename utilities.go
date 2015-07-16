package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func create(path string) {
	os.Mkdir("."+string(filepath.Separator)+path, 0777)
}

func asJson(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		tracer.Warning(err.Error())
		return "{}"
	}
	return string(b)
}
