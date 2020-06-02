package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	httpClient, err := NewClient()
	assert.NotNil(t, httpClient)
	assert.NoError(t, err)
}
