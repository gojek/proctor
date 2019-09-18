package message

import (
	"github.com/stretchr/testify/assert"
	"proctor/pkg/notification/event"
	"testing"
)

func TestExecutionMessage_JSON(t *testing.T) {
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
	evt.On("Type").Return(event.ExecutionEventType)
	evt.On("User").Return(userData)
	evt.On("Content").Return(content)
	defer evt.AssertExpectations(t)

	messageObject := NewExecutionMessage(&evt)
	result, err := messageObject.JSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestExecutionMessage_JSONMismatch(t *testing.T) {
	evt := event.EventMock{}
	evt.On("Type").Return(event.Type("not-execution-event"))
	defer evt.AssertExpectations(t)

	messageObject := NewExecutionMessage(&evt)
	result, err := messageObject.JSON()
	assert.Error(t, err)
	assert.Empty(t, result)
}
