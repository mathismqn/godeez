package update

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

// managedPrefixes are install roots owned by a package manager. Overwriting a
// binary there would leave the package manager's database describing a file
// that no longer matches, and its next upgrade would silently revert the
// self-update. Users on these installs are pointed back at their package
// manager instead.
var managedPrefixes = []string{
	"/nix/store",
	"/opt/homebrew",
	"/usr/local/Cellar",
	"/home/linuxbrew",
	"/snap",
	"/var/lib/flatpak",
}

// resolveTarget returns the binary that should be replaced, or an error
// explaining why self-updating is not appropriate here.
//
// Symlinks are resolved first so the real file is replaced rather than the
// link: package managers commonly expose a binary through a symlink, and
// following it is what makes the managed prefix check meaningful. A build
// that was not produced by a release is refused outright, since there is no
// version to compare against.
func resolveTarget() (string, error) {
	if buildinfo.IsDev() {
		return "", fmt.Errorf("development build cannot self-update; install a release from https://github.com/%s/%s/releases",
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
		if target == prefix || strings.HasPrefix(target, prefix+"/") {
			return "", fmt.Errorf("%s was installed by a package manager; update it with that instead", target)
		}
	}

	return target, nil
}

func CheckUpdatable() error {
	_, err := resolveTarget()

	return err
}

// checkWritable proves the install directory is writable by actually creating
// and removing a file there. Inspecting permission bits would not account for
// read-only mounts or the platform's own rules, and finding out only after
// the download has finished wastes the user's time.
func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		hint := "re-run with sudo"
		if runtime.GOOS == "windows" {
			hint = "re-run from an elevated prompt"
		}

		return fmt.Errorf("cannot write to %s: %w; %s", dir, err, hint)
	}

	name := f.Name()
	f.Close()
	os.Remove(name)

	return nil
}

// Apply downloads release and replaces the running binary with it.
//
// The order of these steps is the safety property. The expected checksum is
// fetched before the asset, so a release that does not publish one fails
// before anything is downloaded. The download lands in a temporary file in
// the install directory, which keeps the final rename on the same filesystem
// and therefore atomic. The binary is only replaced after the checksum
// matches, so a corrupted or tampered download can never be executed.
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

	return u.replaceBinary(target, tmp)
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

// parseChecksums finds the digest for name in a sha256sum style file.
//
// The optional "*" before the filename is the marker sha256sum uses for
// binary mode and is not part of the name.
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

// download writes asset to a temporary file in dir and returns its path and
// sha256. The hash is computed while streaming, so the file is never read a
// second time and never has to be held in memory.
func (u *Updater) download(ctx context.Context, dir string, asset Asset) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

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

// replaceBinary swaps the new binary into place.
//
// Unix lets a running executable be renamed over, so a single atomic rename
// is enough. Windows locks the file of a running process, so the current
// binary has to be moved aside first, which leaves a window where the target
// does not exist; if installing the replacement then fails, the old one is
// moved back. The .old file is removed on the next update rather than
// immediately, since it is still locked while this process runs.
func (u *Updater) replaceBinary(target, tmp string) error {
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
			return fmt.Errorf("failed to install the new binary: %w; the previous one could not be restored from %s: %v",
				err, old, rollbackErr)
		}

		return fmt.Errorf("failed to install the new binary: %w", err)
	}

	os.Remove(old)

	return nil
}
