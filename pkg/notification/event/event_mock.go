package event

import "github.com/stretchr/testify/mock"

type EventMock struct {
	mock.Mock
}

func (m *EventMock) Type() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *EventMock) User() UserData {
	args := m.Called()
	return args.Get(0).(UserData)
}

func (m *EventMock) Content() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}
