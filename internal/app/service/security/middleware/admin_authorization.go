package middleware

import (
	"net/http"
	"proctor/internal/app/service/security/service"

	"github.com/gorilla/mux"

	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"proctor/pkg/auth"
)

type adminAuthorizationMiddleware struct {
	service       service.SecurityService
	requiredGroup []string
	enabled       bool
}

func (middleware *adminAuthorizationMiddleware) Secure(router *mux.Router, path string, handler http.Handler) *mux.Route {
	return router.NewRoute().Path(path).Handler(middleware.MiddlewareFunc(handler))
}

func (middleware *adminAuthorizationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !middleware.enabled {
			next.ServeHTTP(w, r)
			return
		}

		userDetail := r.Context().Value(ContextUserDetailKey)
		if userDetail == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authorized, err := middleware.service.Verify(*(userDetail.(*auth.UserDetail)), middleware.requiredGroup)
		logger.LogErrors(err, "authorization", middleware.requiredGroup, userDetail)
		if !authorized {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewAdminAuthorizationMiddleware(securityService service.SecurityService) AuthorizationMiddleware {
	proctorConfig := config.Config()
	return &adminAuthorizationMiddleware{
		service:       securityService,
		requiredGroup: proctorConfig.AuthRequiredAdminGroup,
		enabled:       proctorConfig.AuthEnabled,
	}
}
