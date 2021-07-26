package server

import (
	"time"
)

type ClientInfo struct {
	ComputerId string    `json:"ComputerId"`
	LastSeen   time.Time `json:"LastSeen"`
	LastIp     string    `json:"LastIp"`
}
