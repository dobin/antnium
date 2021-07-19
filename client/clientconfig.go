package client

import (
	"github.com/rs/xid"
)

type ClientConfig struct {
	ComputerId      string
	DestinationHost string

	CommandSendPath string
	CommandGetPath  string
}

func MakeClientConfig() ClientConfig {
	db := ClientConfig{
		xid.New().String(),
		"http://localhost:4444",
		"/sendCommand",
		"/getCommand/",
	}
	return db
}
