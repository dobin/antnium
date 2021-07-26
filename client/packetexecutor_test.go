package client

import (
	"testing"

	"github.com/dobin/antnium/model"
)

func TestPacket(t *testing.T) {
	packetArgument := make(model.PacketArgument, 3)

	packetArgument["executable"] = "e"
	packetArgument["param0"] = "a0"
	packetArgument["param1"] = "a1"

	executable, args, err := model.MakePacketArgumentFrom(packetArgument)
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
