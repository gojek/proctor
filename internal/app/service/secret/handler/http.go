package handler

import (
	"encoding/json"
	"net/http"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/secret/model"
	"proctor/internal/app/service/secret/repository"

	"proctor/internal/pkg/constant"
)

type handler struct {
	repository repository.SecretRepository
}

type SecretHTTPHandler interface {
	Post() http.HandlerFunc
}

func NewSecretHTTPHandler(repository repository.SecretRepository) SecretHTTPHandler {
	return &handler{
		repository: repository,
	}
}

func (handler *handler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var secret model.Secret
		err := json.NewDecoder(req.Body).Decode(&secret)
		defer req.Body.Close()
		if err != nil {
			logger.Error("parsing json body to secret, failed", err.Error())

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.ClientError))
			return
		}

		err = handler.repository.Save(secret)
		if err != nil {
			logger.Error("saving secret to storage, failed", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
