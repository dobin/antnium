// +build linux

package executor

import (
	"testing"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/model"
)

func TestBashValidCommand(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "bash"
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

func TestBashInvalidCommand(t *testing.T) {
	packetArgument := make(model.PacketArgument, 2)
	packetArgument["shelltype"] = "bash"
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
