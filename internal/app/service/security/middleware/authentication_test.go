package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/security/service"
	"proctor/internal/app/service/server/middleware/parameter"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	authenticationMiddleware authenticationMiddleware
	securityService          service.SecurityService
	testHandler              http.HandlerFunc
}

func (context *testContext) setUp(t *testing.T) {
	context.authenticationMiddleware = authenticationMiddleware{}
	context.securityService = &service.SecurityServiceMock{}
	context.authenticationMiddleware.service = context.securityService
	fn := func(rw http.ResponseWriter, req *http.Request) {
	}
	context.testHandler = http.HandlerFunc(fn)
}

func (context *testContext) tearDown() {
}

func (context *testContext) instance() *testContext {
	return context
}

func newContext() context {
	return &testContext{}
}

func TestAuthenticationMiddleware_MiddlewareFuncSuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	authenticationMiddleware := ctx.instance().authenticationMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authenticationMiddleware.MiddlewareFunc(testHandler))
	defer ts.Close()

	client := &http.Client{}

	body := map[string]string{}
	body["name"] = "a-job"
	requestBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewReader(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(parameter.AccessTokenHeader, "a-token")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthenticationMiddleware_MiddlewareFuncWithoutToken(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	authenticationMiddleware := ctx.instance().authenticationMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authenticationMiddleware.MiddlewareFunc(testHandler))
	defer ts.Close()

	client := &http.Client{}

	body := map[string]string{}
	body["name"] = "a-job"
	requestBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewReader(requestBody))
	req.Header.Add("Content-Type", "application/json")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
