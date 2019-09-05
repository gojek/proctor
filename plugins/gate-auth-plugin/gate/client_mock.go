package gate

import (
	"github.com/stretchr/testify/mock"
	"proctor/pkg/auth"
)

type GateClientMock struct {
	mock.Mock
}

func (g *GateClientMock) GetUserProfile(email string, token string) (*auth.UserDetail, error) {
	args := g.Called(email, token)
	return args.Get(0).(*auth.UserDetail), args.Error(1)
}
