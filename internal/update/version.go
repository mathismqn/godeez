package update

import (
	"strings"

	"golang.org/x/mod/semver"
)

func trimV(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func canonical(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if !semver.IsValid(v) {
		return ""
	}

	return v
}

func IsNewer(current, latest string) bool {
	c, l := canonical(current), canonical(latest)
	if c == "" || l == "" {
		return false
	}

	return semver.Compare(l, c) > 0
}
