package server

import (
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
)

type ServerManager struct {
	srvaddr         string
	Campaign        *campaign.Campaign
	coder           model.Coder
	packetDb        *PacketDb
	clientInfoDb    *ClientInfoDb
	adminWebSocket  *AdminWebSocket
	clientWebSocket *ClientWebSocket
}

func NewServerManager(srvAddr string) ServerManager {
	campaign := campaign.MakeCampaign()
	coder := model.MakeCoder(&campaign)
	packetDb := MakePacketDb()
	clientInfoDb := MakeClientInfoDb()
	adminWebsocket := MakeAdminWebSocket(campaign.AdminApiKey)
	clientWebsocket := MakeClientWebSocket(&campaign)
	d := ServerManager{
		srvAddr,
		&campaign,
		coder,
		&packetDb,
		&clientInfoDb,
		&adminWebsocket,
		&clientWebsocket,
	}

	/*
		// Clients connected via websocket do not send regular ping packets (that's the idea of it)
		// Sadly this makes LastSeen useless - but the user wants to know if the client is still connected.
		// Here we regularly check the clients connected to ClientWebsocket, and update their LastSeen
		// Lifetime: App
		go func() {
			clientInfoDb2 := &clientInfoDb
			for {
				time.Sleep(10 * time.Second)
				c := clientWebsocket.clients
				for computerId, conn := range c {
					if conn == nil {
						continue
					}
					clientInfoDb2.updateFor(computerId, conn.RemoteAddr().String())
				}

				// Todo: When to quit?
			}
		}()
	*/

	return d
}
