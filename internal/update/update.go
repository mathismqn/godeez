// Package update handles both halves of keeping godeez current: the passive
// background check that tells the user a newer release exists, and the
// `godeez update` command that installs it.
//
// Releases come from the GitHub releases API. Downloads are verified against
// the published checksums file before anything replaces the running binary,
// and installs owned by a package manager are refused rather than overwritten.
package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mathismqn/godeez/internal/buildinfo"
)

const (
	// Timeouts are per request rather than for the whole operation, so a slow
	// but progressing download is not killed part way. The generous download
	// timeout covers a binary of a few tens of megabytes on a poor connection.
	apiTimeout      = 30 * time.Second
	downloadTimeout = 5 * time.Minute

	tmpPattern = ".godeez-update-*"
)

type Updater struct {
	client *http.Client
	Out    io.Writer
}

// New returns an Updater that reports nothing. Callers that want the step by
// step progress, such as the update command, set Out themselves; the
// background check leaves it discarding so it cannot write over the download
// output.
func New() *Updater {
	return &Updater{
		client: &http.Client{},
		Out:    io.Discard,
	}
}

func (u *Updater) step(format string, args ...any) {
	fmt.Fprintf(u.Out, format+"...\n", args...)
}

func (u *Updater) get(ctx context.Context, url string, headers map[string]string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", buildinfo.UserAgent())
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()

		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	return resp.Body, nil
}
