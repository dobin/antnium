package model

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

type Campaign struct {
	ApiKey  string `json:"ApiKey"`
	EncKey  []byte `json:"EncKey"`
	WithZip bool   `json:"WithZip"`
	WithEnc bool   `json:"WithEnc"`

	ServerUrl string `json:"ServerUrl"`

	CommandSendPath         string `json:"CommandSendPath"`
	CommandGetPath          string `json:"CommandGetPath"`
	CommandFileUploadPath   string `json:"CommandFileUploadPath"`
	CommandFileDownloadPath string `json:"CommandFileDownloadPath"`
}

func MakeCampaign() Campaign {
	apiKey := "Secret-Api-Key"
	encKey := "Secret-Enc-Key"

	// Generate the actual AES key based on encKey
	key, err := scrypt.Key([]byte(encKey), []byte("antnium-salt"), 2048, 8, 1, 32)
	if err != nil {
		log.Fatal("Could not create key")
	}

	c := Campaign{
		apiKey,
		key,
		true,
		true,
		"http://localhost:4444",
		"/send",
		"/get/",
		"/upload/",
		"/static/",
	}
	return c
}
