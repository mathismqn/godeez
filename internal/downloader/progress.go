package downloader

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/logger"
)

type downloadStats struct {
	downloaded int
	skipped    int
	failed     int
}

type downloadResult struct {
	success  bool
	skipped  bool
	path     string
	warnings []string
	err      error
}

type progressTracker struct {
	logger       *logger.Logger
	stats        *downloadStats
	totalSongs   int
	resourceType string
}

func newProgressTracker(logger *logger.Logger, totalSongs int, resourceType string) *progressTracker {
	return &progressTracker{
		logger:       logger,
		stats:        &downloadStats{},
		totalSongs:   totalSongs,
		resourceType: resourceType,
	}
}

func (pt *progressTracker) startDownload(index int, song *deezer.Song) *spinner.Spinner {
	songTitle := song.GetTitle()
	trackProgress := fmt.Sprintf("[%d/%d]", index+1, pt.totalSongs)

	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Writer = os.Stdout
	sp.Prefix = trackProgress + " "
	sp.Suffix = fmt.Sprintf(" Downloading: %s - %s", song.Artist, songTitle)
	sp.Start()

	return sp
}

func (pt *progressTracker) handleResult(index int, song *deezer.Song, result downloadResult) {
	trackProgress := fmt.Sprintf("[%d/%d]", index+1, pt.totalSongs)
	songTitle := song.GetTitle()

	if result.skipped {
		pt.stats.skipped++
		fmt.Printf("%s ↷ Skipped: %s - %s\n    Already exists at: %s\n",
			trackProgress, song.Artist, songTitle, result.path)
		return
	}

	if result.err != nil {
		pt.stats.failed++
		pt.logger.Errorf("Failed to download %s - %s: %v\n", song.Artist, songTitle, result.err)
		fmt.Printf("%s ✖ Failed: %s - %s:\n    Error: %v\n",
			trackProgress, song.Artist, songTitle, result.err)
		return
	}

	symbol := "✔"
	if len(result.warnings) > 0 {
		symbol = "⚠"
	}

	pt.stats.downloaded++
	pt.logger.Infof("Downloaded %s - %s\n", song.Artist, songTitle)
	fmt.Printf("%s %s Downloaded: %s - %s\n", trackProgress, symbol, song.Artist, songTitle)

	for _, w := range result.warnings {
		pt.logger.Warnf("Warning: %s\n", w)
		fmt.Printf("    Warning: %s\n", w)
	}
}

func (pt *progressTracker) printSummary(resourceTitle, resourceID, outputDir string, elapsed time.Duration) {
	if pt.stats.downloaded > 0 || pt.stats.failed > 0 {
		pt.logger.Infof("Resource %s (%s): %d downloaded, %d skipped, %d failed\n",
			resourceTitle, resourceID, pt.stats.downloaded, pt.stats.skipped, pt.stats.failed)
	}

	if pt.resourceType != "track" {
		fmt.Printf(`
================== [ Summary ] ==================
Downloaded:     %d
Skipped:        %d
Failed:         %d
Elapsed time:   %s
Files saved to: %s
=================================================
`,
			pt.stats.downloaded,
			pt.stats.skipped,
			pt.stats.failed,
			elapsed.Round(time.Second),
			outputDir,
		)
	}
}
