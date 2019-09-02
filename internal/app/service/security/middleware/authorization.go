package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/metadata/repository"
	"proctor/internal/app/service/schedule/model"
	"proctor/internal/app/service/security/service"
	"proctor/pkg/auth"
)

type authorizationMiddleware struct {
	service            service.SecurityService
	metadataRepository repository.MetadataRepository
}

func (middleware *authorizationMiddleware) Secure(router *mux.Router, path string, handler http.Handler) *mux.Route {
	return router.NewRoute().Path(path).Handler(middleware.MiddlewareFunc(handler))
}

func (middleware *authorizationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobName, err := extractName(r)
		logger.LogErrors(err, "decode json", r.Body)
		if jobName == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		jobMetadata, err := middleware.metadataRepository.GetByName(jobName)
		logger.LogErrors(err, "get metadata", jobName)
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

func extractName(r *http.Request) (string, error) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	var job parameter.Job
	err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&job)
	if err != nil {
		return "", err
	}
	if job.Name != "" {
		return job.Name, nil
	}

	var schedule model.Schedule
	err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&schedule)
	if err != nil {
		return "", err
	}
	if schedule.JobName != "" {
		return schedule.JobName, nil
	}

	return "", nil
}

func NewAuthorizationMiddleware(securityService service.SecurityService, metadataRepository repository.MetadataRepository) AuthorizationMiddleware {
	return &authorizationMiddleware{
		service:            securityService,
		metadataRepository: metadataRepository,
	}
}
