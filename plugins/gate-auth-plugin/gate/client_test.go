package gate

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
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

	email := "w.albertusd@gmail.com"
	token := "someunreadabletoken"
	config := NewGateConfig()

	mockGetUserProfileAPI(config, token, email)

	expectedUserDetail := &auth.UserDetail{
		Name:   "William Albertus Dembo",
		Email:  "w.albertusd@gmail.com",
		Active: true,
		Groups: []string{"system", "proctor_executor"},
	}

	actualUserDetail, err := ctx.instance().gateClient.GetUserProfile(email, token)

	assert.NoError(t, err)
	assert.NotNil(t, actualUserDetail)
	assert.Equal(t, expectedUserDetail, actualUserDetail)
	ctx.tearDown()
}

func TestGateClient_GetUserProfileUnauthenticated(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "w.albertusd@gmail.com"
	token := "someunreadabletoken"
	config := NewGateConfig()

	mockGetUserProfileAPI(config, token, email)

	userDetail, err := ctx.instance().gateClient.GetUserProfile(email, "wrong-token")

	assert.Nil(t, userDetail)
	assert.NotNil(t, err)
	assert.Equal(t, "authentication failed, please check your access token", err.Error())
	ctx.tearDown()
}

func TestGateClient_GetUserProfileNotFound(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "w.albertusd@gmail.com"
	token := "someunreadabletoken"
	config := NewGateConfig()

	mockGetUserProfileAPI(config, token, email)

	userDetail, err := ctx.instance().gateClient.GetUserProfile("random.email@gmail.com", token)

	assert.Nil(t, userDetail)
	assert.NotNil(t, err)
	assert.Equal(t, "user not found for email random.email@gmail.com", err.Error())
	ctx.tearDown()
}

func mockGetUserProfileAPI(config GateConfig, token string, email string) {
	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("%s://%s/%s", config.Protocol, config.Host, config.ProfilePath),
		func(req *http.Request) (*http.Response, error) {
			tokenParam := req.URL.Query()["access_token"][0]
			if tokenParam != token {
				return httpmock.NewStringResponse(401, ""), nil
			}
			emailParam := req.URL.Query()["email"][0]
			if emailParam != email {
				return httpmock.NewStringResponse(404, ""), nil
			}
			body := `{
						"email":"w.albertusd@gmail.com",
						"uid": "7",
						"name":"William Albertus Dembo",
						"active":true,
						"admin": true,
						"home_dir": null,
						"shell": null,
						"public_key": "ssh-rsa some-public-key",
						"user_login_id": "william.dembo",
						"product_name": null,
						"groups":[{"id":1,"name":"system"},{"id":2,"name":"proctor_executor"}]
					}`
			response := httpmock.NewStringResponse(200, body)
			response.Header.Set("Content-Type", "application/json")
			return response, nil
		},
	)
}

func TestGateClient_GetUserProfileServerFailure(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "w.albertusd@gmail.com"
	token := "someunreadabletoken"
	config := NewGateConfig()

	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("%s://%s/%s", config.Protocol, config.Host, config.ProfilePath),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(503, "Internal server error"), nil
		},
	)

	userDetail, err := ctx.instance().gateClient.GetUserProfile(email, token)

	assert.Nil(t, userDetail)
	assert.NotNil(t, err)
	ctx.tearDown()
}

func TestGateClient_GetUserProfileConnectionFailure(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "w.albertusd@gmail.com"
	token := "someunreadabletoken"
	config := NewGateConfig()

	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("%s://%s/%s", config.Protocol, config.Host, config.ProfilePath),
		httpmock.ConnectionFailure,
	)

	userDetail, err := ctx.instance().gateClient.GetUserProfile(email, token)

	assert.Nil(t, userDetail)
	assert.NotNil(t, err)
	ctx.tearDown()
}
