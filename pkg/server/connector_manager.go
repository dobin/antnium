package server

import "github.com/dobin/antnium/pkg/campaign"

type ConnectorManager struct {
	campaign   *campaign.Campaign
	middleware *Middleware

	ConnectorRest *ConnectorRest
	ConnectorWs   *ConnectorWs
}

func MakeConnectorManager(campaign *campaign.Campaign, middleware *Middleware) ConnectorManager {
	connectorRest := MakeConnectorRest(campaign, middleware)
	connectorWs := MakeConnectorWs(campaign, middleware)

	f := ConnectorManager{
		campaign:      campaign,
		middleware:    middleware,
		ConnectorRest: &connectorRest,
		ConnectorWs:   &connectorWs,
	}
	return f
}
