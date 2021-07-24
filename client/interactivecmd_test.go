package client

import (
	"fmt"
	"strings"
	"testing"
)

func TestInteractiveCmd(t *testing.T) {
	interactiveCmd := makeInteractiveCmd()
	stdout, stderr, err := interactiveCmd.open()
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Printf("%s", stdout)
	if !strings.Contains(stdout, "Microsoft") {
		t.Errorf("Cmd 1 error")
	}

	stdout, stderr = interactiveCmd.issue("hostname")
	fmt.Printf("%s%s", stdout, stderr)
	if !strings.Contains(stdout, "unreal") {
		t.Errorf("Cmd 1 error")
	}

	stdout, stderr = interactiveCmd.issue("whoami")
	if !strings.Contains(stdout, "dobin") {
		t.Errorf("Cmd 1 error")
	}
	fmt.Printf("%s%s", stdout, stderr)

	stdout, stderr = interactiveCmd.issue("meh")
	fmt.Printf("%s%s", stdout, stderr)
	if !strings.Contains(stderr, "is not recognized") {
		t.Errorf("Cmd 1 error")
	}
}
