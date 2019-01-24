package schedule

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"

	"github.com/badoux/checkmail"

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
	GetScheduledJobs() http.HandlerFunc
	GetScheduledJob() http.HandlerFunc
	RemoveScheduledJob() http.HandlerFunc
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

		if scheduledJob.Group == "" {
			logger.Error("Group Name is missing")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.GroupNameMissingError))
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

		scheduledJob.Time = fmt.Sprintf("0 %s", scheduledJob.Time) 
		scheduledJob.ID, err = scheduler.store.InsertScheduledJob(scheduledJob.Name, scheduledJob.Tags, scheduledJob.Time, scheduledJob.NotificationEmails, userEmail, scheduledJob.Group, scheduledJob.Args)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
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

func (scheduler *scheduler) GetScheduledJobs() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		scheduledJobsStoreFormat, err := scheduler.store.GetEnabledScheduledJobs()
		if err != nil {
			logger.Error("Error fetching scheduled jobs", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}

		if len(scheduledJobsStoreFormat) == 0 {
			logger.Error(utility.NoScheduledJobsError, nil)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		scheduledJobs := FromStoreToHandler(scheduledJobsStoreFormat)

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

func (scheduler *scheduler) GetScheduledJob() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		jobID := mux.Vars(req)["id"]
		scheduledJob, err := scheduler.store.GetScheduledJob(jobID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid input syntax") {
				logger.Error(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid Job ID"))
				return
			}
			logger.Error("Error fetching scheduled job", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}

		if len(scheduledJob) == 0 {
			logger.Error(utility.JobNotFoundError, nil)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(utility.JobNotFoundError))
			return
		}

		job := GetScheduledJob(scheduledJob[0])

		scheduledJobJson, err := json.Marshal(job)
		if err != nil {
			logger.Error("Error marshalling scheduled job", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}

		w.Write(scheduledJobJson)
	}
}

func (scheduler *scheduler) RemoveScheduledJob() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		jobID := mux.Vars(req)["id"]
		removedJobsCount, err := scheduler.store.RemoveScheduledJob(jobID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid input syntax") {
				logger.Error(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid Job ID"))
				return
			}
			logger.Error("Error fetching scheduled job", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}

		if removedJobsCount == 0 {
			logger.Error(utility.JobNotFoundError, nil)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(utility.JobNotFoundError))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Successfully unscheduled Job ID: %s", jobID)))
	}
}
