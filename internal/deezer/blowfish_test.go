package deezer

import (
	"bytes"
	"crypto/cipher"
	"encoding/hex"
	"testing"

	"golang.org/x/crypto/blowfish"
)

func TestBlowfishKey(t *testing.T) {
	tests := []struct {
		trackID string
		want    string
	}{
		{"3135556", "6c6c666b39662c37652575603c643439"},
		{"123456789", "6d34656061377f31322a7336393f626b"},
		{"1", "3464656e343a7d3a672c236a33696061"},
	}

	for _, tt := range tests {
		got := hex.EncodeToString(BlowfishKey(tt.trackID))
		if got != tt.want {
			t.Errorf("BlowfishKey(%q) = %s, want %s", tt.trackID, got, tt.want)
		}
	}
}

func TestDecryptBlowfishRoundTrip(t *testing.T) {
	key := BlowfishKey("3135556")
	plaintext := bytes.Repeat([]byte("01234567"), 16)

	block, err := blowfish.NewCipher(key)
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}

	encrypted := make([]byte, len(plaintext))
	cipher.NewCBCEncrypter(block, blowfishIV).CryptBlocks(encrypted, plaintext)

	decrypted, err := DecryptBlowfish(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptBlowfish: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("round trip mismatch: got %x, want %x", decrypted, plaintext)
	}
}

func TestDecryptBlowfishInvalidKey(t *testing.T) {
	if _, err := DecryptBlowfish(make([]byte, 8), nil); err == nil {
		t.Error("expected error for nil key")
	}
}
