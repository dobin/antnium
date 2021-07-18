package model

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestFromJson(t *testing.T) {
	a := `{ "command": "test", "arguments": [ "arg0", "arg1" ] }`
	command, err := JsonToCommand(a)
	if err != nil {
		t.Errorf("Could not parse command test: %s", err)
	}
	command.Execute()
	if command.GetResponse() != "executed" {
		t.Errorf("Could not execute command test: %s", command.GetResponse())
	}
}

func TestToJson(t *testing.T) {
	c := NewCommandTest("23", "42", []string{"arg0", "arg1"}, "")
	reference := `{"computerid":"23","packetid":"42","command":"test","response":"","arguments":["arg0","arg1"]}`

	u, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	s := string(u)
	fmt.Println(s)
	if s != reference {
		t.Errorf("Error jsonify: " + s)
	}
}
