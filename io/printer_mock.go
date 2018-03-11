package io

import (
	"github.com/fatih/color"
	"github.com/stretchr/testify/mock"
)

type MockPrinter struct {
	mock.Mock
}

func (m *MockPrinter) Println(s string, attr ...color.Attribute) {
	argsCalled := make([]interface{}, 1+len(attr))
	argsCalled[0] = s
	for i, v := range attr {
		argsCalled[i+1] = v
	}

	m.Called(argsCalled...)
}
