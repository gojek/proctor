package middleware

import "net/http"

type Middleware interface {
	MiddlewareFunc(http.Handler) http.Handler
}
