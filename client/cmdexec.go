package client

import (
	"fmt"
	"os/exec"

	"github.com/dobin/antnium/model"
	log "github.com/sirupsen/logrus"
)

type CommandExec struct {
}

func MakeCommandExec() CommandExec {
	commandExec := CommandExec{}
	return commandExec
}

func (s *CommandExec) execute(command *model.CommandBase) error {
	log.WithFields(log.Fields{
		"command": command,
	}).Info("Execute")

	if command.Command == "ping" {
		command.Response = s.actionPing(command.Arguments)
	} else if command.Command == "test" {
		command.Response = s.actionTest(command.Arguments)
	} else if command.Command == "exec" {
		command.Response = s.actionExec(command.Arguments)
	} else if command.Command == "fileupload" {
		command.Response = s.actionFiledownload(command.Arguments)
	} else if command.Command == "filedownload" {
		command.Response = s.actionFileupload(command.Arguments)
	} else {
		command.Response["response"] = "generic"
	}
	return nil
}

func (s *CommandExec) actionPing(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)
	ret["response"] = "ping answer"
	return ret
}

func (s *CommandExec) actionTest(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)
	ret["response"] = "test answer"
	return ret
}
func (s *CommandExec) actionExec(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)
	args := make([]string, 3)

	// Check Input
	executable, ok := cmdArgument["executable"]
	if !ok {
		fmt.Println("No 1")
		ret["error"] = "No executable given"
		return ret
	}

	// Transform Input
	arg1, ok := cmdArgument["arg1"]
	if ok {
		args = append(args, arg1)
	}
	arg2, ok := cmdArgument["arg2"]
	if ok {
		args = append(args, arg2)
	}

	// Execute and return
	cmd := exec.Command(executable, args...)
	stdout, err := cmd.Output()
	ret["stdout"] = string(stdout)
	if err != nil {
		ret["err"] = err.Error()
	}
	return ret
}

func (s *CommandExec) actionFiledownload(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)
	ret["response"] = "download answer"
	return ret
}

func (s *CommandExec) actionFileupload(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)
	ret["response"] = "upload answer"
	return ret
}
