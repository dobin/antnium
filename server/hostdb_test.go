package server

import (
	"testing"
)

func TestHostDb(t *testing.T) {

	hostDb := MakeHostDb()

	hostDb.updateFor("1-1", "1.1.1.1")
	hostDb.updateFor("1-1", "1.1.1.1")
	hostDb.updateFor("1-2", "1.1.1.2")

	hostList := hostDb.getAsList()
	if len(hostList) != 2 {
		t.Errorf("Len wrong")
	}

	// The order here should not matter, but we test it somehow
	// 1
	if hostList[0].ComputerId != "1-1" {
		t.Errorf("Error")
	}
	if hostList[0].LastIp != "1.1.1.1" {
		t.Errorf("Error")
	}
	// 2
	if hostList[1].ComputerId != "1-2" {
		t.Errorf("Error")
	}
	if hostList[1].LastIp != "1.1.1.2" {
		t.Errorf("Error")
	}

	// Checkk
	if hostList[1].LastSeen.After(hostList[0].LastSeen) {
		t.Errorf("Error")
	}
	hostDb.updateFor("1-1", "1.1.1.3")
	hostList = hostDb.getAsList()
	if len(hostList) != 2 {
		t.Errorf("Len wrong")
	}
	if hostList[0].LastIp != "1.1.1.3" {
		t.Errorf("Error: IP is %s", hostList[0].LastIp)
	}
}
