package middleware

import (
	"net/http"

	"proctor/internal/app/service/security/service"
)

type authenticationMiddleware struct {
	service service.SecurityService
}

func (middleware *authenticationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
