// +build windows

package executor

import (
	"strings"
	"testing"
	"time"
)

func TestInteractiveShellCmdexe(t *testing.T) {
	interactiveShell := MakeInteractiveShell()
	stdout, stderr, err := interactiveShell.Open("cmd.exe", []string{"/a"})
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if !strings.Contains(stdout, "Microsoft") {
		t.Errorf("Packet 1 error")
		return
	}

	stdout, stderr, err = interactiveShell.Issue("echo test")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stdout, "test") {
		t.Errorf("Packet 1 error")
		return
	}

	stdout, stderr, err = interactiveShell.Issue("invalid")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stderr, "is not recognized") {
		t.Errorf("Packet 1 error")
		return
	}
}

func TestInteractiveShellCmdexeExit(t *testing.T) {
	interactiveShell := MakeInteractiveShell()
	stdout, _, err := interactiveShell.Open("cmd.exe", []string{"/a"})
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if !strings.Contains(stdout, "Microsoft") {
		t.Errorf("Packet 1 error")
		return
	}

	// exit process
	stdout, _, err = interactiveShell.Issue("exit")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}

	// execute dir with closed process, should generate "shell down" error
	stdout, _, err = interactiveShell.Issue("dir")
	if err == nil {
		t.Errorf("Packet 1 no error")
		return
	}
	if !strings.Contains(err.Error(), "Shell down") {
		t.Errorf("Packet 1 error")
		return
	}

	// execute dir with closed process, should generate "shell not open" error
	stdout, _, err = interactiveShell.Issue("dir")
	if err == nil {
		t.Errorf("Packet 1 no error")
		return
	}
	if !strings.Contains(err.Error(), "Shell not open") {
		t.Errorf("Packet 1 error")
		return
	}

	// try opening it again
	stdout, _, err = interactiveShell.Open("cmd.exe", []string{"/a"})
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if !strings.Contains(stdout, "Microsoft") {
		t.Errorf("Packet 1 error")
		return
	}
}

func TestInteractiveShellPowershell(t *testing.T) {
	interactiveShell := MakeInteractiveShell()
	stdout, stderr, err := interactiveShell.Open("powershell.exe", []string{"-ExecutionPolicy", "Bypass"})
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if !strings.Contains(stdout, "Microsoft") {
		t.Errorf("Packet 1 error")
		return
	}

	// Unittest is not stable without this for some reason
	// We are sometimes too fast, and not capturing the result of "echo test", but
	// just the command prompt of it. Even though we give it enough time...
	time.Sleep(time.Millisecond * 100)

	stdout, stderr, err = interactiveShell.Issue("echo test")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stdout, "test") {
		t.Errorf("Packet 1 error: %s", stdout)
		return
	}

	time.Sleep(time.Millisecond * 100)
	stdout, stderr, err = interactiveShell.Issue("invalid")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stderr, "is not recognized") {
		t.Errorf("Packet 1 error")
		return
	}
}
