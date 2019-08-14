package auth

import (
	"github.com/stretchr/testify/mock"
)

type AuthMock struct {
	mock.Mock
}

func (m *AuthMock) Auth(email string, token string) (*UserDetail, error) {
	args := m.Called(email, token)
	return args.Get(0).(*UserDetail), args.Error(1)
}

func (m *AuthMock) Verify(userDetail UserDetail, group []string) (bool, error) {
	args := m.Called(userDetail, group)
	return args.Get(0).(bool), args.Error(1)
}
