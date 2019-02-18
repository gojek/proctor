package client

import (
	"github.com/google/go-github/github"
	"context"
	"github.com/gojektech/proctor/cmd/version"
)

type LatestReleaseChecker interface {
	IsLatestRelease() bool
}

type GithubClient struct {
	client *github.Client
}

func NewGithubClient() *GithubClient {
	return &GithubClient{github.NewClient(nil)}
}

func (gc *GithubClient) IsLatestRelease() (bool, error) {
	release, _, err := gc.client.Repositories.GetLatestRelease(context.Background(), "gojektech", "proctor")
	return version.ClientVersion == *release.TagName, err
}
