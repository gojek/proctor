package slack

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStandardMessage_JSON(t *testing.T) {
	message := NewStandardMessage("content")
	result, err := message.JSON()
	assert.NoError(t, err)
	assert.Equal(t, "{\"text\":\"content\"}", result)
}
