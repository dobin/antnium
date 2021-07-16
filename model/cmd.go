package model

import (
	"encoding/json"
	"fmt"
	"log"
)

type Command interface {
	Json() string
	Execute()
	Response() string
	SetComputerId(computerId string)
}

func JsonToCommand(jsonStr string) Command {
	var res Command

	var packet Packet
	if err := json.Unmarshal([]byte(jsonStr), &packet); err != nil {
		log.Fatal(err)
	}
	if packet.Command == "exec" {
		var packetExec PacketExec
		if err := json.Unmarshal([]byte(jsonStr), &packetExec); err != nil {
			log.Fatal(err)
		}
		res = NewCommandExec(packetExec.ComputerId, packet.PacketId, packetExec.Arguments, packetExec.Response)
	}
	if packet.Command == "test" {
		var packetTest PacketTest
		if err := json.Unmarshal([]byte(jsonStr), &packetTest); err != nil {
			log.Fatal(err)
		}
		res = NewCommandTest(packetTest.ComputerId, packetTest.PacketId, packetTest.Arguments, packetTest.Response)
	}
	if packet.Command == "info" {
		var packetInfo PacketInfo
		if err := json.Unmarshal([]byte(jsonStr), &packetInfo); err != nil {
			log.Fatal(err)
		}
		res = NewCommandInfo(packetInfo.ComputerId, packetInfo.PacketId, packetInfo.Response)
	}
	if packet.Command == "ping" {
		var packetPing PacketPing
		if err := json.Unmarshal([]byte(jsonStr), &packetPing); err != nil {
			log.Fatal(err)
		}
		/*
			fmt.Println("XXX1: " + jsonStr)
			fmt.Println("XXX2: " + packetPing.PacketId)
			fmt.Println("XXX3: " + packet.PacketId)
			fmt.Printf("YYY: %+v\n", packetPing)*/
		res = NewCommandPing(packetPing.ComputerId, packetPing.PacketId, packetPing.Response)
	}
	return res
}

type CommandBase struct {
	ComputerId   string
	PacketId     string
	responseText string
}

type CommandExec struct {
	CommandBase
	Arguments []string
}

func NewCommandExec(computerId string, packetId string, args []string, response string) *CommandExec {
	c := CommandExec{
		CommandBase{
			computerId,
			packetId,
			response,
		},
		args,
	}
	return &c
}

func (c CommandExec) Execute() {
	fmt.Printf("Execute: Exec: %v\n", c.Arguments)
	c.responseText = "exec executed"
}

func (c CommandExec) Response() string {
	return c.responseText
}

func (c *CommandExec) Json() string {
	p := PacketExec{
		Packet{c.ComputerId, c.PacketId, "exec", c.responseText},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

func (c *CommandExec) SetComputerId(computerId string) {
	c.ComputerId = computerId
}

type CommandInfo struct {
	CommandBase
}

func NewCommandInfo(computerId string, packetId string, response string) *CommandInfo {
	c := CommandInfo{
		CommandBase{
			computerId,
			packetId,
			response,
		},
	}
	return &c
}

func (c CommandInfo) Execute() {
	fmt.Printf("Execute: Info")
}

func (c CommandInfo) Response() string {
	return c.responseText
}

func (c *CommandInfo) Json() string {
	p := PacketInfo{
		Packet{c.ComputerId, c.PacketId, "info", c.responseText},
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

func (c *CommandInfo) SetComputerId(computerId string) {
	c.ComputerId = computerId
}

type CommandTest struct {
	CommandBase
	Arguments []string
}

func NewCommandTest(computerId string, packetId string, args []string, response string) *CommandTest {
	c := CommandTest{
		CommandBase{
			computerId,
			packetId,
			response,
		},
		args,
	}
	return &c
}

func (c *CommandTest) Execute() {
	c.responseText = "executed"
}

func (c *CommandTest) Response() string {
	return c.responseText
}

func (c *CommandTest) Json() string {
	p := PacketTest{
		Packet{c.ComputerId, c.PacketId, "test", c.responseText},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

func (c *CommandTest) SetComputerId(computerId string) {
	c.ComputerId = computerId
}

type CommandPing struct {
	CommandBase
}

func NewCommandPing(computerId string, packetId string, response string) *CommandPing {
	c := CommandPing{
		CommandBase{
			computerId,
			packetId,
			response,
		},
	}
	return &c
}

func (c *CommandPing) Execute() {
	c.responseText = "oy!"
}

func (c *CommandPing) Response() string {
	return c.responseText
}

func (c *CommandPing) Json() string {
	p := PacketPing{
		Packet{c.ComputerId, c.PacketId, "ping", c.responseText},
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

func (c *CommandPing) SetComputerId(computerId string) {
	c.ComputerId = computerId
}
