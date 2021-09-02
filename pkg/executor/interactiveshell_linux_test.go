// +build linux

package executor

import (
	"strings"
	"testing"
	"time"
)

func TestInteractiveShellBash(t *testing.T) {
	interactiveShell := MakeInteractiveShell()
	stdout, stderr, err := interactiveShell.Open("/bin/bash", []string{})
	if err != nil {
		t.Errorf(err.Error())
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
	if !strings.Contains(stderr, "not found") {
		t.Errorf("Packet 1 error")
		return
	}
}
