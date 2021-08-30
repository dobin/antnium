package executor

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestInteractiveShellCmdexe(t *testing.T) {
	interactiveShell := makeInteractiveShell()
	stdout, stderr, err := interactiveShell.open("cmd.exe", []string{"/a"})
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if !strings.Contains(stdout, "Microsoft") {
		t.Errorf("Packet 1 error")
		return
	}

	stdout, stderr, err = interactiveShell.issue("echo test")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stdout, "test") {
		t.Errorf("Packet 1 error")
		return
	}
	fmt.Printf("%s%s", stdout, stderr)

	stdout, stderr, err = interactiveShell.issue("invalid")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stderr, "is not recognized") {
		t.Errorf("Packet 1 error")
		return
	}
}

func TestInteractiveShellPowershell(t *testing.T) {
	interactiveShell := makeInteractiveShell()
	stdout, stderr, err := interactiveShell.open("powershell.exe", []string{"-ExecutionPolicy", "Bypass"})
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

	stdout, stderr, err = interactiveShell.issue("echo test")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stdout, "test") {
		t.Errorf("Packet 1 error: %s", stdout)
		return
	}

	time.Sleep(time.Millisecond * 100)
	stdout, stderr, err = interactiveShell.issue("invalid")
	if err != nil {
		t.Errorf("Packet 1 error")
		return
	}
	if !strings.Contains(stderr, "is not recognized") {
		t.Errorf("Packet 1 error")
		return
	}
}
