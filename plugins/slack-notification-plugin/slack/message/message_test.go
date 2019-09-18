package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStandardMessage_JSON(t *testing.T) {
	messageObject := NewStandardMessage("content")
	result, err := messageObject.JSON()
	assert.NoError(t, err)
	assert.Equal(t, "{\"text\":\"content\"}", result)
}
