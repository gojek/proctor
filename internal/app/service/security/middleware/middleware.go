package middleware

import "net/http"

const ContextUserDetailKey string = "USER_DETAIL"

type Middleware interface {
	MiddlewareFunc(http.Handler) http.Handler
}
