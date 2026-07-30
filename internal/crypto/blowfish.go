package crypto

import (
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"

	"golang.org/x/crypto/blowfish"
)

var (
	blowfishIV        = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	blowfishSecretKey = []byte("g4el58wc0zvf9na1")
)

func GetBlowfishKey(trackID string) []byte {
	hash := md5.Sum([]byte(trackID))
	hashHex := hex.EncodeToString(hash[:])

	key := make([]byte, len(blowfishSecretKey))
	copy(key, blowfishSecretKey)
	for i := range len(hash) {
		key[i] = key[i] ^ hashHex[i] ^ hashHex[i+16]
	}

	return key
}

func DecryptBlowfish(data, key []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, blowfishIV).CryptBlocks(decrypted, data)

	return decrypted, nil
}
