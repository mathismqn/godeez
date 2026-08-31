package download

import (
	"fmt"
	"strconv"

	"github.com/flytam/filenamify"
	"github.com/mathismqn/godeez/internal/deezer"
)

// formatExts maps an output format to its extension. mp3 is the default
// rather than an entry, since every other quality is one of its bitrates.
var formatExts = map[string]string{"FLAC": "flac", "WAV": "wav"}

func formatExt(format string) string {
	if ext, ok := formatExts[format]; ok {
		return ext
	}
	return "mp3"
}

// trackFilename builds the on-disk name for a track, sanitised for the
// current filesystem. Album downloads get a zero padded track number prefix
// so the directory sorts in playing order; the other kinds have no meaningful
// ordering to preserve.
//
// A multi disc album is prefixed with its disc as well, since Deezer restarts
// the track number on every disc and the second disc would otherwise reuse
// the first disc's numbers.
func trackFilename(track *deezer.Track, kind deezer.Kind, format string, discs discLayout) string {
	ext := formatExt(format)

	prefix := ""
	if kind == deezer.KindAlbum {
		prefix = padNumber(string(track.TrackNumber)) + ". "
		if discs.multiDisc() {
			prefix = discNumber(track) + "-" + prefix
		}
	}

	base := fmt.Sprintf("%s%s - %s", prefix, track.Artist, track.FullTitle())
	base, _ = filenamify.Filenamify(base, filenamify.Options{MaxLength: 255})
	// 255 bytes is the per-component limit on ext4 and APFS. The budget also
	// has to cover the extension, its dot, and the "-id3v2" suffix the tagging
	// library appends to its temporary file: without that headroom, tagging a
	// long title fails after the download has already succeeded.
	base = truncateBytes(base, 255-len(ext)-1-len("-id3v2"))

	return base + "." + ext
}

// padNumber zero pads a track number to two digits, falling back to the raw
// value when Deezer sends something that is not a number.
func padNumber(value string) string {
	if n, err := strconv.Atoi(value); err == nil {
		return fmt.Sprintf("%02d", n)
	}

	return value
}

// truncateBytes shortens s to at most maxLen bytes without splitting a rune.
// The limit is in bytes because that is what filesystems enforce, but cutting
// mid-rune would leave an invalid UTF-8 name, so it backs up to the last rune
// boundary that fits.
func truncateBytes(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}

	last := 0
	for i := range s {
		if i > maxLen {
			return s[:last]
		}
		last = i
	}
	return s[:last]
}
