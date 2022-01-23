// +build windows

package executor

import (
	"testing"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/model"
)

func TestCmdValidCommand(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "cmd"
	packetArgument["commandline"] = "hostname"

	stdOut, stdErr, pid, exitCode, err := arch.Exec(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if len(stdOut) == 0 {
		t.Error("No stdout")
	}
	if len(stdErr) != 0 {
		t.Error("Stderr")
	}
	if pid == 0 {
		t.Error("Pid")
	}
	if exitCode != 0 {
		t.Error("ExitCode")
	}
}

func TestPowershellValidCommand(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "powershell"
	packetArgument["commandline"] = "hostname"

	stdOut, stdErr, pid, exitCode, err := arch.Exec(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if len(stdOut) == 0 {
		t.Error("No stdout")
	}
	if len(stdErr) != 0 {
		t.Error("Stderr")
	}
	if pid == 0 {
		t.Error("Pid")
	}
	if exitCode != 0 {
		t.Error("ExitCode")
	}
}

func TestCmdInvalidCommand(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "cmd"
	packetArgument["commandline"] = "invalid"

	stdOut, stdErr, pid, exitCode, err := arch.Exec(packetArgument)
	if err == nil {
		t.Error("No Error")
	}
	if len(stdOut) != 0 {
		t.Error("Stdout")
	}
	if len(stdErr) == 0 {
		t.Error("No Stderr")
	}
	if pid == 0 {
		t.Error("Pid")
	}
	if exitCode == 0 {
		t.Error("ExitCode")
	}
}

func TestPowershellInvalidCommand(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "powershell"
	packetArgument["commandline"] = "invalid"

	stdOut, stdErr, pid, exitCode, err := arch.Exec(packetArgument)
	if err == nil {
		t.Error("No Error")
	}
	if len(stdOut) != 0 {
		t.Error("Stdout")
	}
	if len(stdErr) == 0 {
		t.Error("No Stderr")
	}
	if pid == 0 {
		t.Error("Pid")
	}
	if exitCode == 0 {
		t.Error("ExitCode")
	}
}

func TestRawCopy(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "raw"
	packetArgument["executable"] = "C:\\windows\\system32\\net.exe"
	packetArgument["param0"] = "user"
	packetArgument["param1"] = "dobin"
	packetArgument["copyFirst"] = "C:\\temp\\server.exe"

	stdOut, stdErr, pid, exitCode, err := arch.Exec(packetArgument)
	if err != nil {
		t.Error("Error: " + err.Error())
		return
	}
	if len(stdErr) > 0 {
		t.Error("Stderr: " + string(stdErr))
		return
	}
	if len(stdOut) == 0 {
		t.Error("Stdout")
		return
	}
	if pid == 0 {
		t.Error("Pid")
		return
	}
	if exitCode != 0 {
		t.Error("ExitCode")
		return
	}

}
