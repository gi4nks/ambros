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
	Status    bool
	Output    string
}

type ExecutedCommand struct {
	Order   int
	Id      uint
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
	return "Name : " + c.Name + " --> Arguments : " + c.Arguments
}

func (c *Command) AsExecutedCommand(order int) ExecutedCommand {
	s := c.Name + " " + c.Arguments
	return ExecutedCommand{Order: order, Id: c.ID, Command: s, Status: c.Status, When: c.CreatedAt}
}

func (c ExecutedCommand) AsFlatCommand() string {
	return "{" + c.When.Format("02.01.2006 15:04:05") + "} [id: " + strconv.Itoa(int(c.Id)) + ", status: " + strconv.FormatBool(c.Status) + "] " + c.Command
}
