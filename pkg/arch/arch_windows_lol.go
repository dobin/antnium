// +build windows

package arch

import (
	"context"
	"fmt"
	"os/exec"
)

func execCmdExe(commandStr string) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	commandStr = ResolveWinVar(commandStr)
	executable := `C:\windows\system32\cmd.exe`
	x := fmt.Sprintf(`cmd.exe /S /C "%s"`, commandStr)
	args := []string{x}

	cmd := exec.CommandContext(ctx, executable, args...)

	/* Fix up windows exceptions in process parameter handling */
	cmd.SysProcAttr = getSysProcAttrs()
	cmd.SysProcAttr.CmdLine = args[0]

	return execIt(cmd)
}

func execPowershell(commandStr string) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	commandStr = ResolveWinVar(commandStr)
	executable := `C:\Windows\System32\WindowsPowershell\v1.0\powershell.exe`
	args := []string{"-ExecutionPolicy", "Bypass", "-C", commandStr}

	cmd := exec.CommandContext(ctx, executable, args...)

	/* Fix up windows exceptions in process parameter handling */
	cmd.SysProcAttr = getSysProcAttrs()

	return execIt(cmd)
}
