package github

import (
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) LatestRelease(owner, repository string) (release *github.RepositoryRelease, err error) {
	args := m.Called(owner, repository)
	return args.Get(0).(*github.RepositoryRelease), args.Error(1)
}
