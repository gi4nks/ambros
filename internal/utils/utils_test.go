package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/gi4nks/ambros/internal/utils"
	"github.com/gi4nks/quant"
)

func TestAsJson(t *testing.T) {
	// Create a new instance of Utilities
	u := utils.NewUtilities(quant.Parrot{})

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
	u := utils.NewUtilities(quant.Parrot{})

	// Test case: Check length of generated random string
	result := u.Random()
	if len(result) != 12 {
		t.Errorf("Random() returned string of unexpected length: got %d, want %d", len(result), 12)
	}
}

func TestTail(t *testing.T) {
	// Create a new instance of Utilities
	u := utils.NewUtilities(quant.Parrot{})

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
	// Create a new instance of Utilities
	u := utils.NewUtilities(quant.Parrot{})

	// Test case: No error
	noError := error(nil)
	u.Check(noError) // Ensure no panic or error

	// Test case: Error
	testError := json.Unmarshal([]byte("{"), nil)
	u.Check(testError) // Ensure no panic or error
}

func TestFatal(t *testing.T) {
	// Create a new instance of Utilities
	u := utils.NewUtilities(quant.Parrot{})

	// Test case: No error
	noError := error(nil)
	u.Fatal(noError) // Ensure no panic or error

	// Test case: Error
	testError := json.Unmarshal([]byte("{"), nil)
	// Fatal should panic, so we need to use a recover function
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Fatal() did not panic as expected")
		}
	}()
	u.Fatal(testError)
}
