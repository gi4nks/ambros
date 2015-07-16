package main

import (
	"encoding/json"
	"strconv"
	"time"
)

type Entity struct {
	ID           uint `gorm:"primary_key"`
	CreatedAt    time.Time
	TerminatedAt time.Time
}

type Command struct {
	Entity
	Name      string
	Arguments string
	Status    string
	Output    string
}

type ExecutedCommand struct {
	Order   int
	Command string
	When    time.Time
}

func (c Command) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		tracer.Warning(err.Error())
		return "{}"
	}
	return string(b)
}

func (c *Command) AsHistory() string {
	return "Name : " + c.Name + " --> Arguments : " + c.Arguments
}

func (c *Command) AsExecutedCommand(order int) ExecutedCommand {
	return ExecutedCommand{Order: order, Command: c.Name + " " + c.Arguments, When: c.CreatedAt}
}

func (c ExecutedCommand) String() string {
	return "(" + strconv.Itoa(c.Order) + ") " + c.Command + " [" + c.When.String() + "]"
}
