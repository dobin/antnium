package model

import (
	"testing"
)

func TestFromJson(t *testing.T) {
	a := `{ "command": "test", "arguments": [ "arg0", "arg1" ] }`
	command := JsonToCommand(a)
	command.Execute()
	if command.Response() != "executed" {
		t.Errorf("Could not execute command test: " + command.Response())
	}
}

func TestToJson(t *testing.T) {
	c := NewCommandTest([]string{"arg0", "arg1"})
	reference := `{"command":"test","arguments":["arg0","arg1"]}`
	if c.Json() != reference {
		t.Errorf("Error jsonify: " + c.Json())
	}
}
