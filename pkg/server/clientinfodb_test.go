package server

import (
	"testing"
)

func TestClientInfoDb(t *testing.T) {

	clientInfoDb := MakeClientInfoDb()

	clientInfoDb.updateFor("1-1", "1.1.1.1")
	clientInfoDb.updateFor("1-1", "1.1.1.1")
	clientInfoDb.updateFor("1-2", "1.1.1.2")

	hostList := clientInfoDb.getAsList()
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

	// Check
	if hostList[1].LastSeen.After(hostList[0].LastSeen) {
		t.Errorf("Error")
	}
	clientInfoDb.updateFor("1-1", "1.1.1.3")
	hostList = clientInfoDb.getAsList()
	if len(hostList) != 2 {
		t.Errorf("Len wrong")
	}
	if hostList[0].LastIp != "1.1.1.3" {
		t.Errorf("Error: IP is %s", hostList[0].LastIp)
	}
}
