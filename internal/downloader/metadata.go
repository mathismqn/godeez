package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/provider"
)

type metadataResult struct {
	bpmKey   provider.BPMKey
	genre    string
	warnings []string
}

type metadataFetcher struct {
	httpClient *http.Client
}

func newMetadataFetcher(httpClient *http.Client) *metadataFetcher {
	return &metadataFetcher{
		httpClient: httpClient,
	}
}

func (mf *metadataFetcher) fetch(ctx context.Context, song *deezer.Song, opts Options) metadataResult {
	result := metadataResult{
		bpmKey:   provider.BPMKey{},
		genre:    "",
		warnings: []string{},
	}

	if !opts.BPM && !opts.Genre {
		return result
	}

	bmpChan := make(chan provider.BPMKey, 1)
	bmpErrChan := make(chan error, 1)
	genreChan := make(chan string, 1)
	genreErrChan := make(chan error, 1)

	if opts.BPM {
		go func() {
			p := provider.BPMProvider{}
			bmpKey, err := p.Fetch(ctx, mf.httpClient, song.Artist, song.Title, song.Duration)
			if err != nil {
				bmpErrChan <- err
			} else {
				bmpChan <- bmpKey
			}
		}()
	}

	if opts.Genre {
		go func() {
			p := provider.GenreProvider{}
			genre, err := p.Fetch(ctx, mf.httpClient, song.Artist, song.GetTitle())
			if err != nil {
				genreErrChan <- err
			} else {
				genreChan <- genre
			}
		}()
	}

	if opts.BPM {
		select {
		case bmpKey := <-bmpChan:
			result.bpmKey = bmpKey
		case err := <-bmpErrChan:
			if !errors.Is(err, context.Canceled) {
				result.warnings = append(result.warnings, fmt.Sprintf("failed to fetch BPM and key: %v", err))
			}
		}
	}

	if opts.Genre {
		select {
		case genre := <-genreChan:
			result.genre = genre
		case err := <-genreErrChan:
			if !errors.Is(err, context.Canceled) {
				result.warnings = append(result.warnings, fmt.Sprintf("failed to fetch genre: %v", err))
			}
		}
	}

	return result
}
