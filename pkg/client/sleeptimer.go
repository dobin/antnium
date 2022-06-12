package client

import (
	"time"
)

type SleepTimer struct {
	lastTick time.Time
}

func MakeSleepTimer() SleepTimer {
	d := SleepTimer{
		time.Now(),
	}
	return d
}

func (c *SleepTimer) getSleepDuration() time.Duration {
	timeNow := time.Now()
	timeDiff := timeNow.Sub(c.lastTick)
	diff := timeDiff.Seconds()

	if diff < 10 {
		return time.Second * 1
	} else if diff < 60 { // 1min
		return time.Second * 3
	} else if diff < 600 { // 10min
		return time.Minute * 10
	} else {
		return time.Hour * 1
	}
}

func (c *SleepTimer) tick() {
	c.lastTick = time.Now()
}
