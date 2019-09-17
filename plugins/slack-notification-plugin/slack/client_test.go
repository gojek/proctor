package slack

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type slackClientTest struct {
	client SlackClient
}

func (context *slackClientTest) setUp(t *testing.T) {
	httpClient := resty.New()
	httpmock.ActivateNonDefault(httpClient.GetClient())
	context.client = NewSlackClient(httpClient)
	assert.NotNil(t, context.client)
}

func (context *slackClientTest) tearDown() {
	httpmock.DeactivateAndReset()
}

func newNotificationServiceTestContext() *slackClientTest {
	return &slackClientTest{}
}

func TestSlackClient_Publish(t *testing.T) {
	ctx := newNotificationServiceTestContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	config := NewSlackConfig()
	message := MessageMock{}
	message.On("JSON").Return("message sent", nil)

	httpmock.RegisterResponder(
		"POST",
		config.url,
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, "message sent", string(body))

			contentType := req.Header.Get("Content-type")
			assert.Equal(t, "application/json", contentType)

			response := httpmock.NewStringResponse(200, "")
			return response, nil
		},
	)
	err := ctx.client.Publish(&message)
	assert.NoError(t, err)
	assert.Equal(t, 1, httpmock.GetTotalCallCount())
}

func TestSlackClient_PublishErrorJSON(t *testing.T) {
	ctx := newNotificationServiceTestContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	message := MessageMock{}
	message.On("JSON").Return("", errors.New("JSON unmarshal error")).Once()

	err := ctx.client.Publish(&message)
	assert.Error(t, err)
	assert.Equal(t, 0, httpmock.GetTotalCallCount())
}

func TestSlackClient_PublishErrorRequest(t *testing.T) {
	ctx := newNotificationServiceTestContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	config := NewSlackConfig()
	message := MessageMock{}
	message.On("JSON").Return("message sent", nil)

	httpmock.RegisterResponder(
		"POST",
		config.url,
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, "message sent", string(body))

			contentType := req.Header.Get("Content-type")
			assert.Equal(t, "application/json", contentType)

			response := httpmock.NewStringResponse(503, "")
			return response, errors.New("internal server error")
		},
	)
	err := ctx.client.Publish(&message)
	assert.Error(t, err)
	assert.Equal(t, 1, httpmock.GetTotalCallCount())
}
