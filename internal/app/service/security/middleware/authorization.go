package middleware

import (
	"encoding/json"
	"net/http"
	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/security/service"
)

type authorizationMiddleware struct {
	service service.SecurityService
}

func (middleware *authorizationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var job parameter.Job
		_ = json.NewDecoder(r.Body).Decode(&job)
		if job.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
