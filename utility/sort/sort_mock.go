package sort

import (
	"github.com/stretchr/testify/mock"
	"github.com/gojektech/proctor/proctord/jobs/metadata"
)

type MockSorter struct {
	mock.Mock
}

func (m *MockSorter) Sort(md []metadata.Metadata)  {
	m.Called(md)
}
