package middleware

import (
	"github.com/gorilla/mux"
	"net/http"
)

const ContextUserDetailKey string = "USER_DETAIL"

type Middleware interface {
	MiddlewareFunc(http.Handler) http.Handler
	Exclude(...*mux.Route)
}

type AuthorizationMiddleware interface {
	Secure(router *mux.Router, path string, handler http.Handler) *mux.Route
}
