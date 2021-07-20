package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
		command.Response = s.actionFileupload(command.Arguments)
	} else if command.Command == "filedownload" {
		command.Response = s.actionFiledownload(command.Arguments)
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

	// Check and transform input
	executable, ok := cmdArgument["executable"]
	if !ok {
		ret["error"] = "No executable given"
		return ret
	}
	arg1, ok := cmdArgument["arg1"]
	if ok {
		args = append(args, arg1)
	}
	arg2, ok := cmdArgument["arg2"]
	if ok {
		args = append(args, arg2)
	}

	// Execute and return result
	cmd := exec.Command(executable, args...)
	stdout, err := cmd.Output()
	if err != nil {

		ret["error"] = err.Error()
	} else {
		ret["stdout"] = string(stdout)
	}
	return ret
}

func (s *CommandExec) actionFiledownload(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)

	// Check and transform input
	remoteurl, ok := cmdArgument["remoteurl"]
	if !ok {
		ret["error"] = "No remoteurl given"
		return ret
	}
	destination, ok := cmdArgument["destination"]
	if !ok {
		ret["error"] = "No destination given"
		return ret
	}

	// Download and write file
	resp, err := http.Get(remoteurl)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}
	defer resp.Body.Close()
	out, err := os.Create(destination)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}
	defer out.Close()
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}

	ret["response"] = fmt.Sprintf("Written %d bytes", written)
	return ret
}

func (s *CommandExec) actionFileupload(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)

	// Check and transform input
	remoteurl, ok := cmdArgument["remoteurl"]
	if !ok {
		ret["error"] = "No remoteurl given"
		return ret
	}
	source, ok := cmdArgument["source"]
	if !ok {
		ret["error"] = "No source given"
		return ret
	}

	client := &http.Client{}
	data, err := os.Open(source)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}
	req, err := http.NewRequest("POST", remoteurl, data)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}
	resp, err := client.Do(req)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}

	ret["response"] = fmt.Sprintf("Status: %s", resp.Status)
	return ret
}
