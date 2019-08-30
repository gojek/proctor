package middleware

import (
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"proctor/pkg/auth"
	"testing"

	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/security/service"
	"proctor/internal/pkg/constant"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	authenticationMiddleware authenticationMiddleware
	securityService          *service.SecurityServiceMock
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

	userDetail := &auth.UserDetail{
		Name: "William Dembo",
		Email: "email@gmail.com",
		Active:true,
		Groups: []string{"system", "proctor_maintainer"},
	}
	securityService := ctx.instance().securityService
	securityService.
		On("Auth", "email@gmail.com", "a-token").
		Return(userDetail, nil)

	authenticationMiddleware := ctx.instance().authenticationMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authenticationMiddleware.MiddlewareFunc(testHandler))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Add(constant.AccessTokenHeaderKey, "a-token")
	req.Header.Add(constant.UserEmailHeaderKey, "email@gmail.com")

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

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Add(constant.UserEmailHeaderKey, "email@gmail.com")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticationMiddleware_MiddlewareFuncWithoutEmail(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	authenticationMiddleware := ctx.instance().authenticationMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authenticationMiddleware.MiddlewareFunc(testHandler))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Add(constant.AccessTokenHeaderKey, "a-token")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticationMiddleware_MiddlewareFuncAuthFailed(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	var userDetail *auth.UserDetail
	securityService := ctx.instance().securityService
	securityService.
		On("Auth", "email@gmail.com", "a-token").
		Return(userDetail, errors.New("authentication failed, please check your access token"))

	authenticationMiddleware := ctx.instance().authenticationMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authenticationMiddleware.MiddlewareFunc(testHandler))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Add(constant.AccessTokenHeaderKey, "a-token")
	req.Header.Add(constant.UserEmailHeaderKey, "email@gmail.com")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
