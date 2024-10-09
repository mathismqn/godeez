package crypto

import (
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/mathismqn/godeez/internal/config"
	"golang.org/x/crypto/blowfish"
)

func GetBlowfishKey(songID string) []byte {
	hash := md5.Sum([]byte(songID))
	hashHex := fmt.Sprintf("%x", hash)

	key := []byte(config.Cfg.SecretKey)
	for i := 0; i < len(hash); i++ {
		key[i] = key[i] ^ hashHex[i] ^ hashHex[i+16]
	}

	return key
}

func DecryptBlowfish(data, key []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv, err := hex.DecodeString(config.Cfg.IV)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	return decrypted, nil
}
