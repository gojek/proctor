package middleware

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"net/http"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/server/middleware/parameter"
	"proctor/internal/app/service/server/middleware/status"
)

func ValidateClientVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestHeaderClientVersion := r.Header.Get(parameter.ClientVersionHeader)

		if requestHeaderClientVersion != "" {
			clientVersion, err := version.NewVersion(requestHeaderClientVersion)
			if err != nil {
				logger.Error("Error while creating requestHeaderClientVersion", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(fmt.Sprintf(status.ClientVersionInvalidMessage, "Minimum Client Version Config is Missing")))
				return
			}

			minClientVersion, err := version.NewVersion(config.MinClientVersion())
			if err != nil {
				logger.Error("Error while creating minClientVersion ", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(fmt.Sprintf(status.ServerConfigErrorMessage, "Minimum Client Version Config is Missing")))
				return
			}

			if clientVersion.LessThan(minClientVersion) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(fmt.Sprintf(status.ClientOutdatedErrorMessage, clientVersion)))
				return
			}

			next.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
