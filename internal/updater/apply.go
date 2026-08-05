package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/mathismqn/godeez/internal/fsutil"
)

var managedPrefixes = []string{
	"/nix/store",
	"/opt/homebrew",
	"/usr/local/Cellar",
	"/home/linuxbrew",
	"/snap",
	"/var/lib/flatpak",
}

func resolveTarget() (string, error) {
	if buildinfo.IsDev() {
		return "", fmt.Errorf("development build cannot self-update. Install a release from https://github.com/%s/%s/releases",
			repoOwner, repoName)
	}

	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to locate the running binary: %w", err)
	}

	target := exe
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		target = resolved
	}

	for _, prefix := range managedPrefixes {
		if strings.HasPrefix(target, prefix) {
			return "", fmt.Errorf("%s was installed by a package manager. Update it with that instead", target)
		}
	}

	return target, nil
}

func CheckUpdatable() error {
	_, err := resolveTarget()

	return err
}

func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		hint := "Re-run with sudo"
		if runtime.GOOS == "windows" {
			hint = "Re-run from an elevated prompt"
		}

		return fmt.Errorf("cannot write to %s: %w. %s", dir, err, hint)
	}

	name := f.Name()
	f.Close()
	os.Remove(name)

	return nil
}

func (u *Updater) Apply(ctx context.Context, release *Release) error {
	target, err := resolveTarget()
	if err != nil {
		return err
	}

	dir := filepath.Dir(target)
	if err := checkWritable(dir); err != nil {
		return err
	}

	asset, err := release.assetForRuntime()
	if err != nil {
		return err
	}

	want, err := u.fetchChecksum(ctx, release, asset.Name)
	if err != nil {
		return err
	}

	u.step("Downloading %s", asset.Name)
	tmp, sum, err := u.download(ctx, dir, asset)
	if err != nil {
		return err
	}
	defer fsutil.Remove(tmp)

	u.step("Verifying checksum")
	if sum != want {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", asset.Name, want, sum)
	}

	if err := os.Chmod(tmp, 0755); err != nil {
		return err
	}

	u.step("Replacing %s", target)

	return replaceBinary(target, tmp)
}

func (u *Updater) fetchChecksum(ctx context.Context, release *Release, assetName string) (string, error) {
	asset, ok := release.asset(checksumsAsset)
	if !ok {
		return "", fmt.Errorf("release %s does not publish %s", release.TagName, checksumsAsset)
	}

	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	body, err := u.get(ctx, asset.URL, nil)
	if err != nil {
		return "", err
	}
	defer body.Close()

	return parseChecksums(io.LimitReader(body, maxResponseSize), assetName)
}

func parseChecksums(r io.Reader, name string) (string, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		if strings.TrimPrefix(fields[1], "*") == name {
			return strings.ToLower(fields[0]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no checksum listed for %s", name)
}

func (u *Updater) download(ctx context.Context, dir string, asset Asset) (string, string, error) {
	body, err := u.get(ctx, asset.URL, nil)
	if err != nil {
		return "", "", err
	}
	defer body.Close()

	f, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		return "", "", err
	}
	tmp := f.Name()

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(f, hash), body); err != nil {
		f.Close()
		fsutil.Remove(tmp)

		return "", "", fmt.Errorf("failed to download %s: %w", asset.Name, err)
	}
	if err := f.Close(); err != nil {
		fsutil.Remove(tmp)

		return "", "", err
	}

	return tmp, hex.EncodeToString(hash.Sum(nil)), nil
}

func replaceBinary(target, tmp string) error {
	if runtime.GOOS != "windows" {
		return os.Rename(tmp, target)
	}

	old := target + ".old"
	os.Remove(old)

	if err := os.Rename(target, old); err != nil {
		return fmt.Errorf("failed to move the current binary aside: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		if rollbackErr := os.Rename(old, target); rollbackErr != nil {
			return fmt.Errorf("failed to install the new binary: %w. The previous one could not be restored from %s: %v",
				err, old, rollbackErr)
		}

		return fmt.Errorf("failed to install the new binary: %w", err)
	}

	os.Remove(old)

	return nil
}
