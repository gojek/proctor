package execution

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockExecutioner struct {
	mock.Mock
}

func (m *MockExecutioner) Handle() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}
