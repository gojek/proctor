package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/badoux/checkmail"
	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"

	"proctor/internal/app/service/infra/logger"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	modelSchedule "proctor/internal/app/service/schedule/model"
	scheduleRepository "proctor/internal/app/service/schedule/repository"
	"proctor/internal/pkg/status"
)

type scheduler struct {
	repository         scheduleRepository.ScheduleRepository
	metadataRepository metadataRepository.MetadataRepository
}

type Scheduler interface {
	Schedule() http.HandlerFunc
	GetScheduledJobs() http.HandlerFunc
	GetScheduledJob() http.HandlerFunc
	RemoveScheduledJob() http.HandlerFunc
}

func NewScheduler(repository scheduleRepository.ScheduleRepository, metadataRepository metadataRepository.MetadataRepository) Scheduler {
	return &scheduler{
		metadataRepository: metadataRepository,
		repository:         repository,
	}
}

func (scheduler *scheduler) Schedule() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		var schedule modelSchedule.Schedule
		err := json.NewDecoder(request.Body).Decode(&schedule)
		defer request.Body.Close()
		if err != nil {
			logger.Error("Error parsing request body for schedule: ", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.MalformedRequestError))

			return
		}

		if schedule.Tags == "" {
			logger.Error("Tag(s) are missing")
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.ScheduleTagMissingError))
			return
		}

		_, err = cron.Parse(schedule.Cron)
		if err != nil {
			logger.Error(fmt.Sprintf("Cron format is invalid: %s ", schedule.Tags), schedule.JobName, schedule.Cron)

			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.ScheduleCronFormatInvalidError))
			return
		}

		notificationEmails := strings.Split(schedule.NotificationEmails, ",")

		for _, notificationEmail := range notificationEmails {
			err = checkmail.ValidateFormat(notificationEmail)
			if err != nil {
				logger.Error(fmt.Sprintf("Email address provided is invalid: %s: ", schedule.Tags), schedule.JobName, notificationEmail)
				response.WriteHeader(http.StatusBadRequest)
				_, _ = response.Write([]byte(status.EmailInvalidError))
				return
			}
		}

		if schedule.Group == "" {
			logger.Error(fmt.Sprintf("Group is missing %s: ", schedule.Tags), schedule.JobName)
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.ScheduleGroupMissingError))
			return
		}

		_, err = scheduler.metadataRepository.GetByName(schedule.JobName)
		if err != nil {
			if err.Error() == "redigo: nil returned" {
				logger.Error(fmt.Sprintf("Metadata not found: %s ", schedule.Tags), schedule.JobName)

				response.WriteHeader(http.StatusNotFound)
				_, _ = response.Write([]byte(status.MetadataNotFoundError))
			} else {
				logger.Error(fmt.Sprintf("Error fetching metadata for proc %s ", schedule.Tags), schedule.JobName, err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

				response.WriteHeader(http.StatusInternalServerError)
				_, _ = response.Write([]byte(status.GenericServerError))
			}

			return
		}

		schedule.Cron = fmt.Sprintf("0 %s", schedule.Cron)
		schedule.ID, err = scheduler.repository.Insert(&schedule)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				logger.Error(fmt.Sprintf("Duplicate combination of scheduled job name and args: %s ", schedule.Tags), schedule.JobName, schedule.Args)
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

				response.WriteHeader(http.StatusConflict)
				_, _ = response.Write([]byte(status.ScheduleDuplicateJobNameArgsError))

				return
			} else {
				logger.Error(fmt.Sprintf("Error persisting scheduled job %s ", schedule.Tags), schedule.JobName, err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

				response.WriteHeader(http.StatusInternalServerError)
				_, _ = response.Write([]byte(status.GenericServerError))

				return
			}
		}

		responseBody, err := json.Marshal(schedule)
		if err != nil {
			logger.Error(fmt.Sprintf("Error marshaling response body %s ", schedule.Tags), schedule.JobName, err.Error())
			raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(status.GenericServerError))

			return
		}

		response.WriteHeader(http.StatusCreated)
		_, _ = response.Write(responseBody)
		return
	}
}

func (scheduler *scheduler) GetScheduledJobs() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		schedules, err := scheduler.repository.GetAllEnabled()
		if err != nil {
			logger.Error("Error fetching scheduled jobs", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(status.GenericServerError))
			return
		}

		if len(scheduleList) == 0 {
			logger.Error(status.ScheduleListNotFoundError, nil)

			response.WriteHeader(http.StatusNoContent)
			return
		}

		schedulesJson, err := json.Marshal(schedules)
		if err != nil {
			logger.Error("Error marshalling schedule list", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(status.GenericServerError))
			return
		}

		_, _ = response.Write(schedulesJson)
	}
}

func (scheduler *scheduler) GetScheduledJob() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		jobId := mux.Vars(request)["id"]
		scheduleId, err := strconv.ParseUint(jobId, 10, 64)
		logger.LogErrors(err, "parse execution context id from path parameter:", jobId)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.PathParameterError))
			return
		}

		schedule, err := scheduler.repository.GetById(scheduleId)
		if err != nil {
			if strings.Contains(err.Error(), "invalid input syntax") {
				logger.Error(err.Error())
				response.WriteHeader(http.StatusBadRequest)
				_, _ = response.Write([]byte(status.ScheduleIdInvalidError))
				return
			}
			logger.Error("Error fetching scheduled job", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(status.GenericServerError))
			return
		}

		scheduleJson, err := json.Marshal(schedule)
		if err != nil {
			logger.Error("Error marshalling scheduled job", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(status.GenericServerError))
			return
		}

		_, _ = response.Write(scheduleJson)
	}
}

func (scheduler *scheduler) RemoveScheduledJob() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		jobId := mux.Vars(request)["id"]
		scheduleId, err := strconv.ParseUint(jobId, 10, 64)
		logger.LogErrors(err, "parse execution context id from path parameter:", jobId)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.PathParameterError))
			return
		}

		err = scheduler.repository.Delete(scheduleId)
		if err != nil {
			if strings.Contains(err.Error(), "invalid input syntax") {
				logger.Error(err.Error())

				response.WriteHeader(http.StatusBadRequest)
				_, _ = response.Write([]byte(status.ScheduleIdInvalidError))
				return
			}
			logger.Error("Error fetching schedule", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(status.GenericServerError))
			return
		}

		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(status.ScheduleDeleteSuccess))
	}
}
