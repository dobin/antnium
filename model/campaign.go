package model

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

type Campaign struct {
	ApiKey string
	EncKey []byte

	WithZip bool
	WithEnc bool

	ServerUrl string

	CommandSendPath   string
	CommandGetPath    string
	CommandUploadPath string
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
		"/sendCommand",
		"/getCommand/",
		"/upload/",
	}
	return c
}
