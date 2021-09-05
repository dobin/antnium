package server

import "github.com/dobin/antnium/pkg/campaign"

type FrontendManager struct {
	campaign     *campaign.Campaign
	FrontendRest FrontendRest
	FrontendWs   FrontendWs
}

func MakeFrontendManager(campaign *campaign.Campaign, middleware *Middleware) FrontendManager {
	frontendRest := MakeFrontendRest(campaign, middleware)
	frontendWs := MakeFrontendWs(campaign)

	f := FrontendManager{
		campaign:     campaign,
		FrontendRest: frontendRest,
		FrontendWs:   frontendWs,
	}
	return f
}
