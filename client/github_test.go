package client

import (
	"testing"
	"context"
)

func TestGithubClient_IsLatestRelease(t *testing.T) {
	githubClient := NewGithubClient()
	_, _, _ = githubClient.client.Repositories.GetLatestRelease(context.Background(), "gojektech", "proctor")
}