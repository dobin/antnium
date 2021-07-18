package server

import (
	"time"
)

type HostBase struct {
	ComputerId string
	LastSeen   time.Time
}
