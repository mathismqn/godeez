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
	apiTimeout = 30 * time.Second
	tmpPattern = ".godeez-update-*"
)

type Updater struct {
	client *http.Client
	Out    io.Writer
}

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
