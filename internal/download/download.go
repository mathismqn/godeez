// Package download drives the end to end download of a Deezer resource.
//
// Run fetches the resource, then walks its tracks in order: resolve a media
// source, decide whether the track can be skipped, stream and decrypt it,
// optionally convert to wav, write tags, and record the result in the store
// so a later run can skip it. Tracks are processed one at a time.
//
// Most per-track failures are collected as warnings rather than aborting the
// run, so an unavailable cover or a failed BPM lookup does not cost the user
// the rest of an album. Only context cancellation stops the loop early.
package download

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fsutil"
	"github.com/mathismqn/godeez/internal/store"
)

type Downloader struct {
	appConfig    *config.Config
	store        *store.Store
	kind         deezer.Kind
	deezerClient *deezer.Client

	hashIndexOnce sync.Once
	hashIndex     *hashIndex
	hashIndexErr  error
}

func New(appConfig *config.Config, st *store.Store, kind deezer.Kind) *Downloader {
	return &Downloader{
		appConfig: appConfig,
		store:     st,
		kind:      kind,
	}
}

// Run downloads every track of the resource identified by id. opts is
// expected to have passed Validate already, which the cmd package does while
// parsing flags.
func (d *Downloader) Run(ctx context.Context, opts Options, id string) error {
	if err := d.initDeezerClient(ctx, opts); err != nil {
		return err
	}

	resource, outputDir, err := d.prepareResource(ctx, id, opts)
	if err != nil {
		return err
	}

	return d.downloadAllTracks(ctx, resource, opts, outputDir)
}

// initDeezerClient authenticates and rejects quality settings the account
// cannot serve.
//
// The check runs against sourceQuality rather than the raw option because wav
// is produced locally from a flac source, so it carries the same premium
// requirement as flac. mp3_128 is the only format available without a
// subscription. Failing here keeps the user from watching a whole album
// download at a silently downgraded quality.
func (d *Downloader) initDeezerClient(ctx context.Context, opts Options) error {
	var err error
	d.deezerClient, err = deezer.NewClient(ctx, d.appConfig.ARLCookie)
	if err != nil {
		return err
	}

	if !d.deezerClient.Session.Premium && opts.sourceQuality() != "mp3_128" {
		return fmt.Errorf("premium account required for '%s' quality", opts.Quality)
	}

	return nil
}

// prepareResource fetches the resource, applies the artist track limit, and
// makes sure the output directory exists.
//
// The limit only applies to artists because that is the one kind whose track
// list is unbounded: it is the artist's top tracks, not a finite album or
// playlist.
//
// Sweeping the part files last clears leftovers from a previous run that was
// killed mid-write. They are ignorable on their own, but they accumulate and
// would otherwise be mistaken for real downloads.
func (d *Downloader) prepareResource(ctx context.Context, id string, opts Options) (deezer.Resource, string, error) {
	resource, err := d.deezerClient.FetchResource(ctx, d.kind, id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch resource: %w", err)
	}

	tracks := resource.Tracks()
	if len(tracks) == 0 {
		if d.kind == deezer.KindTrack {
			return nil, "", fmt.Errorf("track with ID %s not found", id)
		}
		return nil, "", fmt.Errorf("%s has no tracks", d.kind)
	}

	if d.kind == deezer.KindArtist && len(tracks) > opts.Limit {
		resource.SetTracks(tracks[:opts.Limit])
	}

	outputDir := resource.OutputDir(d.appConfig.OutputDir)
	if err := fsutil.EnsureDir(outputDir); err != nil {
		return nil, "", fmt.Errorf("failed to create output directory: %w", err)
	}
	sweepPartFiles(outputDir)

	return resource, outputDir, nil
}

// downloadAllTracks runs the per-track pipeline over the whole resource and
// prints the summary.
//
// Cancellation is checked both before each track and against the result,
// because a track cancelled mid-stream surfaces the error through the result
// rather than through ctx. Any other per-track error is recorded and the loop
// continues.
func (d *Downloader) downloadAllTracks(ctx context.Context, resource deezer.Resource, opts Options, outputDir string) error {
	tracks := resource.Tracks()
	startTime := time.Now()

	if d.kind != deezer.KindTrack {
		fmt.Printf("%s\n\nStarting download...\n\n", resourceInfo(resource))
	}

	progress := newProgressTracker(len(tracks), d.kind)

	for i, track := range tracks {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		sp := progress.startDownload(i, track)
		result := d.downloadTrack(ctx, resource, track, opts, outputDir)
		sp.Stop()

		if result.err != nil && errors.Is(result.err, context.Canceled) {
			return result.err
		}

		progress.handleResult(i, track, result)
	}

	progress.printSummary(outputDir, time.Since(startTime))

	return nil
}
