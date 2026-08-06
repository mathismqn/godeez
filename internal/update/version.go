package update

import (
	"strings"

	"golang.org/x/mod/semver"
)

func trimV(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// canonical normalises a version for comparison, returning "" if it is not
// valid semver. Tags carry a leading "v" and buildinfo reports versions
// without one, so the prefix is added back before validating rather than
// requiring callers to agree on a spelling.
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

// IsNewer reports whether latest is a strictly newer release than current.
//
// An unparseable version on either side yields false rather than an error or
// a guess: this decides whether to nag the user about an update, and staying
// quiet is the right failure mode when the comparison is meaningless.
func IsNewer(current, latest string) bool {
	c, l := canonical(current), canonical(latest)
	if c == "" || l == "" {
		return false
	}

	return semver.Compare(l, c) > 0
}
