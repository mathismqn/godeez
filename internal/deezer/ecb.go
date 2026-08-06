package deezer

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// zeroPad right pads data with zero bytes to a whole number of AES blocks.
// This is not PKCS#7 and is not unambiguously reversible, but it is what the
// mobile gateway expects for the password field.
func zeroPad(data []byte) []byte {
	bs := aes.BlockSize
	padded := make([]byte, len(data)+(bs-len(data)%bs)%bs)
	copy(padded, data)

	return padded
}

func ecbEncrypt(key, data []byte) ([]byte, error) {
	return ecbTransform(key, data, (cipher.Block).Encrypt)
}

func ecbDecrypt(key, data []byte) ([]byte, error) {
	return ecbTransform(key, data, (cipher.Block).Decrypt)
}

// ecbTransform applies op block by block in ECB mode.
//
// ECB leaks equality between identical plaintext blocks and would be the
// wrong choice for anything designed today, but it is the mode Deezer's
// mobile gateway uses, so interoperating requires it. The standard library
// deliberately ships no ECB mode, which is why this exists. Do not reuse it
// for anything outside the gateway handshake.
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
