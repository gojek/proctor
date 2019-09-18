package main

import (
	"proctor/pkg/notification"
	"proctor/pkg/notification/event"
	"proctor/plugins/slack-notification-plugin/slack"
	"testing"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	slackNotification notification.Observer
	slackClient       *slack.SlackClientMock
	event             *event.EventMock
}

func (context *testContext) setUp(t *testing.T) {
	context.slackClient = &slack.SlackClientMock{}
	context.slackNotification = NewSlackNotification(context.slackClient)
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

func TestSlackNotification_OnNotify(t *testing.T) {
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

	message := slack.NewStandardMessage("{\"Args\":\"args\",\"ExecutionID\":\"7\",\"ImageTag\":\"test\",\"JobName\":\"test-job\",\"Status\":\"CREATED\"}")
	ctx.instance().slackClient.On("Publish", message).Return(nil)

	ctx.instance().slackNotification.OnNotify(evt)

	ctx.instance().slackClient.AssertExpectations(t)
}
