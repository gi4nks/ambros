package models

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/ttacon/chalk"

	"github.com/gi4nks/quant"
)

type Entity struct {
	ID           string
	CreatedAt    time.Time
	TerminatedAt time.Time
}

type Command struct {
	Entity

	Name      string
	Arguments []string
	Status    bool
	Output    string
	Error     string
}

type ExecutedCommand struct {
	parrot *quant.Parrot

	Order   int
	ID      string
	Command string
	Status  bool
	When    time.Time
}

func (c *Command) Clone() *Command {
	// Create a new Command object with the same field values as the original
	clone := &Command{
		Entity: Entity{
			ID:           c.ID,
			CreatedAt:    c.CreatedAt,
			TerminatedAt: c.TerminatedAt,
		},
		Name:      c.Name,
		Arguments: make([]string, len(c.Arguments)),
		Status:    c.Status,
		Output:    c.Output,
		Error:     c.Error,
	}

	// Copy the elements of the Arguments slice to the clone's Arguments slice
	copy(clone.Arguments, c.Arguments)

	return clone
}

func (c Command) String() (string, error) {
	b, err := json.Marshal(c)

	if err != nil {
		return "{}", err
	}
	return string(b), nil
}

func (c *Command) AsHistory() string {
	return "Name : " + c.Name + " --> Arguments : " + strings.Join(c.Arguments, " ")
}

func (c *Command) AsExecutedCommand(order int) ExecutedCommand {
	s := c.Name + " " + strings.Join(c.Arguments, " ")
	return ExecutedCommand{Order: order, ID: c.ID, Command: s, Status: c.Status, When: c.CreatedAt}
}

func (c Command) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ID":           c.ID,
		"Name":         c.Name,
		"Arguments":    c.Arguments,
		"Status":       c.Status,
		"Output":       c.Output,
		"Error":        c.Error,
		"CreatedAt":    c.CreatedAt,
		"TerminatedAt": c.TerminatedAt,
	}
}

func (c *Command) FromMap(frommap map[string]interface{}) {
	c.ID = frommap["ID"].(string)
	c.Name = frommap["Name"].(string)
	c.Arguments = frommap["Arguments"].([]string)
	c.Status = frommap["Status"].(bool)
	c.Output = frommap["Output"].(string)
	c.Error = frommap["Error"].(string)
	c.CreatedAt = frommap["CreatedAt"].(time.Time)
	c.TerminatedAt = frommap["TerminatedAt"].(time.Time)
}

func (c Command) AsStoredCommand() string {
	return "[" + c.ID + "] " + c.Name + " " + strings.Join(c.Arguments, " ")
}

func (c ExecutedCommand) AsFlatCommand() string {
	return "{" + c.When.Format("02.01.2006 15:04:05") + "} [id: " + c.ID + ", status: " + strconv.FormatBool(c.Status) + "] " + c.Command
}

func (c ExecutedCommand) Print(parrot *quant.Parrot) {
	parrot.Print("{", chalk.Yellow, c.When.Format("02.01.2006 15:04:05"), chalk.Reset, "} ")

	if c.Status {
		parrot.Print("[", chalk.Green, c.ID, chalk.Reset, "] ")
	} else {
		parrot.Print("[", chalk.Red, c.ID, chalk.Reset, "] ")
	}
	parrot.Println(c.Command)
}
