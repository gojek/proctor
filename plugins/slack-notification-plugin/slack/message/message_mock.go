package message

import (
	"github.com/stretchr/testify/mock"
)

type MessageMock struct {
	mock.Mock
}

func (m *MessageMock) JSON() (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}
