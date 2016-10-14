package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/ttacon/chalk"
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
	Order   int
	ID      string
	Command string
	Status  bool
	When    time.Time
}

func (c Command) String() string {
	b, err := json.Marshal(c)

	if err != nil {
		parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
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

func (c ExecutedCommand) AsFlatCommand() string {
	return "{" + c.When.Format("02.01.2006 15:04:05") + "} [id: " + c.ID + ", status: " + strconv.FormatBool(c.Status) + "] " + c.Command
}

func (c ExecutedCommand) Print() {
	parrot.Print("{", chalk.Yellow, c.When.Format("02.01.2006 15:04:05"), chalk.Reset, "} ")

	if c.Status {
		parrot.Print("[", chalk.Green, c.ID, chalk.Reset, "] ")
	} else {
		parrot.Print("[", chalk.Red, c.ID, chalk.Reset, "] ")
	}
	parrot.Println(c.Command)
}
