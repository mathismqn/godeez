package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/provider"
)

type bpmKey struct {
	BPM string
	Key string
}

type metadataResult struct {
	bpmKey   bpmKey
	genre    string
	warnings []string
}

func fetchMetadata(httpClient *http.Client, ctx context.Context, song *deezer.Song, opts Options) metadataResult {
	if !opts.BPM && !opts.Genre {
		return metadataResult{}
	}

	type bpmResult struct {
		value bpmKey
		err   error
	}
	type genreResult struct {
		value string
		err   error
	}

	bpmChan := make(chan bpmResult, 1)
	genreChan := make(chan genreResult, 1)

	if opts.BPM {
		go func() {
			result, err := provider.FetchBPM(ctx, httpClient, song.Artist, song.Title, song.Duration)
			bpmChan <- bpmResult{value: bpmKey{BPM: result.BPM, Key: result.Key}, err: err}
		}()
	}

	if opts.Genre {
		go func() {
			genre, err := provider.FetchGenre(ctx, httpClient, song.Artist, song.GetTitle())
			genreChan <- genreResult{value: genre, err: err}
		}()
	}

	var result metadataResult

	if opts.BPM {
		r := <-bpmChan
		if r.err != nil {
			if !errors.Is(r.err, context.Canceled) {
				result.warnings = append(result.warnings, fmt.Sprintf("failed to fetch BPM and key: %v", r.err))
			}
		} else {
			result.bpmKey = r.value
		}
	}

	if opts.Genre {
		r := <-genreChan
		if r.err != nil {
			if !errors.Is(r.err, context.Canceled) {
				result.warnings = append(result.warnings, fmt.Sprintf("failed to fetch genre: %v", r.err))
			}
		} else {
			result.genre = r.value
		}
	}

	return result
}
