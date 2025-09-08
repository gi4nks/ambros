package utils_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gi4nks/ambros/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsJson(t *testing.T) {
	// Create a new instance of Utilities
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	u := utils.NewUtilities(logger)

	// Test case: Valid input
	input := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}
	expected := `{"age":30,"email":"john@example.com","name":"John"}`
	result := u.AsJson(input)
	if result != expected {
		t.Errorf("AsJson() returned unexpected result: got %s, want %s", result, expected)
	}

	// Test case: Error case (non-serializable input)
	nonSerializableInput := make(chan int)
	expectedError := "{}"
	result = u.AsJson(nonSerializableInput)
	if result != expectedError {
		t.Errorf("AsJson() with non-serializable input returned unexpected result: got %s, want %s", result, expectedError)
	}
}

func TestRandom(t *testing.T) {
	// Create a new instance of Utilities
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	tests := []struct {
		name    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid random string",
			wantLen: 12,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := utils.NewUtilities(logger)
			got := u.Random()
			assert.Equal(t, tt.wantLen, len(got))
		})
	}
}

func TestTail(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	// Create a new instance of Utilities
	u := utils.NewUtilities(logger)

	// Test case: Slice with more than 2 elements
	input := []string{"a", "b", "c"}
	expected := []string{"b", "c"}
	result := u.Tail(input)
	if len(result) != len(expected) || result[0] != expected[0] || result[1] != expected[1] {
		t.Errorf("Tail() returned unexpected result: got %v, want %v", result, expected)
	}

	// Test case: Slice with less than 2 elements
	input = []string{"a"}
	expected = []string{}
	result = u.Tail(input)
	if len(result) != len(expected) {
		t.Errorf("Tail() returned unexpected result: got %v, want %v", result, expected)
	}
}

func TestCheck(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	// Create a new instance of Utilities
	u := utils.NewUtilities(logger)

	// Test case: No error
	noError := error(nil)
	u.Check(noError) // Ensure no panic or error

	// Test case: Error
	var dummy interface{}
	testError := json.Unmarshal([]byte("{"), &dummy)
	u.Check(testError) // Ensure no panic or error
}

func TestFatal(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	// Create a new instance of Utilities
	u := utils.NewUtilities(logger)

	// Test case: No error
	noError := error(nil)
	u.Fatal(noError) // Ensure no panic or error

	// Test case: Error
	var dummy interface{}
	testError := json.Unmarshal([]byte("{"), &dummy)
	// Fatal should panic, so we need to use a recover function
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Fatal() did not panic as expected")
		}
	}()
	u.Fatal(testError)
}

func TestUtilities_ExistsPath(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	// Create a new instance of Utilities
	u := utils.NewUtilities(logger)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "utils-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing directory",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "non-existing path",
			path:     filepath.Join(tmpDir, "nonexistent"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := u.ExistsPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUtilities_CreatePath(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	// Create a new instance of Utilities
	u := utils.NewUtilities(logger)

	// Create a temporary base directory
	tmpDir, err := os.MkdirTemp("", "utils-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create new directory",
			path:    filepath.Join(tmpDir, "newdir"),
			wantErr: false,
		},
		{
			name:    "create nested directories",
			path:    filepath.Join(tmpDir, "nested", "dirs"),
			wantErr: false,
		},
		{
			name:    "create in readonly directory",
			path:    "/root/test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.CreatePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, u.ExistsPath(tt.path))
		})
	}
}

func TestUtilities_GetAbsolutePath(t *testing.T) {
	f := setupTest(t)
	defer f.tearDown()
	logger := f.logger

	u := utils.NewUtilities(logger)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "current directory",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "test/path",
			wantErr: false,
		},
		{
			name:    "absolute path",
			path:    "/tmp/test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs, err := u.GetAbsolutePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, filepath.IsAbs(abs))
		})
	}
}
