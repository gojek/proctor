package gate

import (
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type integrationContext interface {
	setUp(t *testing.T)
	tearDown()
	instance() *integrationTestContext
}

type integrationTestContext struct {
	gateClient GateClient
	email      string
	token      string
}

func (context *integrationTestContext) setUp(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_PLUGIN_INTEGRATION_TEST")
	if available != true || value != "true" {
		t.SkipNow()
	}
	client := resty.New()
	context.gateClient = NewGateClient(client)
	email, _ := os.LookupEnv("GATE_PLUGIN_EMAIL")
	assert.NotEmpty(t, email)
	context.email = email
	token, _ := os.LookupEnv("GATE_PLUGIN_TOKEN")
	assert.NotEmpty(t, token)
	context.token = token
}

func (context *integrationTestContext) tearDown() {
}

func (context *integrationTestContext) instance() *integrationTestContext {
	return context
}

func newIntegrationContext() integrationContext {
	return &integrationTestContext{}
}

func TestIntegrationGateClient_GetUserProfileSuccess(t *testing.T) {
	ctx := newIntegrationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	email := ctx.instance().email
	token := ctx.instance().token
	userDetail, err := ctx.instance().gateClient.GetUserProfile(email, token)

	assert.NoError(t, err)
	assert.NotNil(t, userDetail)
	assert.NotEmpty(t, userDetail.Email)
	assert.NotEmpty(t, userDetail.Name)
	assert.NotEmpty(t, userDetail.Active)
	assert.NotNil(t, userDetail.Groups)
}

func TestIntegrationGateClient_GetUserProfileUnauthenticated(t *testing.T) {
	ctx := newIntegrationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	email := ctx.instance().email
	userDetail, err := ctx.instance().gateClient.GetUserProfile(email, "some-random-token")

	assert.Nil(t, userDetail)
	assert.Error(t, err)
	assert.Equal(t, "authentication failed, please check your access token", err.Error())
}
