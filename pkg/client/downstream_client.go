package client

import (
	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
)

type DownstreamClient struct {
	packetExecutor executor.PacketExecutor
}

func MakeDownstreamClient() DownstreamClient {
	u := DownstreamClient{
		executor.MakePacketExecutor(),
	}
	return u
}

func (d *DownstreamClient) do(packet model.Packet) (model.Packet, error) {
	packet, err := d.packetExecutor.Execute(packet)
	if err != nil {
		packet.Response["error"] = err.Error()
		return packet, err
	}
	return packet, nil
}
