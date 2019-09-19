package main

import (
	"github.com/go-resty/resty/v2"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"proctor/pkg/notification/event"
	"proctor/plugins/slack-notification-plugin/slack"
)

type integrationContext interface {
	setUp(t *testing.T)
	tearDown()
	instance() *integrationTestContext
}

type integrationTestContext struct {
	slackNotification *slackNotification
	event             *event.EventMock
}

func (context *integrationTestContext) setUp(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_PLUGIN_INTEGRATION_TEST")
	if available != true || value != "true" {
		t.SkipNow()
	}
	slackUrl, _ := os.LookupEnv("SLACK_PLUGIN_URL")
	assert.NotEmpty(t, slackUrl)

	context.slackNotification = &slackNotification{}
	context.slackNotification.slackClient = slack.NewSlackClient(resty.New())
	context.event = &event.EventMock{}
}

func (context *integrationTestContext) tearDown() {
}

func (context *integrationTestContext) instance() *integrationTestContext {
	return context
}

func newIntegrationContext() integrationContext {
	return &integrationTestContext{}
}

func TestSlackNotificationIntegration_OnNotifyExecution(t *testing.T) {
	ctx := newIntegrationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	userData := event.UserData{
		Email: "slack_notification_integration_execution_event@example.com",
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

	err := ctx.instance().slackNotification.OnNotify(evt)
	assert.NoError(t, err)
}

func TestSlackNotificationIntegration_OnNotifyUnsupportedEvent(t *testing.T) {
	ctx := newIntegrationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	userData := event.UserData{
		Email: "slack_notification_integration_unsupported_event@example.com",
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

	err := ctx.instance().slackNotification.OnNotify(evt)
	assert.NoError(t, err)
}
