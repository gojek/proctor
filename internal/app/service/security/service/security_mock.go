package service

import (
	"github.com/stretchr/testify/mock"
	"proctor/pkg/auth"
)

type SecurityServiceMock struct {
	mock.Mock
}

func (m *SecurityServiceMock) initializePlugin() error {
	args := m.Called()
	return args.Error(0)
}

func (m *SecurityServiceMock) Auth(email string, token string) (auth.UserDetail, error) {
	args := m.Called(email, token)
	return args.Get(0).(auth.UserDetail), args.Error(1)
}

func (m *SecurityServiceMock) Verify(userDetail auth.UserDetail, group []string) (bool, error) {
	args := m.Called(userDetail, group)
	return args.Get(0).(bool), args.Error(1)
}
