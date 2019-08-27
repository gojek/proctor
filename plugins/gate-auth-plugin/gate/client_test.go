package gate

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/thingful/httpmock"
	"net/http"
	"proctor/pkg/auth"
	"testing"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	gateClient GateClient
}

func (context *testContext) setUp(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	context.gateClient = NewGateClient(client)
	assert.NotNil(t, context.gateClient)
}

func (context *testContext) tearDown() {
	httpmock.DeactivateAndReset()
}

func (context *testContext) instance() *testContext {
	return context
}

func newContext() context {
	return &testContext{}
}

func TestGateClient_GetUserProfileSuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	config := NewGateConfig()
	body := `{"email":"w.albertusd@gmail.com","name":"William Albertus Dembo","active":true,"groups":[{"id":1,"name":"system"},{"id":2,"name":"proctor_executor"}]}`

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("%s://%s/%s", config.Protocol, config.Host, config.ProfilePath),
			func(req *http.Request) (*http.Response, error) {
				response := httpmock.NewStringResponse(200, body)
				response.Header.Set("Content-Type", "application/json")
				return response, nil
			},
		),
	)

	expectedUserDetail := &auth.UserDetail{
		Name:   "William Albertus Dembo",
		Email:  "w.albertusd@gmail.com",
		Active: true,
		Groups: []string{"system", "proctor_executor"},
	}

	email := "w.albertusd@gmail.com"
	token := "someunreadabletoken"
	actualUserDetail, err := ctx.instance().gateClient.GetUserProfile(email, token)

	assert.NoError(t, err)
	assert.NotNil(t, actualUserDetail)
	assert.Equal(t, expectedUserDetail, actualUserDetail)
	assert.NoError(t, httpmock.AllStubsCalled())
	ctx.tearDown()
}
