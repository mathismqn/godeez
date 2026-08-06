// Package buildinfo reports the version, commit and build date of the running
// binary.
//
// Release builds have these stamped in by goreleaser through -ldflags. When
// that has not happened, as with `go build` or `go install`, the values are
// recovered from the module metadata the toolchain embeds. Anything that
// cannot be established as a real release is reported as a development build,
// which is what disables the update machinery.
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// devVersion is the version reported by builds that were not produced by a
// release. Both the update check and `godeez update` refuse to run on them.
const devVersion = "dev"

// Injected at link time by goreleaser. They are unexported and read through
// the accessors below so nothing can depend on their zero values directly.
var (
	version = devVersion
	commit  = ""
	date    = ""
)

// Version returns the release version without a leading "v", or devVersion
// for anything that is not a release build. Binaries built with `go install`
// carry no ldflags but do record the module version, so that is consulted
// before giving up.
func Version() string {
	if version != devVersion {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if v := releaseVersion(info.Main.Version); v != "" {
			return v
		}
	}

	return devVersion
}

// releaseVersion accepts v only if it names a published release, returning ""
// otherwise.
//
// Pseudo-versions describe a commit that was never tagged, and a build suffix
// marks a local or modified build. Treating either as a release would offer
// the user an update path from a version that does not exist.
func releaseVersion(v string) string {
	if !semver.IsValid(v) {
		return ""
	}
	if semver.Build(v) != "" || module.IsPseudoVersion(v) {
		return ""
	}

	return strings.TrimPrefix(v, "v")
}

func IsDev() bool {
	return Version() == devVersion
}

// Commit returns the revision the binary was built from, falling back to the
// VCS stamp the Go toolchain records when building inside a repository. It
// returns "" when neither is available.
func Commit() string {
	if commit != "" {
		return commit
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				return s.Value
			}
		}
	}

	return ""
}

func Date() string {
	return date
}

func UserAgent() string {
	return fmt.Sprintf("godeez/%s (%s/%s)", Version(), runtime.GOOS, runtime.GOARCH)
}
