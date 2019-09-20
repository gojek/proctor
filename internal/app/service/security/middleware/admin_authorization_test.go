package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/security/service"
	"proctor/pkg/auth"
)

type adminAuthorizationContext struct {
	authorizationMiddleware adminAuthorizationMiddleware
	requestHandler          func(http.Handler) http.Handler
	securityService         *service.SecurityServiceMock
}

func (context *adminAuthorizationContext) setUp(t *testing.T) {
	context.authorizationMiddleware = adminAuthorizationMiddleware{}
	context.requestHandler = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	context.securityService = &service.SecurityServiceMock{}
	context.authorizationMiddleware.service = context.securityService
	context.authorizationMiddleware.enabled = true
	context.authorizationMiddleware.requiredGroup = []string{"proctor_admin", "system"}
}

func (context *adminAuthorizationContext) tearDown() {
}

func (context *adminAuthorizationContext) instance() *adminAuthorizationContext {
	return context
}

func newAdminAuthorizationContext() *adminAuthorizationContext {
	return &adminAuthorizationContext{}
}

func TestAdminAuthorizationMiddleware_MiddlewareFuncExecutionSuccess(t *testing.T) {
	ctx := newAdminAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	requestBody := map[string]string{}
	requestBody["name"] = "a-job"
	body, _ := json.Marshal(requestBody)
	userDetail := &auth.UserDetail{
		Name:   "William Dembo",
		Email:  "email@gmail.com",
		Active: true,
		Groups: []string{"system", "proctor_admin"},
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	requestContext := context.WithValue(request.Context(), ContextUserDetailKey, userDetail)
	request = request.WithContext(requestContext)
	requestHandler := ctx.instance().requestHandler

	securityService := ctx.securityService
	securityService.On("Verify", *userDetail, ctx.authorizationMiddleware.requiredGroup).Return(true, nil)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authzMiddleware := ctx.instance().authorizationMiddleware
	requestHandler(authzMiddleware.MiddlewareFunc(testHandler)).ServeHTTP(response, request)

	responseResult := response.Result()
	assert.Equal(t, http.StatusOK, responseResult.StatusCode)
}

func TestAdminAuthorizationMiddleware_MiddlewareFuncScheduleSuccess(t *testing.T) {
	ctx := newAdminAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	requestBody := map[string]string{}
	requestBody["jobName"] = "a-job"
	body, _ := json.Marshal(requestBody)
	userDetail := &auth.UserDetail{
		Name:   "William Dembo",
		Email:  "email@gmail.com",
		Active: true,
		Groups: []string{"system", "proctor_maintainer"},
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	requestContext := context.WithValue(request.Context(), ContextUserDetailKey, userDetail)
	request = request.WithContext(requestContext)
	requestHandler := ctx.instance().requestHandler


	securityService := ctx.securityService
	securityService.On("Verify", *userDetail, ctx.authorizationMiddleware.requiredGroup).Return(true, nil)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authzMiddleware := ctx.instance().authorizationMiddleware
	requestHandler(authzMiddleware.MiddlewareFunc(testHandler)).ServeHTTP(response, request)

	responseResult := response.Result()
	assert.Equal(t, http.StatusOK, responseResult.StatusCode)
}

func TestAdminAuthorizationMiddleware_MiddlewareFuncWithoutUserDetail(t *testing.T) {
	ctx := newAdminAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	requestBody := map[string]string{}
	requestBody["name"] = "a-job"
	body, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	requestHandler := ctx.instance().requestHandler

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authzMiddleware := ctx.instance().authorizationMiddleware
	requestHandler(authzMiddleware.MiddlewareFunc(testHandler)).ServeHTTP(response, request)

	responseResult := response.Result()
	assert.Equal(t, http.StatusUnauthorized, responseResult.StatusCode)
}

func TestAdminAuthorizationMiddleware_MiddlewareFuncFailed(t *testing.T) {
	ctx := newAdminAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	requestBody := map[string]string{}
	requestBody["name"] = "a-job"
	body, _ := json.Marshal(requestBody)
	userDetail := &auth.UserDetail{
		Name:   "William Dembo",
		Email:  "email@gmail.com",
		Active: true,
		Groups: []string{"system", "not_proctor_maintainer"},
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	requestContext := context.WithValue(request.Context(), ContextUserDetailKey, userDetail)
	request = request.WithContext(requestContext)
	requestHandler := ctx.instance().requestHandler

	securityService := ctx.securityService
	securityService.On("Verify", *userDetail, ctx.authorizationMiddleware.requiredGroup).Return(false, nil)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authzMiddleware := ctx.instance().authorizationMiddleware
	requestHandler(authzMiddleware.MiddlewareFunc(testHandler)).ServeHTTP(response, request)

	responseResult := response.Result()
	assert.Equal(t, http.StatusForbidden, responseResult.StatusCode)
}

func TestAdminAuthorizationMiddleware_MiddlewareFuncDisabled(t *testing.T) {
	ctx := newAdminAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	authzMiddleware := ctx.instance().authorizationMiddleware
	authzMiddleware.enabled = false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(authzMiddleware.MiddlewareFunc(testHandler))
	defer ts.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminAuthorizationMiddleware_Secure(t *testing.T) {
	ctx := newAdminAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	router := mux.NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	authzMiddleware := ctx.instance().authorizationMiddleware
	securedRouter := authzMiddleware.Secure(router, "/secure/path", handler)

	handledPath, err := securedRouter.GetPathTemplate()
	assert.NoError(t, err)
	assert.Equal(t, "/secure/path", handledPath)
}
