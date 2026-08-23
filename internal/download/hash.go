package download

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// hashFile returns the sha256 of the file at path along with its length.
//
// The size comes from the same read as the hash rather than a separate stat,
// so the two can never describe different versions of the file: whatever
// bytes were hashed are exactly the bytes counted.
func hashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	h := sha256.New()
	size, err := io.Copy(h, file)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(h.Sum(nil)), size, nil
}
