package notification

import (
	"github.com/stretchr/testify/mock"
	"proctor/pkg/notification/event"
)

type ObserverMock struct {
	mock.Mock
}

func (m *ObserverMock) OnNotify(evt event.Event) error {
	args := m.Called(evt)
	return args.Error(0)
}
