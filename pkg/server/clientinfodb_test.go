package server

import (
	"testing"
	"time"
)

func TestClientInfoDb(t *testing.T) {

	clientInfoDb := MakeClientInfoDb()

	clientInfoDb.updateFor("1-1", "1.1.1.1", "")
	clientInfoDb.updateFor("1-2", "1.1.1.2", "")

	time.Sleep(time.Millisecond * 10) // Needs some time
	clientInfoDb.updateFor("1-1", "1.1.1.1", "")

	hostList := clientInfoDb.AllAsList()
	if len(hostList) != 2 {
		t.Errorf("Len wrong")
		return
	}

	// Get order as map is randomized...
	var a int
	var b int
	if hostList[0].ClientId == "1-1" {
		a = 0
		b = 1
	} else if hostList[0].ClientId == "1-2" {
		a = 1
		b = 0
	} else {
		t.Errorf("Hmm")
		return
	}

	// The order here should not matter, but we test it somehow
	// 1
	if hostList[a].ClientId != "1-1" {
		t.Errorf("Error 1")
		return
	}
	if hostList[a].LastIp != "1.1.1.1" {
		t.Errorf("Error 2")
		return
	}
	// 2
	if hostList[b].ClientId != "1-2" {
		t.Errorf("Error 3")
		return
	}
	if hostList[b].LastIp != "1.1.1.2" {
		t.Errorf("Error 4")
		return
	}

	// Check order
	if hostList[b].LastSeen.After(hostList[a].LastSeen) {
		t.Errorf("Error host order: %v", hostList)
	}

	// Check update
	time.Sleep(time.Millisecond * 10) // Needs some time
	clientInfoDb.updateFor("1-1", "1.1.1.3", "")
	hostList = clientInfoDb.AllAsList()
	if len(hostList) != 2 {
		t.Errorf("Len wrong")
	}
	if !((hostList[0].LastIp == "1.1.1.3" && hostList[1].LastIp == "1.1.1.2") || (hostList[1].LastIp == "1.1.1.3" && hostList[0].LastIp == "1.1.1.2")) {
		t.Errorf("Wrong: %v", hostList)
	}
}
