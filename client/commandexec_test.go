package client

import (
	"testing"

	"github.com/dobin/antnium/model"
)

func TestCmd(t *testing.T) {
	cmdArgument := make(model.CmdArgument, 3)

	cmdArgument["executable"] = "e"
	cmdArgument["param0"] = "a0"
	cmdArgument["param1"] = "a1"

	executable, args, err := model.MakeCmdArgumentFrom(cmdArgument)
	if err != nil {
		t.Errorf("Make error")
	}
	if executable != "e" {
		t.Errorf("executable error")
	}
	if args[0] != "a0" {
		t.Errorf("arg0 error")
	}
	if args[1] != "a1" {
		t.Errorf("arg1 error")
	}
}
