#!/usr/bin/env bash
#
# Extracts the CHANGELOG.md section for a version and formats it as the GitHub
# release body. Usage: release-notes.sh v1.5.0
#
# Exits non-zero when the version has no section, so a tag can never be
# published with empty or stale release notes.

set -euo pipefail

if [ $# -ne 1 ]; then
	echo "usage: $0 <version>" >&2
	exit 2
fi

version="${1#v}"
changelog="${CHANGELOG_FILE:-CHANGELOG.md}"

if [ ! -f "$changelog" ]; then
	echo "$changelog not found" >&2
	exit 1
fi

notes=$(awk -v ver="$version" '
	BEGIN { heading = "## [" ver "]" }
	index($0, heading) == 1 { found = 1; next }
	found && index($0, "## [") == 1 { exit }
	found { print }
' "$changelog")

if [ -z "${notes//[[:space:]]/}" ]; then
	echo "no $changelog section found for version $version" >&2
	exit 1
fi

printf "## What's new in v%s\n%s\n" "$version" "$notes"
