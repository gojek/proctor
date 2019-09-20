package main

import (
	"errors"
	"proctor/plugins/slack-notification-plugin/slack/message"
	"testing"

	"github.com/stretchr/testify/assert"

	"proctor/pkg/notification/event"
	"proctor/plugins/slack-notification-plugin/slack"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	slackNotification *slackNotification
	slackClient       *slack.SlackClientMock
	event             *event.EventMock
}

func (context *testContext) setUp(t *testing.T) {
	context.slackClient = &slack.SlackClientMock{}
	context.slackNotification = &slackNotification{}
	context.slackNotification.slackClient = context.slackClient
	context.event = &event.EventMock{}
}

func (context *testContext) tearDown() {
}

func (context *testContext) instance() *testContext {
	return context
}

func newContext() context {
	return &testContext{}
}

func TestSlackNotification_OnNotifyExecution(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

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
	evt := ctx.instance().event
	evt.On("Type").Return(event.ExecutionEventType)
	evt.On("User").Return(userData)
	evt.On("Content").Return(content)

	messageObject := message.NewExecutionMessage(evt)
	ctx.instance().slackClient.On("Publish", messageObject).Return(nil)

	err := ctx.instance().slackNotification.OnNotify(evt)
	assert.NoError(t, err)

	ctx.instance().slackClient.AssertExpectations(t)
}

func TestSlackNotification_OnNotifyExecutionErrorPublish(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

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
	evt := ctx.instance().event
	evt.On("Type").Return(event.ExecutionEventType)
	evt.On("User").Return(userData)
	evt.On("Content").Return(content)

	messageObject := message.NewExecutionMessage(evt)
	ctx.instance().slackClient.On("Publish", messageObject).Return(errors.New("publish error"))

	err := ctx.instance().slackNotification.OnNotify(evt)
	assert.Error(t, err)

	ctx.instance().slackClient.AssertExpectations(t)
}

func TestSlackNotification_OnNotifyUnsupportedEvent(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

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
	evt := ctx.instance().event
	evt.On("Type").Return(event.Type("Unsupported event"))
	evt.On("User").Return(userData)
	evt.On("Content").Return(content)

	messageObject := message.NewStandardMessage(evt)
	ctx.instance().slackClient.On("Publish", messageObject).Return(nil)

	err := ctx.instance().slackNotification.OnNotify(evt)
	assert.NoError(t, err)

	ctx.instance().slackClient.AssertExpectations(t)
}
