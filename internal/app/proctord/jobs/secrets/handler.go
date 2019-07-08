package secrets

import (
	"encoding/json"
	"github.com/getsentry/raven-go"
	"net/http"
	 "proctor/internal/app/service/infra/logger"

	"proctor/internal/pkg/constant"
)

type handler struct {
	secretsStore Store
}

type Handler interface {
	HandleSubmission() http.HandlerFunc
}

func NewHandler(secretsStore Store) Handler {
	return &handler{
		secretsStore: secretsStore,
	}
}

func (handler *handler) HandleSubmission() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var secret Secret
		err := json.NewDecoder(req.Body).Decode(&secret)
		defer req.Body.Close()
		if err != nil {
			logger.Error("Error parsing request body", err.Error())

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.ClientError))
			return
		}

		err = handler.secretsStore.CreateOrUpdateJobSecret(secret)
		if err != nil {
			logger.Error("Error updating secrets", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
