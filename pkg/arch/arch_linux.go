// +build linux freebsd netbsd openbsd

package arch

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/dobin/antnium/pkg/model"
	//log "github.com/sirupsen/logrus"
)

func Permissions() (bool, bool, error) {
	isElevated := false
	isAdmin := false
	/* GO1.16
	currentUser, err := user.Current()
	if err != nil {
		return isElevated, isAdmin, err
	}
	if currentUser.Username == "root" {
		isAdmin = true
	}
	*/

	return isElevated, isAdmin, nil
}

func Exec(packetArgument model.PacketArgument) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	stdOut = make([]byte, 0)
	stdErr = make([]byte, 0)
	pid = 0
	exitCode = 0
	err = nil

	processTimeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	shellType, ok := packetArgument["shelltype"]
	if !ok {
		return stdOut, stdErr, pid, exitCode, fmt.Errorf("no argument 'shelltype' given")
	}

	var cmd *exec.Cmd
	switch shellType {
	case "bash":
		commandLine, ok := packetArgument["commandline"]
		if !ok {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("no argument 'commandline' given")
		}
		cmd = exec.CommandContext(ctx, "/bin/bash", "-c", commandLine)

	case "raw":
		executable, args, err := model.MakePacketArgumentFrom(packetArgument)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("invalid packet arguments given")
		}
		cmd = exec.CommandContext(ctx, executable, args...)

	default:
		return stdOut, stdErr, pid, exitCode, fmt.Errorf("shelltype %s unkown", shellType)
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

func ExecOutputDecode(data []byte) string {
	return string(data)
}
