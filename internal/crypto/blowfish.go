package crypto

import (
	"crypto/cipher"
	"crypto/md5"
	"fmt"

	"golang.org/x/crypto/blowfish"
)

func GetKey(secretKey, songID string) []byte {
	hash := md5.Sum([]byte(songID))
	hashHex := fmt.Sprintf("%x", hash)

	key := []byte(secretKey)
	for i := 0; i < len(hash); i++ {
		key[i] = key[i] ^ hashHex[i] ^ hashHex[i+16]
	}

	return key
}

func Decrypt(data, key, iv []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	return decrypted, nil
}
