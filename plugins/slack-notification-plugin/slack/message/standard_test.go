package message

import (
	"github.com/stretchr/testify/assert"
	"proctor/pkg/notification/event"
	"testing"
)

func TestStandardMessage_JSON(t *testing.T) {
	userData := event.UserData{
		Email: "proctor@example.com",
	}
	content := map[string]string{
		"ExecutionID": "7",
		"JobName":     "test-job",
		"ImageTag":    "test",
		"Args":        "args",
		"Status":      "CREATED",
	}
	evt := event.EventMock{}
	evt.On("Type").Return(event.Type("unsupported-event"))
	evt.On("User").Return(userData)
	evt.On("Content").Return(content)
	defer evt.AssertExpectations(t)

	messageObject := NewStandardMessage(&evt)
	result, err := messageObject.JSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}
