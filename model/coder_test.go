package model

import (
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	campaign := MakeCampaign()
	coder := MakeCoder(campaign)

	packetId := "1234"
	arguments := make(CmdArgument)
	arguments["remoteurl"] = "http://127.0.0.1:4444/upload/" + packetId
	arguments["source"] = "README.md"
	response := make(CmdResponse)
	command := NewCommand("fileupload", "0", packetId, arguments, response)

	data, err := coder.EncodeData(command)
	if err != nil {
		t.Errorf("Error encoding: " + err.Error())
	}

	command2, err := coder.DecodeData(data)
	if err != nil {
		t.Errorf("Error decoding: " + err.Error())
	}

	if command.PacketId != command2.PacketId {
		t.Errorf("Comparison error")
	}
}

func TestEncryption(t *testing.T) {
	campaign := MakeCampaign()
	coder := MakeCoder(campaign)

	reference := "verysecret"

	encryptedData, err := coder.encryptData([]byte(reference))
	if err != nil {
		t.Errorf("Encrypt error")
	}
	decryptedData, err := coder.decryptData(encryptedData)
	if err != nil {
		t.Errorf("Decrypt error")
	}

	if string(decryptedData) != reference {
		t.Errorf("Comparison error")
	}
}
