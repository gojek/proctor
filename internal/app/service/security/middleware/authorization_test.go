package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/metadata/repository"
	"proctor/internal/pkg/model/metadata"
)

type authorizationContext struct {
	authorizationMiddleware authorizationMiddleware
	requestHandler          func(http.Handler) http.Handler
	metadataRepository      *repository.MockMetadataRepository
}

func (context *authorizationContext) setUp(t *testing.T) {
	context.authorizationMiddleware = authorizationMiddleware{}
	context.requestHandler = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	context.metadataRepository = &repository.MockMetadataRepository{}
	context.authorizationMiddleware.metadataRepository = context.metadataRepository
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

	requestBody := map[string]string{}
	requestBody["name"] = "a-job"
	body, _ := json.Marshal(requestBody)
	jobMetadata := &metadata.Metadata{
		Name:             "a-job",
		Description:      "jobMetadata of a job",
		ImageName:        "ubuntu-18.04",
		AuthorizedGroups: []string{"system", "proctor_maintainer"},
		Author:           "systeam team",
		Contributors:     "proctor team",
		Organization:     "GoJek",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	requestHandler := ctx.instance().requestHandler
	metadataRepository := ctx.metadataRepository

	metadataRepository.On("GetByName", "a-job").Return(jobMetadata, nil)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authorizationMiddleware := ctx.instance().authorizationMiddleware
	requestHandler(authorizationMiddleware.MiddlewareFunc(testHandler)).ServeHTTP(response, request)

	responseResult := response.Result()
	assert.Equal(t, http.StatusOK, responseResult.StatusCode)
}

func TestAuthorizationMiddleware_MiddlewareFuncWithoutName(t *testing.T) {
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
	assert.Equal(t, http.StatusBadRequest, responseResult.StatusCode)
}

func TestAuthorizationMiddleware_MiddlewareFuncMetadataError(t *testing.T) {
	ctx := newAuthorizationContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	var jobMetadata *metadata.Metadata
	requestBody := map[string]string{}
	requestBody["name"] = "a-job"
	body, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	requestHandler := ctx.instance().requestHandler

	metadataRepository := ctx.metadataRepository
	err := errors.New("metadata not found")
	metadataRepository.On("GetByName", "a-job").Return(jobMetadata, err)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authorizationMiddleware := ctx.instance().authorizationMiddleware
	requestHandler(authorizationMiddleware.MiddlewareFunc(testHandler)).ServeHTTP(response, request)

	responseResult := response.Result()
	assert.Equal(t, http.StatusInternalServerError, responseResult.StatusCode)
}
