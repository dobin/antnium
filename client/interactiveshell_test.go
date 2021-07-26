package client

import (
	"fmt"
	"strings"
	"testing"
)

func TestInteractiveShell(t *testing.T) {
	interactiveShell := makeInteractiveShell()
	stdout, stderr, err := interactiveShell.open()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Printf("%s", stdout)
	if !strings.Contains(stdout, "Microsoft") {
		t.Errorf("Packet 1 error")
		return
	}

	stdout, stderr = interactiveShell.issue("hostname")
	fmt.Printf("%s%s", stdout, stderr)
	if !strings.Contains(stdout, "unreal") {
		t.Errorf("Packet 1 error")
		return
	}

	stdout, stderr = interactiveShell.issue("whoami")
	if !strings.Contains(stdout, "dobin") {
		t.Errorf("Packet 1 error")
		return
	}
	fmt.Printf("%s%s", stdout, stderr)

	stdout, stderr = interactiveShell.issue("meh")
	fmt.Printf("%s%s", stdout, stderr)
	if !strings.Contains(stderr, "is not recognized") {
		t.Errorf("Packet 1 error")
		return
	}
}
