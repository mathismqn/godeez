package download

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mathismqn/godeez/internal/deezer"
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

func fetchMetadata(httpClient *http.Client, ctx context.Context, track *deezer.Track, opts Options) metadataResult {
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
			result, err := fetchBPM(ctx, httpClient, track.Artist, track.Title, track.Duration)
			bpmChan <- bpmResult{value: result, err: err}
		}()
	}

	if opts.Genre {
		go func() {
			genre, err := fetchGenre(ctx, httpClient, track.Artist, track.GetTitle())
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
