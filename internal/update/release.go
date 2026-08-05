package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
)

const (
	repoOwner        = "mathismqn"
	repoName         = "godeez"
	latestReleaseURL = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
	checksumsAsset   = "checksums.txt"
	maxResponseSize  = 1 << 20
)

var githubAPIHeaders = map[string]string{
	"Accept":               "application/vnd.github+json",
	"X-GitHub-Api-Version": "2022-11-28",
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func (u *Updater) Latest(ctx context.Context) (*Release, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	body, err := u.get(ctx, latestReleaseURL, githubAPIHeaders)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var release Release
	if err := json.NewDecoder(io.LimitReader(body, maxResponseSize)).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("release has no tag name")
	}

	return &release, nil
}

func (r *Release) Version() string {
	return trimV(r.TagName)
}

func (r *Release) asset(name string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}

	return Asset{}, false
}

func (r *Release) assetForRuntime() (Asset, error) {
	name := fmt.Sprintf("%s_%s_%s_%s", repoName, r.Version(), runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	asset, ok := r.asset(name)
	if !ok {
		return Asset{}, fmt.Errorf("release %s has no binary for %s/%s (expected %s)",
			r.TagName, runtime.GOOS, runtime.GOARCH, name)
	}

	return asset, nil
}
