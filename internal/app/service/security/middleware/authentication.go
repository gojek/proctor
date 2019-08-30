package middleware

import (
	"net/http"

	"proctor/internal/app/service/security/service"
	"proctor/internal/pkg/constant"
)

type authenticationMiddleware struct {
	service service.SecurityService
}

func (middleware *authenticationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(constant.AccessTokenHeaderKey)
		userEmail := r.Header.Get(constant.UserEmailHeaderKey)
		if token == "" || userEmail == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
