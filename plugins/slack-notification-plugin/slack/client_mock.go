package slack

import "github.com/stretchr/testify/mock"

type SlackClientMock struct {
	mock.Mock
}

func (m *SlackClientMock) Publish(message Message) error {
	args := m.Called(message)
	return args.Error(0)
}
