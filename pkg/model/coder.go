package model

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/dobin/antnium/pkg/campaign"
)

type Coder struct {
	campaign *campaign.Campaign
}

func MakeCoder(campaign *campaign.Campaign) Coder {
	w := Coder{campaign}
	return w
}

/****/

func (k *Coder) EncodeData(packet Packet) ([]byte, error) {
	// Go to JSON
	data, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}

	// JSON to ZIP
	if k.campaign.WithZip {
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(data)
		w.Close()
		data = b.Bytes()
	}

	// encrypt ZIP
	if k.campaign.WithEnc {
		data, err = k.EncryptData(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (k *Coder) DecodeData(data []byte) (Packet, error) {
	var err error

	// Cypertext to ZIP
	if k.campaign.WithEnc {
		data, err = k.DecryptData(data)
		if err != nil {
			return Packet{}, err
		}
	}

	// ZIP to JSON
	if k.campaign.WithZip {
		var out bytes.Buffer
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return Packet{}, fmt.Errorf("ZLIB: %v", err)
		}
		_, err = io.Copy(&out, r)
		if err != nil {
			return Packet{}, fmt.Errorf("ZLIB copy: %v", err)
		}
		data = out.Bytes()
		r.Close()
	}

	// JSON to GO
	var packet Packet
	err = json.Unmarshal(data, &packet)
	if err != nil {
		return Packet{}, fmt.Errorf("JSON Unmarshall: %s: %v", string(data), err)
	}

	return packet, nil
}

/****/

func (k *Coder) ZipData(data []byte) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()
	data = b.Bytes()

	return data
}

func (k *Coder) UnzipData(data []byte) ([]byte, error) {
	var out bytes.Buffer
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ZLIB: %v", err)
	}
	_, err = io.Copy(&out, r)
	if err != nil {
		return nil, fmt.Errorf("ZLIB copy: %v", err)
	}
	data = out.Bytes()
	r.Close()

	return data, nil
}

func (k *Coder) EncryptData(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.campaign.EncKey)
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

func (k *Coder) DecryptData(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.campaign.EncKey)
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

func (k *Coder) EncryptDataB64(plaintext []byte) ([]byte, error) {
	data, err := k.EncryptData(plaintext)
	if err != nil {
		return nil, err
	}

	//dataStr := b64.StdEncoding.EncodeToString(data)
	buf := bytes.Buffer{}
	encoder := b64.NewEncoder(b64.URLEncoding, &buf)
	encoder.Write(data)
	encoder.Close()

	return buf.Bytes(), nil
}

func (k *Coder) DecryptDataB64(data []byte) ([]byte, error) {
	//decoded, err := b64.StdEncoding.DecodeString(ciphertext)
	buf := bytes.NewBufferString(string(data))
	decoder := b64.NewDecoder(b64.URLEncoding, buf)
	decoded, err := ioutil.ReadAll(decoder)
	if err != nil {
		return nil, err
	}

	data, err = k.DecryptData(decoded)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (k *Coder) EncryptB64Zip(data []byte) ([]byte, error) {
	data = k.ZipData(data)

	data, err := k.EncryptDataB64(data)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (k *Coder) DecryptB64Zip(data []byte) ([]byte, error) {
	data, err := k.DecryptDataB64(data)
	if err != nil {
		return nil, err
	}

	data, err = k.UnzipData(data)
	if err != nil {
		return nil, err
	}

	return data, err
}
