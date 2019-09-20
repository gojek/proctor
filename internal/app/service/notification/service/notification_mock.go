package service

import (
	"github.com/stretchr/testify/mock"
	"proctor/pkg/notification/event"
)

type NotificationServiceMock struct {
	mock.Mock
}

func (m NotificationServiceMock) Notify(evt event.Event) {
	m.Called(evt)
}
