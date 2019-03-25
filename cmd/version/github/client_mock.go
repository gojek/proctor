package github

import (
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) LatestRelease(owner, repository string) (release string, err error) {
	args := m.Called(owner, repository)
	return args.Get(0).(string), args.Error(1)
}
