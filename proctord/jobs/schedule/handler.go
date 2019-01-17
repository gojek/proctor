package schedule

import (
	"encoding/json"
	"github.com/badoux/checkmail"
	"net/http"
	"strings"

	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/robfig/cron"
)

type scheduler struct {
	store         storage.Store
	metadataStore metadata.Store
}

type Scheduler interface {
	Schedule() http.HandlerFunc
}

func NewScheduler(store storage.Store, metadataStore metadata.Store) Scheduler {
	return &scheduler{
		metadataStore: metadataStore,
		store:         store,
	}
}

func (scheduler *scheduler) Schedule() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var scheduledJob ScheduledJob
		err := json.NewDecoder(req.Body).Decode(&scheduledJob)
		userEmail := req.Header.Get(utility.UserEmailHeaderKey)
		defer req.Body.Close()
		if err != nil {
			logger.Error("Error parsing request body for scheduling jobs: ", err.Error())

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.ClientError))

			return
		}

		_, err = cron.Parse(scheduledJob.Time)
		if err != nil {
			logger.Error("Client provided invalid cron expression: ", scheduledJob.Time)

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.InvalidCronExpressionClientError))
			return
		}

		notificationEmails := strings.Split(scheduledJob.NotificationEmails, ",")

		for _, notificationEmail := range notificationEmails {
			err = checkmail.ValidateFormat(notificationEmail)
			if err != nil {
				logger.Error("Client provided invalid email address: ", notificationEmail)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(utility.InvalidEmailIdClientError))
				return
			}
		}

		if scheduledJob.Tags == "" {
			logger.Error("Tag(s) are missing")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.InvalidTagError))
			return
		}

		_, err = scheduler.metadataStore.GetJobMetadata(scheduledJob.Name)
		if err != nil {
			if err.Error() == "redigo: nil returned" {
				logger.Error("Client provided non existent proc name: ", scheduledJob.Name)

				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(utility.NonExistentProcClientError))
			} else {
				logger.Error("Error fetching metadata for proc", err.Error())

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(utility.ServerError))
			}

			return
		}

		scheduledJob.ID, err = scheduler.store.InsertScheduledJob(scheduledJob.Name, scheduledJob.Tags, scheduledJob.Time, scheduledJob.NotificationEmails, userEmail, scheduledJob.Args)
		if err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"unique_jobs_schedule_name_args\"" {
				logger.Error("Client provided duplicate combination of scheduled job name and args: ", scheduledJob.Name, scheduledJob.Args)

				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(utility.DuplicateJobNameArgsClientError))

				return
			} else {
				logger.Error("Error persisting scheduled job", err.Error())

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(utility.ServerError))

				return
			}
		}

		responseBody, err := json.Marshal(scheduledJob)
		if err != nil {
			logger.Error("Error marshaling response body", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))

			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(responseBody)
		return
	}
}
