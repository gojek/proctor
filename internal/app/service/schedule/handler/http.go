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
	"proctor/internal/app/service/schedule/handler/status"
	modelSchedule "proctor/internal/app/service/schedule/model"
	scheduleRepository "proctor/internal/app/service/schedule/repository"
	"proctor/internal/pkg/constant"
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
			logger.Error("Error parsing request body for scheduling jobs: ", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(constant.ClientError))

			return
		}

		if schedule.Tags == "" {
			logger.Error("Tag(s) are missing")
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(constant.InvalidTagError))
			return
		}

		_, err = cron.Parse(schedule.Cron)
		if err != nil {
			logger.Error(fmt.Sprintf("Client provided invalid cron expression: %s ", schedule.Tags), schedule.JobName, schedule.Cron)

			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(constant.InvalidCronExpressionClientError))
			return
		}

		notificationEmails := strings.Split(schedule.NotificationEmails, ",")

		for _, notificationEmail := range notificationEmails {
			err = checkmail.ValidateFormat(notificationEmail)
			if err != nil {
				logger.Error(fmt.Sprintf("Client provided invalid email address: %s: ", schedule.Tags), schedule.JobName, notificationEmail)
				response.WriteHeader(http.StatusBadRequest)
				_, _ = response.Write([]byte(constant.InvalidEmailIdClientError))
				return
			}
		}

		if schedule.Group == "" {
			logger.Error(fmt.Sprintf("Group Name is missing %s: ", schedule.Tags), schedule.JobName)
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(constant.GroupNameMissingError))
			return
		}

		_, err = scheduler.metadataRepository.GetByName(schedule.JobName)
		if err != nil {
			if err.Error() == "redigo: nil returned" {
				logger.Error(fmt.Sprintf("Client provided non existent proc name: %s ", schedule.Tags), schedule.JobName)

				response.WriteHeader(http.StatusNotFound)
				_, _ = response.Write([]byte(constant.NonExistentProcClientError))
			} else {
				logger.Error(fmt.Sprintf("Error fetching metadata for proc %s ", schedule.Tags), schedule.JobName, err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

				response.WriteHeader(http.StatusInternalServerError)
				_, _ = response.Write([]byte(constant.ServerError))
			}

			return
		}

		schedule.Cron = fmt.Sprintf("0 %s", schedule.Cron)
		schedule.ID, err = scheduler.repository.Insert(&schedule)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				logger.Error(fmt.Sprintf("Client provided duplicate combination of scheduled job name and args: %s ", schedule.Tags), schedule.JobName, schedule.Args)
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

				response.WriteHeader(http.StatusConflict)
				_, _ = response.Write([]byte(constant.DuplicateJobNameArgsClientError))

				return
			} else {
				logger.Error(fmt.Sprintf("Error persisting scheduled job %s ", schedule.Tags), schedule.JobName, err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

				response.WriteHeader(http.StatusInternalServerError)
				_, _ = response.Write([]byte(constant.ServerError))

				return
			}
		}

		responseBody, err := json.Marshal(schedule)
		if err != nil {
			logger.Error(fmt.Sprintf("Error marshaling response body %s ", schedule.Tags), schedule.JobName, err.Error())
			raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(constant.ServerError))

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
			_, _ = response.Write([]byte(constant.ServerError))
			return
		}

		if len(schedules) == 0 {
			logger.Error(constant.NoScheduledJobsError, nil)

			response.WriteHeader(http.StatusNoContent)
			return
		}

		schedulesJson, err := json.Marshal(schedules)
		if err != nil {
			logger.Error("Error marshalling scheduled jobs", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(constant.ServerError))
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
				_, _ = response.Write([]byte("Invalid Job ID"))
				return
			}
			logger.Error("Error fetching scheduled job", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(constant.ServerError))
			return
		}

		scheduleJson, err := json.Marshal(schedule)
		if err != nil {
			logger.Error("Error marshalling scheduled job", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(constant.ServerError))
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
				_, _ = response.Write([]byte("Invalid Job ID"))
				return
			}
			logger.Error("Error fetching scheduled job", err.Error())
			raven.CaptureError(err, nil)

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(constant.ServerError))
			return
		}

		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(fmt.Sprintf("Successfully unscheduled Job ID: %d", scheduleId)))
	}
}
