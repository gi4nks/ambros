package utils_test

import (
	"path/filepath"
	"testing"

	"github.com/gi4nks/ambros/internal/utils"
)

func TestNewConfiguration(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	config := utils.NewConfiguration(logger)

	if config == nil {
		t.Error("Expected Configuration to be created, got nil")
	}

	// Add more tests for specific configuration values if needed
}

func TestConfiguration_RepositoryFullName(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	config := utils.NewConfiguration(logger)

	expected := config.RepositoryDirectory + string(filepath.Separator) + config.RepositoryFile
	if result := config.RepositoryFullName(); result != expected {
		t.Errorf("Expected repository full name %q, got %q", expected, result)
	}

	// Add more tests for specific cases if needed
}
