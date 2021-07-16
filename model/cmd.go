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
}

var ComputerId = "23"

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
		res = NewCommandExec(packetExec.PacketId, packetExec.Arguments, packetExec.Response)
	}
	if packet.Command == "test" {
		var packetTest PacketTest
		if err := json.Unmarshal([]byte(jsonStr), &packetTest); err != nil {
			log.Fatal(err)
		}
		res = NewCommandTest(packetTest.PacketId, packetTest.Arguments, packetTest.Response)
	}
	if packet.Command == "info" {
		var packetInfo PacketInfo
		if err := json.Unmarshal([]byte(jsonStr), &packetInfo); err != nil {
			log.Fatal(err)
		}
		res = NewCommandInfo(packetInfo.PacketId, packetInfo.Response)
	}
	if packet.Command == "ping" {
		var packetPing PacketPing
		if err := json.Unmarshal([]byte(jsonStr), &packetPing); err != nil {
			log.Fatal(err)
		}
		res = NewCommandPing(packetPing.PacketId, packetPing.Response)
	}
	return res
}

type CommandExec struct {
	PacketId     string
	Arguments    []string
	responseText string
}

func NewCommandExec(packetId string, args []string, response string) *CommandExec {
	c := CommandExec{
		packetId,
		args,
		response,
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

func (c CommandExec) Json() string {
	p := PacketExec{
		Packet{ComputerId, c.PacketId, "exec", c.responseText},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

type CommandInfo struct {
	PacketId     string
	responseText string
}

func NewCommandInfo(packetId string, response string) *CommandInfo {
	c := CommandInfo{packetId, response}
	return &c
}

func (c CommandInfo) Execute() {
	fmt.Printf("Execute: Info")
}

func (c CommandInfo) Response() string {
	return c.responseText
}

func (c CommandInfo) Json() string {
	p := PacketInfo{
		Packet{ComputerId, c.PacketId, "info", c.responseText},
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

type CommandTest struct {
	PacketId     string
	Arguments    []string
	responseText string
}

func NewCommandTest(packetId string, args []string, response string) *CommandTest {
	c := CommandTest{
		packetId,
		args,
		response,
	}
	return &c
}

func (c *CommandTest) Execute() {
	c.responseText = "executed"
}

func (c *CommandTest) Response() string {
	return c.responseText
}

func (c CommandTest) Json() string {
	p := PacketTest{
		Packet{ComputerId, c.PacketId, "test", c.responseText},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

type CommandPing struct {
	PacketId     string
	responseText string
}

func NewCommandPing(packetId string, response string) *CommandPing {
	c := CommandPing{
		packetId,
		response,
	}
	return &c
}

func (c *CommandPing) Execute() {
	c.responseText = "oy!"
}

func (c *CommandPing) Response() string {
	return c.responseText
}

func (c CommandPing) Json() string {
	p := PacketPing{
		Packet{ComputerId, c.PacketId, "ping", c.responseText},
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}
