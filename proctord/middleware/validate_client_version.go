package middleware

import (
	"fmt"
	"github.com/gojektech/proctor/proctord/config"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/hashicorp/go-version"
	"github.com/gojektech/proctor/proctord/logger"
	"net/http"
)

func ValidateClientVersion(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		requestHeaderClientVersion := r.Header.Get(utility.ClientVersionHeaderKey)

		if requestHeaderClientVersion != "" {
			clientVersion, err := version.NewVersion(requestHeaderClientVersion)
			if err != nil {
				logger.Error("Error while creating requestHeaderClientVersion", err.Error())
			}

			minClientVersion, err := version.NewVersion(config.MinClientVersion())
			if err != nil {
				logger.Error("Error while creating minClientVersion", err.Error())
			}

			if clientVersion.LessThan(minClientVersion) {
				w.WriteHeader(400)
				w.Write([]byte(fmt.Sprintf(utility.ClientOutdatedErrorMessage, clientVersion)))
				return
			}
			next.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	}
}
