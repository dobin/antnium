package client

import "github.com/dobin/antnium/model"

type CommandExec struct {
}

func MakeCommandExec() CommandExec {
	commandExec := CommandExec{}
	return commandExec
}

func (s *CommandExec) execute(command *model.CommandBase) error {
	if command.Command == "ping" {
		command.Response["response"] = "ping answer"
	} else if command.Command == "test" {
		command.Response["response"] = "test answer"
	} else {
		command.Response["response"] = "generic"
	}
	return nil
}
