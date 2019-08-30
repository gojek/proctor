package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type authorizationContext struct {
	authorizationMiddleware authorizationMiddleware
	requestHandler          func(http.Handler) http.Handler
}

func (context *authorizationContext) setUp(t *testing.T) {
	context.authorizationMiddleware = authorizationMiddleware{}
	context.requestHandler = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func (context *authorizationContext) tearDown() {
}

func (context *authorizationContext) instance() *authorizationContext {
	return context
}

func newAuthorizationContext() *authorizationContext {
	return &authorizationContext{}
}

func TestAuthorizationMiddleware_MiddlewareFuncSuccess(t *testing.T) {
	ctx := newAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	requestHandler := ctx.instance().requestHandler

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authorizationMiddleware := ctx.instance().authorizationMiddleware
	requestHandler(authorizationMiddleware.MiddlewareFunc(testHandler)).ServeHTTP(response, request)

	responseResult := response.Result()
	assert.Equal(t, http.StatusOK, responseResult.StatusCode)
}
