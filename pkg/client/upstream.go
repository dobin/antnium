package client

import (
	"fmt"
	"net/url"

	"github.com/aus/proxyplease"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
)

type Upstream interface {
	Connect() error
	Connected() bool
	Start()
	ChanIncoming() chan model.Packet
	SendPacket(packet model.Packet) error
}

//func (u *UpstreamWs) NewDialContext() (*http.Client, error) {
func NewDialContext(campaign *campaign.Campaign) (proxyplease.DialContext, error) {
	var dialContext proxyplease.DialContext

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
