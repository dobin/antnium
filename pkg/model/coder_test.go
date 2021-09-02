package model

import (
	"testing"

	"github.com/dobin/antnium/pkg/campaign"
)

func TestEncodeDecode(t *testing.T) {
	campaign := campaign.MakeCampaign()
	coder := MakeCoder(&campaign)

	packetId := "1234"
	arguments := make(PacketArgument)
	arguments["remoteurl"] = "http://127.0.0.1:4444/upload/" + packetId
	arguments["source"] = "README.md"
	response := make(PacketResponse)
	packet := NewPacket("fileupload", "0", packetId, arguments, response)

	data, err := coder.EncodeData(packet)
	if err != nil {
		t.Errorf("Error encoding: " + err.Error())
	}

	packet2, err := coder.DecodeData(data)
	if err != nil {
		t.Errorf("Error decoding: " + err.Error())
	}

	if packet.PacketId != packet2.PacketId {
		t.Errorf("Comparison error")
	}
}

func TestEncryptionSuccess(t *testing.T) {
	campaign := campaign.MakeCampaign()
	coder := MakeCoder(&campaign)

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

func TestEncryptionFail(t *testing.T) {
	campaign := campaign.MakeCampaign()
	coder := MakeCoder(&campaign)

	reference := "verysecret"

	encryptedData, err := coder.encryptData([]byte(reference))
	if err != nil {
		t.Errorf("Encrypt error")
	}
	coder.campaign.EncKey = []byte("12345678123456781234567812345678")
	_, err = coder.decryptData(encryptedData)
	if err == nil {
		t.Errorf("Decrypt error, was able to decrypt with wrong key")
	}
}
