package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gi4nks/ambros/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCommand_Clone(t *testing.T) {
	now := time.Now()
	original := &models.Command{
		Entity: models.Entity{
			ID:           "test-id",
			CreatedAt:    now,
			TerminatedAt: now.Add(time.Hour),
		},
		Name:      "test-command",
		Arguments: []string{"arg1", "arg2"},
		Status:    true,
		Output:    "test output",
		Error:     "test error",
	}

	clone := original.Clone()

	assert.Equal(t, original.ID, clone.ID)
	assert.Equal(t, original.CreatedAt, clone.CreatedAt)
	assert.Equal(t, original.TerminatedAt, clone.TerminatedAt)
	assert.Equal(t, original.Name, clone.Name)
	assert.Equal(t, original.Arguments, clone.Arguments)
	assert.Equal(t, original.Status, clone.Status)
	assert.Equal(t, original.Output, clone.Output)
	assert.Equal(t, original.Error, clone.Error)

	// Verify deep copy of slices
	clone.Arguments[0] = "modified"
	assert.NotEqual(t, original.Arguments[0], clone.Arguments[0])
}

func TestCommand_String(t *testing.T) {
	cmd := &models.Command{
		Entity: models.Entity{
			ID:           "test-id",
			CreatedAt:    time.Now(),
			TerminatedAt: time.Now(),
		},
		Name:      "test",
		Arguments: []string{"arg1", "arg2"},
		Status:    true,
		Output:    "output",
		Error:     "",
	}

	str, err := cmd.String()
	assert.NoError(t, err)

	var unmarshaled models.Command
	err = json.Unmarshal([]byte(str), &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, cmd.ID, unmarshaled.ID)
}

func TestCommand_AsHistory(t *testing.T) {
	cmd := &models.Command{
		Name:      "test",
		Arguments: []string{"arg1", "arg2"},
	}

	history := cmd.AsHistory()
	expected := "Name : test --> Arguments : arg1 arg2"
	assert.Equal(t, expected, history)
}

func TestCommand_AsExecutedCommand(t *testing.T) {
	now := time.Now()
	cmd := &models.Command{
		Entity: models.Entity{
			ID:        "test-id",
			CreatedAt: now,
		},
		Name:      "test",
		Arguments: []string{"arg1", "arg2"},
		Status:    true,
	}

	executed := cmd.AsExecutedCommand(1)
	assert.Equal(t, 1, executed.Order)
	assert.Equal(t, cmd.ID, executed.ID)
	assert.Equal(t, "test arg1 arg2", executed.Command)
	assert.Equal(t, cmd.Status, executed.Status)
	assert.Equal(t, cmd.CreatedAt, executed.When)
}

func TestCommand_ToMap(t *testing.T) {
	now := time.Now()
	cmd := &models.Command{
		Entity: models.Entity{
			ID:           "test-id",
			CreatedAt:    now,
			TerminatedAt: now.Add(time.Hour),
		},
		Name:      "test",
		Arguments: []string{"arg1"},
		Status:    true,
		Output:    "output",
		Error:     "error",
	}

	m := cmd.ToMap()
	assert.Equal(t, cmd.ID, m["ID"])
	assert.Equal(t, cmd.Name, m["Name"])
	assert.Equal(t, cmd.Arguments, m["Arguments"])
	assert.Equal(t, cmd.Status, m["Status"])
	assert.Equal(t, cmd.Output, m["Output"])
	assert.Equal(t, cmd.Error, m["Error"])
	assert.Equal(t, cmd.CreatedAt, m["CreatedAt"])
	assert.Equal(t, cmd.TerminatedAt, m["TerminatedAt"])
}

func TestCommand_FromMap(t *testing.T) {
	now := time.Now()
	m := map[string]interface{}{
		"ID":           "test-id",
		"Name":         "test",
		"Arguments":    []string{"arg1"},
		"Status":       true,
		"Output":       "output",
		"Error":        "error",
		"CreatedAt":    now,
		"TerminatedAt": now.Add(time.Hour),
	}

	var cmd models.Command
	cmd.FromMap(m)

	assert.Equal(t, m["ID"], cmd.ID)
	assert.Equal(t, m["Name"], cmd.Name)
	assert.Equal(t, m["Arguments"], cmd.Arguments)
	assert.Equal(t, m["Status"], cmd.Status)
	assert.Equal(t, m["Output"], cmd.Output)
	assert.Equal(t, m["Error"], cmd.Error)
	assert.Equal(t, m["CreatedAt"], cmd.CreatedAt)
	assert.Equal(t, m["TerminatedAt"], cmd.TerminatedAt)
}

func TestCommand_AsStoredCommand(t *testing.T) {
	cmd := &models.Command{
		Entity: models.Entity{
			ID: "test-id",
		},
		Name:      "test",
		Arguments: []string{"arg1", "arg2"},
	}

	stored := cmd.AsStoredCommand()
	expected := "[test-id] test arg1 arg2"
	assert.Equal(t, expected, stored)
}

func TestExecutedCommand_AsFlatCommand(t *testing.T) {
	now := time.Now()
	executed := models.ExecutedCommand{
		Order:   1,
		ID:      "test-id",
		Command: "test arg1",
		Status:  true,
		When:    now,
	}

	flat := executed.AsFlatCommand()
	expected := "{" + now.Format("02.01.2006 15:04:05") + "} [id: test-id, status: true] test arg1"
	assert.Equal(t, expected, flat)
}
