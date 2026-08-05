package deezer

import (
	"bytes"
	"testing"
)

func TestZeroPad(t *testing.T) {
	tests := []struct {
		length int
		want   int
	}{
		{0, 0},
		{1, 16},
		{15, 16},
		{16, 16},
		{17, 32},
	}

	for _, tt := range tests {
		padded := zeroPad(make([]byte, tt.length))
		if len(padded) != tt.want {
			t.Errorf("zeroPad(len %d) = len %d, want %d", tt.length, len(padded), tt.want)
		}
	}
}

func TestZeroPadPreservesData(t *testing.T) {
	data := []byte("secret")
	padded := zeroPad(data)

	if !bytes.Equal(padded[:len(data)], data) {
		t.Errorf("zeroPad changed data: got %q", padded[:len(data)])
	}
	for _, b := range padded[len(data):] {
		if b != 0 {
			t.Errorf("padding is not zero: %v", padded)
		}
	}
}

func TestECBRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef")
	plaintext := zeroPad([]byte("some secret data"))

	encrypted, err := ecbEncrypt(key, plaintext)
	if err != nil {
		t.Fatalf("ecbEncrypt: %v", err)
	}
	if bytes.Equal(encrypted, plaintext) {
		t.Fatal("encrypted data equals plaintext")
	}

	decrypted, err := ecbDecrypt(key, encrypted)
	if err != nil {
		t.Fatalf("ecbDecrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("round trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestECBRejectsPartialBlock(t *testing.T) {
	key := []byte("0123456789abcdef")

	if _, err := ecbEncrypt(key, make([]byte, 15)); err == nil {
		t.Error("expected error for data not a multiple of the block size")
	}
	if _, err := ecbDecrypt(key, make([]byte, 17)); err == nil {
		t.Error("expected error for data not a multiple of the block size")
	}
}
