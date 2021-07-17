package model

import (
	"encoding/json"
	"fmt"
	"log"
)

type Command interface {
	Execute()
	SetComputerId(computerId string)
	GetComputerId() string
	GetResponse() string
}

func JsonToCommand(jsonStr string) Command {
	var res Command

	var packet CommandBase
	if err := json.Unmarshal([]byte(jsonStr), &packet); err != nil {
		log.Fatal(err)
	}
	if packet.Command == "exec" {
		var commandExec CommandExec
		if err := json.Unmarshal([]byte(jsonStr), &commandExec); err != nil {
			log.Fatal(err)
		}
		res = &commandExec
	}
	if packet.Command == "test" {
		var commandTest CommandTest
		if err := json.Unmarshal([]byte(jsonStr), &commandTest); err != nil {
			log.Fatal(err)
		}
		res = &commandTest
	}
	if packet.Command == "info" {
		var commandInfo CommandInfo
		if err := json.Unmarshal([]byte(jsonStr), &commandInfo); err != nil {
			log.Fatal(err)
		}
		res = &commandInfo
	}
	if packet.Command == "ping" {
		var commandPing CommandPing
		if err := json.Unmarshal([]byte(jsonStr), &commandPing); err != nil {
			log.Fatal(err)
		}
		res = &commandPing
	}
	return res
}

type CommandBase struct {
	ComputerId string `json:"computerid"`
	PacketId   string `json:"packetid"`
	Command    string `json:"command"`
	Response   string `json:"response"`
}

func (c *CommandBase) GetResponse() string {
	return c.Response
}

func (c *CommandBase) SetComputerId(computerId string) {
	c.ComputerId = computerId
}

func (c *CommandBase) GetComputerId() string {
	return c.ComputerId
}

type CommandExec struct {
	CommandBase
	Arguments []string `json:"arguments"`
}

func NewCommandExec(computerId string, packetId string, args []string, response string) *CommandExec {
	c := CommandExec{
		CommandBase{
			computerId,
			packetId,
			"exec",
			response,
		},
		args,
	}
	return &c
}

func (c *CommandExec) Execute() {
	fmt.Printf("Execute: Exec: %v\n", c.Arguments)
	c.CommandBase.Response = "exec executed"
}

type CommandInfo struct {
	CommandBase
}

func NewCommandInfo(computerId string, packetId string, response string) *CommandInfo {
	c := CommandInfo{
		CommandBase{
			computerId,
			packetId,
			"info",
			response,
		},
	}
	return &c
}

func (c CommandInfo) Execute() {
	fmt.Printf("Execute: Info")
}

type CommandTest struct {
	CommandBase
	Arguments []string `json:"arguments"`
}

func NewCommandTest(computerId string, packetId string, args []string, response string) *CommandTest {
	c := CommandTest{
		CommandBase{
			computerId,
			packetId,
			"test",
			response,
		},
		args,
	}
	return &c
}

func (c *CommandTest) Execute() {
	c.CommandBase.Response = "executed"
}

type CommandPing struct {
	CommandBase
}

func NewCommandPing(computerId string, packetId string, response string) *CommandPing {
	c := CommandPing{
		CommandBase{
			computerId,
			packetId,
			"ping",
			response,
		},
	}
	return &c
}

func (c *CommandPing) Execute() {
	c.CommandBase.Response = "oy!"
}

func (c *CommandPing) SetComputerId(computerId string) {
	c.ComputerId = computerId
}
