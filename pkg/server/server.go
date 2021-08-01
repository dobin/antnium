package server

import (
	"math/rand"
	"time"

	"github.com/dobin/antnium/pkg/model"
)

type Server struct {
	srvaddr        string
	campgain       model.Campaign
	coder          model.Coder
	packetDb       PacketDb
	clientInfoDb   ClientInfoDb
	adminWebSocket AdminWebSocket
}

func NewServer(srvAddr string) Server {
	campaign := model.MakeCampaign()
	coder := model.MakeCoder(&campaign)
	w := Server{
		srvAddr,
		campaign,
		coder,
		MakePacketDb(), MakeClientInfoDb(), MakeAdminWebSocket(campaign.AdminApiKey)}

	// Init random for packet id generation
	// Doesnt need to be secure
	rand.Seed(time.Now().Unix())
	return w
}
