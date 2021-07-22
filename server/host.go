package server

import (
	"time"
)

type HostBase struct {
	ComputerId string    `json:"ComputerId"`
	LastSeen   time.Time `json:"LastSeen"`
	LastIp     string    `json:"LastIp"`
}
