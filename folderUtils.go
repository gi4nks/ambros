package main

import (
	"os"
	"path/filepath"

	"runtime"

	homedir "github.com/mitchellh/go-homedir"
)

func appDataFolder() (string, error) {
	if runtime.GOOS == "windows" {
		// First try the %APPDATA% directory
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "Ambros"), nil
		}
	}
	// %APPDATA% is not usable, use the home folder with hidden ".ambros" directory
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ambros"), nil
}
