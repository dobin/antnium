package client

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/aus/proxyplease"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type Upstream interface {
	Connect() error
	Connected() bool
	Start()
	ChanIncoming() chan model.Packet
	SendPacket(packet model.Packet) error
}

func NewDialContext(campaign *campaign.Campaign) (proxyplease.DialContext, error) {
	var dialContext proxyplease.DialContext

	if campaign.DisableProxy {
		log.Info("Disabled proxy, use direct")
		dialContext = (&net.Dialer{
			KeepAlive: 5 * time.Second,
		}).DialContext

		return dialContext, nil
	}

	// Automatic proxy configuration
	proxyUrl, ok := campaign.GetProxy()
	if ok {
		// Manual proxy configuration
		parsedUrl, err := url.Parse(proxyUrl)
		if err != nil {
			//return fmt.Errorf("could not parse ProxyUrl %s: %s", proxyUrl, err.Error())
			return nil, fmt.Errorf("could not parse ProxyUrl %s: %s", proxyUrl, err.Error())
		}
		dialContext = proxyplease.NewDialContext(proxyplease.Proxy{URL: parsedUrl})
	} else {
		dialContext = proxyplease.NewDialContext(proxyplease.Proxy{})
	}
	return dialContext, nil
}
