package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}

func DeleteFile(path string) error {
	if !FileExists(path) {
		return nil
	}

	return os.Remove(path)
}

func GetFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func FindFileByHash(root, targetHash string) (string, error) {
	var found string

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return nil
		}
		sum := hex.EncodeToString(h.Sum(nil))
		if sum == targetHash {
			found = path
			return filepath.SkipDir
		}
		return nil
	}); err != nil {
		return "", err
	}

	return found, nil
}
