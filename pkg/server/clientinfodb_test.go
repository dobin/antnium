package server

import (
	"testing"
	"time"
)

func TestClientInfoDb(t *testing.T) {

	clientInfoDb := MakeClientInfoDb()

	clientInfoDb.updateFor("1-1", "1.1.1.1")
	clientInfoDb.updateFor("1-2", "1.1.1.2")

	time.Sleep(time.Millisecond * 10) // Needs some time
	clientInfoDb.updateFor("1-1", "1.1.1.1")

	hostList := clientInfoDb.getAsList()
	if len(hostList) != 2 {
		t.Errorf("Len wrong")
		return
	}

	// The order here should not matter, but we test it somehow
	// 1
	if hostList[0].ComputerId != "1-1" {
		t.Errorf("Error 1")
		return
	}
	if hostList[0].LastIp != "1.1.1.1" {
		t.Errorf("Error 2")
		return
	}
	// 2
	if hostList[1].ComputerId != "1-2" {
		t.Errorf("Error 3")
		return
	}
	if hostList[1].LastIp != "1.1.1.2" {
		t.Errorf("Error 4")
		return
	}

	// Check
	if hostList[1].LastSeen.After(hostList[0].LastSeen) {
		t.Errorf("Error host order: %v", hostList)
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
