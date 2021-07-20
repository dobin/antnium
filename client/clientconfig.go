package client

import (
	"github.com/rs/xid"
)

type ClientConfig struct {
	ComputerId string
}

func MakeClientConfig() ClientConfig {
	db := ClientConfig{
		xid.New().String(),
	}
	return db
}
