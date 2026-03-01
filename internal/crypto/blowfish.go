package crypto

import (
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"

	"golang.org/x/crypto/blowfish"
)

var (
	iv        = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	secretKey = []byte("g4el58wc0zvf9na1")
)

func GetKey(songID string) []byte {
	hash := md5.Sum([]byte(songID))
	hashHex := hex.EncodeToString(hash[:])

	key := make([]byte, len(secretKey))
	copy(key, secretKey)
	for i := 0; i < len(hash); i++ {
		key[i] = key[i] ^ hashHex[i] ^ hashHex[i+16]
	}

	return key
}

func Decrypt(data, key []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(decrypted, data)

	return decrypted, nil
}
