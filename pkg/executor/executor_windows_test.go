// +build windows

package executor

import (
	"os"
	"strings"
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

func TestCopyFirst(t *testing.T) {
	copyFirst := "C:\\temp\\server.exe"
	os.Remove(copyFirst)

	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "raw"
	packetArgument["executable"] = "C:\\windows\\system32\\net.exe"
	packetArgument["param0"] = "user"
	packetArgument["param1"] = "dobin"
	packetArgument["copyFirst"] = copyFirst

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
	if _, err := os.Stat(copyFirst); err != nil {
		t.Error("Did not copy")
		return
	} else {
		os.Remove(copyFirst)
	}
}

func TestHollow(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "raw"
	packetArgument["executable"] = "C:\\windows\\system32\\net.exe"
	packetArgument["param0"] = "user"
	packetArgument["param1"] = "dobin"
	packetArgument["hollow"] = "c:\\windows\\system32\\hostname.exe"

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
	out := string(stdOut)
	if !strings.Contains(out, "User name") {
		t.Errorf("Output: %s", out)
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
