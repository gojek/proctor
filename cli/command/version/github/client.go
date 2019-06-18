package github

import (
	"context"
	"github.com/google/go-github/github"
)

type LatestReleaseFetcher interface {
	LatestRelease(owner, repository string) (string, error)
}

type client struct {
	client *github.Client
}

func NewClient() *client {
	return &client{github.NewClient(nil)}
}

func (gc *client) LatestRelease(owner, repository string) (string, error) {
	release, _, err := gc.client.Repositories.GetLatestRelease(context.Background(), owner, repository)
	releaseTag := ""

	if err == nil {
		releaseTag = *release.TagName
	}

	return releaseTag, err
}
