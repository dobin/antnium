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

		res = NewCommandExec(packetExec.Arguments)
	}
	if packet.Command == "test" {
		var packetTest PacketTest
		if err := json.Unmarshal([]byte(jsonStr), &packetTest); err != nil {
			log.Fatal(err)
		}

		res = NewCommandTest(packetTest.Arguments)
	}
	if packet.Command == "info" {
		res = NewCommandInfo()
	}

	return res
}

type CommandExec struct {
	Arguments    []string
	responseText string
}

func NewCommandExec(args []string) *CommandExec {
	c := CommandExec{
		args,
		"",
	}
	return &c
}

func (c CommandExec) Execute() {
	fmt.Printf("Execute: Exec: %v\n", c.Arguments)
}

func (c CommandExec) Response() string {
	return "respo exec"
}

func (c CommandExec) Json() string {
	p := PacketExec{
		Packet{"exec"},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}

type CommandInfo struct {
}

func NewCommandInfo() *CommandInfo {
	c := CommandInfo{}
	return &c
}

func (c CommandInfo) Execute() {
	fmt.Printf("Execute: Info")
}

func (c CommandInfo) Response() string {
	return "respo info"
}

func (c CommandInfo) Json() string {
	p := PacketInfo{
		Packet{"info"},
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

func NewCommandTest(args []string) *CommandTest {
	c := CommandTest{
		args,
		"",
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
		Packet{"test"},
		c.Arguments,
	}
	json, err := json.Marshal(p)
	if err != nil {
	}
	return string(json)
}
