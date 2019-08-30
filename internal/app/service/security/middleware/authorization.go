package middleware

import (
	"encoding/json"
	"net/http"

	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/metadata/repository"
	"proctor/internal/app/service/security/service"
	"proctor/pkg/auth"
)

type authorizationMiddleware struct {
	service            service.SecurityService
	metadataRepository repository.MetadataRepository
}

func (middleware *authorizationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var job parameter.Job
		err := json.NewDecoder(r.Body).Decode(&job)
		logger.LogErrors(err, "decode json", r.Body)
		if job.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		jobMetadata, err := middleware.metadataRepository.GetByName(job.Name)
		logger.LogErrors(err, "get metadata", job.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userDetail := r.Context().Value(ContextUserDetailKey)
		if userDetail == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authorized, err := middleware.service.Verify(*(userDetail.(*auth.UserDetail)), jobMetadata.AuthorizedGroups)
		logger.LogErrors(err, "authorization", jobMetadata, userDetail)
		if !authorized {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
