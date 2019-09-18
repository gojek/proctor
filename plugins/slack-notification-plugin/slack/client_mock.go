package slack

import (
	"github.com/stretchr/testify/mock"
	"proctor/plugins/slack-notification-plugin/slack/message"
)

type SlackClientMock struct {
	mock.Mock
}

func (m *SlackClientMock) Publish(message message.Message) error {
	args := m.Called(message)
	return args.Error(0)
}
