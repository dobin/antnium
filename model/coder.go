package model

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

type Coder struct {
	campaign Campaign
}

func MakeCoder(campaign Campaign) Coder {
	w := Coder{campaign}
	return w
}

func (s *Coder) EncodeData(command CommandBase) ([]byte, error) {
	// Go to JSON
	data, err := json.Marshal(command)
	if err != nil {
		log.Error("Could not JSON marshal")
		return nil, err
	}

	// JSON to ZIP
	if s.campaign.WithZip {
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(data)
		w.Close()
		data = b.Bytes()
	}

	// encrypt ZIP
	if s.campaign.WithEnc {
		data, err = s.encryptData(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (s *Coder) DecodeData(data []byte) (CommandBase, error) {
	var err error

	// Cypertext to ZIP
	if s.campaign.WithEnc {
		data, err = s.decryptData(data)
		if err != nil {
			return CommandBase{}, err
		}
	}

	// ZIP to JSON
	if s.campaign.WithZip {
		var out bytes.Buffer
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return CommandBase{}, fmt.Errorf("ZLIB: %v", err)
		}
		_, err = io.Copy(&out, r)
		if err != nil {
			return CommandBase{}, fmt.Errorf("ZLIB copy: %v", err)
		}
		data = out.Bytes()
		r.Close()
	}

	// JSON to GO
	var command CommandBase
	err = json.Unmarshal(data, &command)
	if err != nil {
		log.WithFields(log.Fields{
			"body":  data,
			"error": err,
		}).Info("Error sending command")
		return CommandBase{}, fmt.Errorf("JSON Unmarshall: %s: %v", string(data), err)
	}

	return command, nil
}

func (s *Coder) encryptData(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.campaign.EncKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func (s *Coder) decryptData(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.campaign.EncKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("Length of data is %d, which is smaller than nonceSize %d", len(ciphertext), nonceSize)
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
