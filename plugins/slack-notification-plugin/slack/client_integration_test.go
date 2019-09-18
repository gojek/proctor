package slack

import (
	"os"
	"testing"

	"proctor/plugins/slack-notification-plugin/slack/message"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type integrationContext interface {
	setUp(t *testing.T)
	tearDown()
	instance() *integrationTestContext
}

type integrationTestContext struct {
	slackClient SlackClient
	slackUrl    string
}

func (context *integrationTestContext) setUp(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_PLUGIN_INTEGRATION_TEST")
	if available != true || value != "true" {
		t.SkipNow()
	}
	client := resty.New()
	context.slackClient = NewSlackClient(client)
	slackUrl, _ := os.LookupEnv("SLACK_PLUGIN_URL")
	assert.NotEmpty(t, slackUrl)
	context.slackUrl = slackUrl
}

func (context *integrationTestContext) tearDown() {
}

func (context *integrationTestContext) instance() *integrationTestContext {
	return context
}

func newIntegrationContext() integrationContext {
	return &integrationTestContext{}
}

func TestSlackClientIntegration_Publish(t *testing.T) {
	ctx := newIntegrationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	messageObject := message.NewStandardMessage("Message from slack plugin integration test with standard message")
	err := ctx.instance().slackClient.Publish(messageObject)
	assert.NoError(t, err)
}
