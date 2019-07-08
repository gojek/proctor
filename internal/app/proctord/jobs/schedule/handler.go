package schedule

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	"net/http"
	jobMetadata "proctor/internal/app/proctord/jobs/metadata"
	"proctor/internal/app/proctord/storage"
	"proctor/internal/app/service/infra/logger"
	"strings"

	"github.com/badoux/checkmail"

	"github.com/robfig/cron"
	"proctor/internal/pkg/constant"
	modelSchedule "proctor/internal/pkg/model/schedule"
)

type scheduler struct {
	store         storage.Store
	metadataStore jobMetadata.Store
}

type Scheduler interface {
	Schedule() http.HandlerFunc
	GetScheduledJobs() http.HandlerFunc
	GetScheduledJob() http.HandlerFunc
	RemoveScheduledJob() http.HandlerFunc
}

func NewScheduler(store storage.Store, metadataStore jobMetadata.Store) Scheduler {
	return &scheduler{
		metadataStore: metadataStore,
		store:         store,
	}
}

func (scheduler *scheduler) Schedule() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var scheduledJob modelSchedule.ScheduledJob
		err := json.NewDecoder(req.Body).Decode(&scheduledJob)
		userEmail := req.Header.Get(constant.UserEmailHeaderKey)
		defer req.Body.Close()
		if err != nil {
			logger.Error("Error parsing request body for scheduling jobs: ", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.ClientError))

			return
		}

		if scheduledJob.Tags == "" {
			logger.Error("Tag(s) are missing")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.InvalidTagError))
			return
		}

		_, err = cron.Parse(scheduledJob.Time)
		if err != nil {
			logger.Error(fmt.Sprintf("Client provided invalid cron expression: %s ", scheduledJob.Tags), scheduledJob.Name, scheduledJob.Time)

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.InvalidCronExpressionClientError))
			return
		}

		notificationEmails := strings.Split(scheduledJob.NotificationEmails, ",")

		for _, notificationEmail := range notificationEmails {
			err = checkmail.ValidateFormat(notificationEmail)
			if err != nil {
				logger.Error(fmt.Sprintf("Client provided invalid email address: %s: ", scheduledJob.Tags), scheduledJob.Name, notificationEmail)
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(constant.InvalidEmailIdClientError))
				return
			}
		}

		if scheduledJob.Group == "" {
			logger.Error(fmt.Sprintf("Group Name is missing %s: ", scheduledJob.Tags), scheduledJob.Name)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.GroupNameMissingError))
			return
		}

		_, err = scheduler.metadataStore.GetJobMetadata(scheduledJob.Name)
		if err != nil {
			if err.Error() == "redigo: nil returned" {
				logger.Error(fmt.Sprintf("Client provided non existent proc name: %s ", scheduledJob.Tags), scheduledJob.Name)

				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(constant.NonExistentProcClientError))
			} else {
				logger.Error(fmt.Sprintf("Error fetching metadata for proc %s ", scheduledJob.Tags), scheduledJob.Name, err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name})

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(constant.ServerError))
			}

			return
		}

		scheduledJob.Time = fmt.Sprintf("0 %s", scheduledJob.Time)
		scheduledJob.ID, err = scheduler.store.InsertScheduledJob(scheduledJob.Name, scheduledJob.Tags, scheduledJob.Time, scheduledJob.NotificationEmails, userEmail, scheduledJob.Group, scheduledJob.Args)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				logger.Error(fmt.Sprintf("Client provided duplicate combination of scheduled job name and args: %s ", scheduledJob.Tags), scheduledJob.Name, scheduledJob.Args)
				raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name})

				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(constant.DuplicateJobNameArgsClientError))

				return
			} else {
				logger.Error(fmt.Sprintf("Error persisting scheduled job %s ", scheduledJob.Tags), scheduledJob.Name, err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name})

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(constant.ServerError))

				return
			}
		}

		responseBody, err := json.Marshal(scheduledJob)
		if err != nil {
			logger.Error(fmt.Sprintf("Error marshaling response body %s ", scheduledJob.Tags), scheduledJob.Name, err.Error())
			raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name})

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))

			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(responseBody)
		return
	}
}

func (scheduler *scheduler) GetScheduledJobs() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		scheduledJobsStoreFormat, err := scheduler.store.GetEnabledScheduledJobs()
		if err != nil {
			logger.Error("Error fetching scheduled jobs", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		if len(scheduledJobsStoreFormat) == 0 {
			logger.Error(constant.NoScheduledJobsError, nil)

			w.WriteHeader(http.StatusNoContent)
			return
		}

		scheduledJobs, err := FromStoreToHandler(scheduledJobsStoreFormat)
		if err != nil {
			logger.Error("Error deserializing scheduled job args to map: ", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		scheduledJobsJson, err := json.Marshal(scheduledJobs)
		if err != nil {
			logger.Error("Error marshalling scheduled jobs", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		_, _ = w.Write(scheduledJobsJson)
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
				_, _ = w.Write([]byte("Invalid Job ID"))
				return
			}
			logger.Error("Error fetching scheduled job", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		if len(scheduledJob) == 0 {
			logger.Error(constant.JobNotFoundError, nil)

			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(constant.JobNotFoundError))
			return
		}

		job, err := GetScheduledJob(scheduledJob[0])
		if err != nil {
			logger.Error("Error deserializing scheduled job args to map: ", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		scheduledJobJson, err := json.Marshal(job)
		if err != nil {
			logger.Error("Error marshalling scheduled job", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		_, _ = w.Write(scheduledJobJson)
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
				_, _ = w.Write([]byte("Invalid Job ID"))
				return
			}
			logger.Error("Error fetching scheduled job", err.Error())
			raven.CaptureError(err, nil)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			return
		}

		if removedJobsCount == 0 {
			logger.Error(constant.JobNotFoundError, nil)

			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(constant.JobNotFoundError))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("Successfully unscheduled Job ID: %s", jobID)))
	}
}
