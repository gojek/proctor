package middleware

import (
	"context"
	"net/http"

	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/security/service"
	"proctor/internal/pkg/constant"
)

type authenticationMiddleware struct {
	service service.SecurityService
	enabled bool
}

func (middleware *authenticationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !middleware.enabled {
			next.ServeHTTP(w, r)
			return
		}
		token := r.Header.Get(constant.AccessTokenHeaderKey)
		userEmail := r.Header.Get(constant.UserEmailHeaderKey)
		if token == "" || userEmail == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userDetail, err := middleware.service.Auth(userEmail, token)
		logger.LogErrors(err, "authentication user", userEmail)
		if userDetail == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ContextUserDetailKey, userDetail)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewAuthenticationMiddleware(securityService service.SecurityService) Middleware {
	proctorConfig := config.Load()
	return &authenticationMiddleware{
		service: securityService,
		enabled: proctorConfig.AuthEnabled,
	}
}
