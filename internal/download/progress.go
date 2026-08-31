package download

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/mathismqn/godeez/internal/deezer"
)

// skipReason is why a track was passed over. It travels as a value rather
// than as a finished sentence so that the wording stays in this file, the
// same way it does for err.
type skipReason int

const (
	skipNone skipReason = iota
	skipAlreadyDownloaded
	skipPersonalUpload
)

type downloadResult struct {
	skip     skipReason
	path     string
	warnings []string
	err      error
}

func (r downloadResult) skipMessage() string {
	switch r.skip {
	case skipAlreadyDownloaded:
		return fmt.Sprintf("Already exists at: %s", r.path)
	case skipPersonalUpload:
		return "Personal upload, not available for download"
	}

	return ""
}

type downloadStats struct {
	downloaded int
	skipped    int
	failed     int
	warnings   int
}

type progressTracker struct {
	stats       downloadStats
	totalTracks int
	kind        deezer.Kind
}

func newProgressTracker(totalTracks int, kind deezer.Kind) *progressTracker {
	return &progressTracker{
		totalTracks: totalTracks,
		kind:        kind,
	}
}

// startLedgerScan announces the one time pass that rewrites records written
// before the ledger tracked file sizes. It is only called once that pass has
// to read the library, which can stall for minutes; a database that only
// needs stat-ing is upgraded silently.
func startLedgerScan() *spinner.Spinner {
	return newSpinner(" Upgrading the download database")
}

func finishLedgerScan(sp *spinner.Spinner) {
	sp.Stop()
	fmt.Print("✔ Download database upgraded\n\n")
}

func (pt *progressTracker) startDownload(index int, track *deezer.Track) *spinner.Spinner {
	trackProgress := fmt.Sprintf("[%d/%d]", index+1, pt.totalTracks)

	sp := newSpinner(fmt.Sprintf(" Downloading: %s - %s", track.Artist, track.FullTitle()))
	sp.Prefix = trackProgress + " "

	return sp
}

// newSpinner starts a spinner on stdout, so that every long step of a run
// shares one look.
func newSpinner(suffix string) *spinner.Spinner {
	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Writer = os.Stdout
	sp.Suffix = suffix
	sp.Start()

	return sp
}

func (pt *progressTracker) handleResult(index int, track *deezer.Track, result downloadResult) {
	trackProgress := fmt.Sprintf("[%d/%d]", index+1, pt.totalTracks)
	trackTitle := track.FullTitle()

	if result.skip != skipNone {
		pt.stats.skipped++
		fmt.Printf("%s ↷ Skipped: %s - %s\n    %s\n",
			trackProgress, track.Artist, trackTitle, result.skipMessage())
		return
	}

	if result.err != nil {
		pt.stats.failed++
		fmt.Printf("%s ✖ Failed: %s - %s:\n    Error: %v\n",
			trackProgress, track.Artist, trackTitle, result.err)
		return
	}

	pt.stats.downloaded++
	if len(result.warnings) > 0 {
		pt.stats.warnings++
	}

	symbol := "✔"
	if len(result.warnings) > 0 {
		symbol = "⚠"
	}
	fmt.Printf("%s %s Downloaded: %s - %s\n", trackProgress, symbol, track.Artist, trackTitle)

	for _, w := range result.warnings {
		fmt.Printf("    Warning: %s\n", w)
	}
}

func (pt *progressTracker) printSummary(outputDir string, elapsed time.Duration) {
	if pt.kind != deezer.KindTrack {
		warningsLine := ""
		if pt.stats.warnings > 0 {
			warningsLine = fmt.Sprintf("\nWarnings:       %d", pt.stats.warnings)
		}
		fmt.Printf(`
================== [ Summary ] ==================
Downloaded:     %d
Skipped:        %d
Failed:         %d%s
Elapsed time:   %s
Files saved to: %s
=================================================
`,
			pt.stats.downloaded,
			pt.stats.skipped,
			pt.stats.failed,
			warningsLine,
			elapsed.Round(time.Second),
			outputDir,
		)

		if pt.stats.downloaded > 0 {
			pt.showSupportMessage()
		}
	}
}

// showSupportMessage nudges the user to star the repository, but only on
// roughly one run in ten.
func (*progressTracker) showSupportMessage() {
	if rand.Float64() < 0.1 {
		fmt.Println("\n⭐ Enjoying GoDeez? Star it on GitHub: https://github.com/mathismqn/godeez")
	}
}
