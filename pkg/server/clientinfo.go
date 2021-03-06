package server

import (
	"time"
)

type ClientInfo struct {
	// From every packet
	ClientId  string    `json:"ClientId"`
	FirstSeen time.Time `json:"FirstSeen"`
	LastSeen  time.Time `json:"LastSeen"`
	LastIp    string    `json:"LastIp"`

	// From ClientInfo
	Hostname      string   `json:"Hostname"`
	LocalIps      []string `json:"LocalIps"`
	Arch          string   `json:"Arch"`
	Processes     []string `json:"Processes"`
	IsAdmin       string   `json:"isAdmin"`
	IsElevated    string   `json:"isElevated"`
	ConnectorType string   `json:"ConnectorType"`
	WorkingDir    string   `json:"WorkingDir"`
	Env           []string `json:"Env"`
}
