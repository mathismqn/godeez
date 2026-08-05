package downloader

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/mathismqn/godeez/internal/deezer"
)

type downloadResult struct {
	skipped  bool
	path     string
	warnings []string
	err      error
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

func (pt *progressTracker) startDownload(index int, track *deezer.Track) *spinner.Spinner {
	trackProgress := fmt.Sprintf("[%d/%d]", index+1, pt.totalTracks)

	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Writer = os.Stdout
	sp.Prefix = trackProgress + " "
	sp.Suffix = fmt.Sprintf(" Downloading: %s - %s", track.Artist, track.GetTitle())
	sp.Start()

	return sp
}

func (pt *progressTracker) handleResult(index int, track *deezer.Track, result downloadResult) {
	trackProgress := fmt.Sprintf("[%d/%d]", index+1, pt.totalTracks)
	trackTitle := track.GetTitle()

	if result.skipped {
		pt.stats.skipped++
		fmt.Printf("%s ↷ Skipped: %s - %s\n    Already exists at: %s\n",
			trackProgress, track.Artist, trackTitle, result.path)
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

func (*progressTracker) showSupportMessage() {
	if rand.Float64() < 0.1 {
		fmt.Println("\n⭐ Enjoying GoDeez? Star it on GitHub: https://github.com/mathismqn/godeez")
	}
}
