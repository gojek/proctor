package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/security/service"
	"proctor/internal/pkg/constant"
)

type authenticationMiddleware struct {
	service        service.SecurityService
	enabled        bool
	excludedRoutes []*mux.Route
}

func (middleware *authenticationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !middleware.enabled {
			next.ServeHTTP(w, r)
			return
		}
		if middleware.isRequestExcluded(r) {
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

func (middleware *authenticationMiddleware) Exclude(routes ...*mux.Route) {
	for _, route := range routes {
		middleware.excludedRoutes = append(middleware.excludedRoutes, route)
	}
}

func (middleware *authenticationMiddleware) isRequestExcluded(r *http.Request) bool {
	for _, route := range middleware.excludedRoutes {
		match := mux.RouteMatch{}
		if route.Match(r, &match) {
			return true
		}
	}
	return false
}

func NewAuthenticationMiddleware(securityService service.SecurityService) Middleware {
	proctorConfig := config.Load()
	return &authenticationMiddleware{
		service: securityService,
		enabled: proctorConfig.AuthEnabled,
	}
}
