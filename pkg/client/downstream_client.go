package client

import (
	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
)

type DownstreamClient struct {
	executor executor.Executor
}

func MakeDownstreamClient() DownstreamClient {
	u := DownstreamClient{
		executor.MakeExecutor(),
	}
	return u
}

func (d *DownstreamClient) Do(packet model.Packet) (model.Packet, error) {
	packet, err := d.executor.Execute(packet)
	if err != nil {
		return packet, err
	}
	return packet, nil
}
