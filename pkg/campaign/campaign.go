package campaign

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

type Campaign struct {
	ApiKey  string `json:"ApiKey"`
	EncKey  []byte `json:"EncKey"`
	WithZip bool   `json:"WithZip"`
	WithEnc bool   `json:"WithEnc"`

	ServerUrl string `json:"ServerUrl"` // URL of the server, as viewed from the clients

	PacketSendPath      string `json:"PacketSendPath"`
	PacketGetPath       string `json:"PacketGetPath"`
	FileUploadPath      string `json:"FileUploadPath"`
	FileDownloadPath    string `json:"FileDownloadPath"`
	SecureDownloadPath  string `json:"SecureDownloadPath"`
	ClientWebsocketPath string `json:"ClientWebsocketPath"`
	AuthHeader          string `json:"AuthHeader"`
	UserAgent           string `json:"UserAgent"`

	HttpJitter           int    `json:"HttpJitter"` // in percent
	ClientUseWebsocket   bool   `json:"ClientUseWebsocket"`
	ProxyUrl             string `json:"ProxyUrl"`             // Default campaign proxy url, empty for none
	DisableProxy         bool   `json:"DisableProxy"`         // If we want to disable any proxy (mostly for testing)
	AutoStartDownstreams bool   `json:"AutoStartDownstreams"` // opens local tcp socket when true automagically
	DoClientInfo         bool   `json:"DoClientInfo"`         // If we want to disable ClientInfos
}

func MakeCampaign() Campaign {
	apiKey := "Secret-Api-Key"
	encKey := "Secret-Enc-Key"
	serverUrl := "http://0.0.0.0:8080"
	proxyUrl := ""

	key := GetKey(encKey)

	c := Campaign{
		ApiKey:    apiKey,
		EncKey:    key,
		WithZip:   true,
		WithEnc:   true,
		ServerUrl: serverUrl,

		PacketSendPath:      "/send",
		PacketGetPath:       "/get/",
		FileUploadPath:      "/upload/",
		FileDownloadPath:    "/static/",
		SecureDownloadPath:  "/secure/",
		ClientWebsocketPath: "/ws",
		AuthHeader:          "X-Session-Token",
		UserAgent:           "Go-http-client/1.1",

		HttpJitter:           20,
		ClientUseWebsocket:   true,
		ProxyUrl:             proxyUrl,
		DisableProxy:         false, //
		AutoStartDownstreams: false, // broken, do not use
		DoClientInfo:         true,
	}
	return c
}

func GetKey(encKey string) []byte {
	key, err := scrypt.Key([]byte(encKey), []byte("antnium-salt"), 32768, 8, 1, 32)
	if err != nil {
		log.Fatal("Could not create key")
	}
	return key
}

func (c Campaign) GetProxy() (string, bool) {
	if c.ProxyUrl != "" {
		return c.ProxyUrl, true
	} else {
		return "", false
	}
}
