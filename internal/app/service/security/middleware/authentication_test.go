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

type testContext struct {
	authMiddleware  authenticationMiddleware
	securityService *service.SecurityServiceMock
	testHandler     http.HandlerFunc
}

func (context *testContext) setUp(t *testing.T) {
	context.authMiddleware = authenticationMiddleware{}
	context.securityService = &service.SecurityServiceMock{}
	context.authMiddleware.service = context.securityService
	fn := func(w http.ResponseWriter, r *http.Request) {
	}
	context.testHandler = fn
}

func (context *testContext) tearDown() {
}

func (context *testContext) instance() *testContext {
	return context
}

func newContext() *testContext {
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

	authMiddleware := ctx.instance().authMiddleware
	fn := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, userDetail, r.Context().Value("USER_DETAIL"))
	}
	testHandler := http.HandlerFunc(fn)
	ts := httptest.NewServer(authMiddleware.MiddlewareFunc(testHandler))
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

	authMiddleware := ctx.instance().authMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authMiddleware.MiddlewareFunc(testHandler))
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

	authMiddleware := ctx.instance().authMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authMiddleware.MiddlewareFunc(testHandler))
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

	authMiddleware := ctx.instance().authMiddleware
	testHandler := ctx.instance().testHandler
	ts := httptest.NewServer(authMiddleware.MiddlewareFunc(testHandler))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Add(constant.AccessTokenHeaderKey, "a-token")
	req.Header.Add(constant.UserEmailHeaderKey, "email@gmail.com")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
