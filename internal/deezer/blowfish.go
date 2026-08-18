package deezer

import (
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"

	"golang.org/x/crypto/blowfish"
)

// blowfishIV and blowfishSecretKey are Deezer's own constants, not values
// chosen by this project. They are the same for every user and every track,
// and are widely published; the per-track key derived from them in
// blowfishKey is what actually varies. Changing either one simply produces
// audio that will not decode.
var (
	blowfishIV        = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	blowfishSecretKey = []byte("g4el58wc0zvf9na1")
)

// blowfishKey derives the per-track decryption key for trackID.
//
// Deezer takes the MD5 of the track ID as a 32 character hex string and folds
// its two halves back into the 16 byte secret, XORing byte i of the secret
// with hex digits i and i+16. The loop therefore runs over the 16 bytes of
// the raw digest, not the 32 characters of its hex encoding, and the result
// is the same 16 byte length as the secret.
func blowfishKey(trackID string) []byte {
	hash := md5.Sum([]byte(trackID))
	hashHex := hex.EncodeToString(hash[:])

	key := make([]byte, len(blowfishSecretKey))
	copy(key, blowfishSecretKey)
	for i := range len(hash) {
		key[i] = key[i] ^ hashHex[i] ^ hashHex[i+16]
	}

	return key
}

// DecryptBlowfish decrypts a single stream chunk with key and returns the
// plaintext.
//
// This is not a whole-file operation. Only every third chunk of a Deezer
// stream is encrypted, so callers are responsible for applying that stripe
// pattern; see streamToTempFile in the download package.
func DecryptBlowfish(data, key []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, blowfishIV).CryptBlocks(decrypted, data)

	return decrypted, nil
}
