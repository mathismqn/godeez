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

var (
	version = devVersion
	commit  = ""
	date    = ""
)

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
