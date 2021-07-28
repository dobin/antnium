package client

import (
	"time"
)

type ClientState struct {
	lastPacketSent time.Time
}

func MakeClientState() ClientState {
	d := ClientState{
		time.Now(),
	}
	return d
}

func (c *ClientState) getSleepDuration(gotPacket bool) time.Duration {
	if gotPacket {
		c.lastPacketSent = time.Now()
		return time.Second * 1
	}

	timeNow := time.Now()
	timeDiff := timeNow.Sub(c.lastPacketSent)
	duration := timeDiff.Seconds()

	if duration < 10 {
		return time.Second * 1
	} else if duration < 60 {
		return time.Second * 3
	} else if duration < 600 {
		return time.Second * 10
	} else {
		return time.Second * 60
	}
}
