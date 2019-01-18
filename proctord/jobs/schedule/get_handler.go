package schedule

import (
	"encoding/json"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/utility"
	"net/http"
)

type handler struct {
	store storage.Store
}


type Handler interface {
	GetScheduledJobs() http.HandlerFunc
}

func NewGetHandler(store storage.Store) Handler {
	return &handler{
		store: store,
	}
}

func (handler *handler) GetScheduledJobs() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		scheduledJobs, err := handler.store.GetScheduledJobs()
		if err != nil {
			logger.Error("Error fetching scheduled jobs", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}
		scheduledJobsJson, err := json.Marshal(scheduledJobs)
		if err != nil {
			logger.Error("Error marshalling scheduled jobs", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}

		w.Write(scheduledJobsJson)
	}
}