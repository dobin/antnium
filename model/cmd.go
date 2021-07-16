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
		res = NewCommandExec(packetExec.Arguments, packetExec.Response)
	}
	if packet.Command == "test" {
		var packetTest PacketTest
		if err := json.Unmarshal([]byte(jsonStr), &packetTest); err != nil {
			log.Fatal(err)
		}
		res = NewCommandTest(packetTest.Arguments, packetTest.Response)
	}
	if packet.Command == "info" {
		var packetInfo PacketInfo
		if err := json.Unmarshal([]byte(jsonStr), &packetInfo); err != nil {
			log.Fatal(err)
		}
		res = NewCommandInfo(packetInfo.Response)
	}

	return res
}

type CommandExec struct {
	Arguments    []string
	responseText string
}

func NewCommandExec(args []string, response string) *CommandExec {
	c := CommandExec{
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
		Packet{"exec", c.responseText},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

type CommandInfo struct {
	responseText string
}

func NewCommandInfo(response string) *CommandInfo {
	c := CommandInfo{response}
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
		Packet{"info", c.responseText},
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

type CommandTest struct {
	Arguments    []string
	responseText string
}

func NewCommandTest(args []string, response string) *CommandTest {
	c := CommandTest{
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
	p := PacketExec{
		Packet{"test", c.responseText},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}
