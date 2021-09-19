package campaign

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

type Campaign struct {
	ApiKey      string `json:"ApiKey"`
	AdminApiKey string `json:"AdminApiKey"`
	EncKey      []byte `json:"EncKey"`
	WithZip     bool   `json:"WithZip"`
	WithEnc     bool   `json:"WithEnc"`

	ServerUrl string `json:"ServerUrl"` // URL of the server, as viewed from the clients

	PacketSendPath      string `json:"PacketSendPath"`
	PacketGetPath       string `json:"PacketGetPath"`
	FileUploadPath      string `json:"FileUploadPath"`
	FileDownloadPath    string `json:"FileDownloadPath"`
	ClientWebsocketPath string `json:"ClientWebsocketPath"`

	HttpJitter           int    `json:"HttpJitter"` // in percent
	ClientUseWebsocket   bool   `json:"ClientUseWebsocket"`
	ProxyUrl             string `json:"ProxyUrl"`             // Default campaign proxy url, empty for none
	AutoStartDownstreams bool   `json:"AutoStartDownstreams"` // opens local tcp socket when true automagically
	DoClientInfo         bool   `json:"DoClientInfo"`         // If we want to disable ClientInfos
}

func MakeCampaign() Campaign {
	apiKey := "Secret-Api-Key"
	adminApiKey := "Secret-AdminApi-Key"
	encKey := "Secret-Enc-Key"
	serverUrl := "http://localhost:8080"
	proxyUrl := ""

	// Generate the actual AES key based on encKey
	key, err := scrypt.Key([]byte(encKey), []byte("antnium-salt"), 2048, 8, 1, 32)
	if err != nil {
		log.Fatal("Could not create key")
	}

	c := Campaign{
		ApiKey:      apiKey,
		AdminApiKey: adminApiKey,
		EncKey:      key,
		WithZip:     true,
		WithEnc:     true,
		ServerUrl:   serverUrl,

		PacketSendPath:      "/send",
		PacketGetPath:       "/get/",
		FileUploadPath:      "/upload/",
		FileDownloadPath:    "/static/",
		ClientWebsocketPath: "/ws",

		HttpJitter:           20,
		ClientUseWebsocket:   true,
		ProxyUrl:             proxyUrl,
		AutoStartDownstreams: false,
		DoClientInfo:         false,
	}
	return c
}

func (c Campaign) GetProxy() (string, bool) {
	if c.ProxyUrl != "" {
		return c.ProxyUrl, true
	} else {
		return "", false
	}
}
