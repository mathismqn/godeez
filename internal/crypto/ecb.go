package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func ZeroPad(data []byte) []byte {
	bs := aes.BlockSize
	padded := make([]byte, len(data)+(bs-len(data)%bs)%bs)
	copy(padded, data)

	return padded
}

func EncryptECB(key, data []byte) ([]byte, error) {
	return ecbTransform(key, data, (cipher.Block).Encrypt)
}

func DecryptECB(key, data []byte) ([]byte, error) {
	return ecbTransform(key, data, (cipher.Block).Decrypt)
}

func ecbTransform(key, data []byte, op func(cipher.Block, []byte, []byte)) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return nil, fmt.Errorf("data length %d is not a multiple of the AES block size", len(data))
	}

	out := make([]byte, len(data))
	for i := 0; i < len(data); i += bs {
		op(block, out[i:i+bs], data[i:i+bs])
	}

	return out, nil
}
