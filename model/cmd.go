package model

import (
	"encoding/json"
	"fmt"
	"log"
)

type Command interface {
	Execute() error

	// FIXME?
	SetComputerId(computerId string)
	GetComputerId() string
	GetResponse() string
	SetResponse(response string)
	GetPacketId() string
}

func JsonToCommand(jsonStr string) (Command, error) {
	var packet CommandBase
	if err := json.Unmarshal([]byte(jsonStr), &packet); err != nil {
		log.Fatal(err)
	}
	if packet.Command == "exec" {
		var commandExec CommandExec
		if err := json.Unmarshal([]byte(jsonStr), &commandExec); err != nil {
			log.Fatal(err)
		}
		return &commandExec, nil
	}
	if packet.Command == "test" {
		var commandTest CommandTest
		if err := json.Unmarshal([]byte(jsonStr), &commandTest); err != nil {
			log.Fatal(err)
		}
		return &commandTest, nil
	}
	if packet.Command == "info" {
		var commandInfo CommandInfo
		if err := json.Unmarshal([]byte(jsonStr), &commandInfo); err != nil {
			log.Fatal(err)
		}
		return &commandInfo, nil
	}
	if packet.Command == "ping" {
		var commandPing CommandPing
		if err := json.Unmarshal([]byte(jsonStr), &commandPing); err != nil {
			log.Fatal(err)
		}
		return &commandPing, nil
	}

	return nil, fmt.Errorf("Could not parse json %s", jsonStr)
}

type CommandBase struct {
	ComputerId string `json:"computerid"`
	PacketId   string `json:"packetid"`
	Command    string `json:"command"`
	Response   string `json:"response"`
}

func (c *CommandBase) GetPacketId() string {
	return c.PacketId
}

func (c *CommandBase) GetResponse() string {
	return c.Response
}

func (c *CommandBase) SetResponse(response string) {
	c.Response = response
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

func (c *CommandExec) Execute() error {
	fmt.Printf("Execute: Exec: %v\n", c.Arguments)
	c.CommandBase.Response = "exec executed"
	return nil
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

func (c CommandInfo) Execute() error {
	fmt.Printf("Execute: Info")
	return nil
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

func (c *CommandTest) Execute() error {
	c.CommandBase.Response = "executed"
	return nil
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

func (c *CommandPing) Execute() error {
	c.CommandBase.Response = "oy!"
	return nil
}

func (c *CommandPing) SetComputerId(computerId string) {
	c.ComputerId = computerId
}
