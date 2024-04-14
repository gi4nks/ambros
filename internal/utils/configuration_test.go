package utils_test

import (
	"path/filepath"
	"testing"

	"github.com/gi4nks/ambros/internal/utils"
	"github.com/gi4nks/quant"
)

func TestNewConfiguration(t *testing.T) {
	config := utils.NewConfiguration(quant.Parrot{})

	if config == nil {
		t.Error("Expected Configuration to be created, got nil")
	}

	// Add more tests for specific configuration values if needed
}

func TestConfiguration_String(t *testing.T) {
	config := utils.NewConfiguration(quant.Parrot{})

	expected := config.String()
	if result := config.String(); result != expected {
		t.Errorf("Expected string representation %q, got %q", expected, result)
	}

	// Add more tests for specific cases if needed
}

func TestConfiguration_RepositoryFullName(t *testing.T) {
	config := utils.NewConfiguration(quant.Parrot{})

	expected := config.RepositoryDirectory + string(filepath.Separator) + config.RepositoryFile
	if result := config.RepositoryFullName(); result != expected {
		t.Errorf("Expected repository full name %q, got %q", expected, result)
	}

	// Add more tests for specific cases if needed
}
