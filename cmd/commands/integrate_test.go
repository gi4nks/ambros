package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestIntegrateInstallUninstall(t *testing.T) {
	tmp, err := ioutil.TempDir("", "ambros-integ-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	// set HOME to tmp so install writes into temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", oldHome)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ic := NewIntegrateCommand(logger)

	// Run install non-interactively targeting default shells
	if err := ic.install(nil, nil, "", true); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	target := filepath.Join(tmp, ".ambros-integration.sh")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Fatalf("expected integration script at %s", target)
	}

	bashrc := filepath.Join(tmp, ".bashrc")
	data, _ := ioutil.ReadFile(bashrc)
	if string(data) == "" || !containsLine(string(data), "source ~/.ambros-integration.sh") {
		t.Fatalf("bashrc not updated")
	}

	// Now uninstall
	if err := ic.uninstall(nil, nil, "", true); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected integration script removed")
	}
}
