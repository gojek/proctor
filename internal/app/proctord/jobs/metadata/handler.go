package metadata

import (
	"encoding/json"
	"github.com/getsentry/raven-go"
	"net/http"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/pkg/constant"
	modelMetadata "proctor/internal/pkg/model/metadata"
)

type handler struct {
	store Store
}

type Handler interface {
	HandleSubmission() http.HandlerFunc
	HandleBulkDisplay() http.HandlerFunc
}

func NewHandler(store Store) Handler {
	return &handler{
		store: store,
	}
}

func (handler *handler) HandleSubmission() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var jobMetadata []modelMetadata.Metadata
		err := json.NewDecoder(req.Body).Decode(&jobMetadata)
		defer req.Body.Close()
		if err != nil {
			logger.Error("Error parsing request body", err.Error())

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.ClientError))
			return
		}

		for _, metadata := range jobMetadata {
			err = handler.store.CreateOrUpdateJobMetadata(metadata)
			if err != nil {
				logger.Error("Error updating metadata", err.Error())
				raven.CaptureError(err, nil)

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(constant.ServerError))
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (handler *handler) HandleBulkDisplay() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		jobMetadata, err := handler.store.GetAllJobsMetadata()
		if err != nil {
			logger.Error("Error fetching metadata", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(constant.ServerError))
			return
		}

		jobsMetadataInJSON, err := json.Marshal(jobMetadata)
		if err != nil {
			logger.Error("Error marshalling jobs metadata in json", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(constant.ServerError))
			return
		}

		w.Write(jobsMetadataInJSON)
	}
}
