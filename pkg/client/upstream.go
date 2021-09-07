package client

import (
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
)

type Upstream interface {
	Connect() error
	Connected() bool
	Start()
	ChanIncoming() chan model.Packet
	ChanOutgoing() chan model.Packet
}

func getProxy(campaign *campaign.Campaign) (string, bool) {
	if campaign.ProxyUrl != "" {
		return campaign.ProxyUrl, true
	} else {
		return "", false
	}
}
