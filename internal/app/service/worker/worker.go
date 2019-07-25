package worker

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/robfig/cron"

	executionContextRepository "proctor/internal/app/service/execution/repository"
	executionService "proctor/internal/app/service/execution/service"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/app/service/infra/kubernetes"
	"proctor/internal/app/service/infra/kubernetes/http"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/infra/mail"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	scheduleModel "proctor/internal/app/service/schedule/model"
	scheduleRepository "proctor/internal/app/service/schedule/repository"
	secretRepository "proctor/internal/app/service/secret/repository"
)

const WorkerEmail = "worker@proctor"

type worker struct {
	executionService           executionService.ExecutionService
	executionContextRepository executionContextRepository.ExecutionContextRepository
	scheduleRepository         scheduleRepository.ScheduleRepository
	mailer                     mail.Mailer
	inMemorySchedules          map[uint64]*cron.Cron
}

type Worker interface {
	Run(<-chan time.Time, <-chan os.Signal)
}

func NewWorker(executionSvc executionService.ExecutionService, executionContextRepo executionContextRepository.ExecutionContextRepository, scheduleRepo scheduleRepository.ScheduleRepository, mailer mail.Mailer) Worker {
	return &worker{
		executionService:           executionSvc,
		executionContextRepository: executionContextRepo,
		scheduleRepository:         scheduleRepo,
		mailer:                     mailer,
		inMemorySchedules:          make(map[uint64]*cron.Cron),
	}
}

func (worker *worker) disableScheduleIfItExists(scheduleID uint64) {
	if scheduledCronJob, ok := worker.inMemorySchedules[scheduleID]; ok {
		scheduledCronJob.Stop()
		delete(worker.inMemorySchedules, scheduleID)
	}
}

func (worker *worker) enableScheduleIfItDoesNotExist(schedule scheduleModel.Schedule) {
	if _, ok := worker.inMemorySchedules[schedule.ID]; !ok {
		cronJob := cron.New()
		err := cronJob.AddFunc(schedule.Cron, func() {
			executionContext, _, err := worker.executionService.Execute(schedule.JobName, WorkerEmail, schedule.Args)
			if err != nil {
				logger.Error(fmt.Sprintf("Error submitting job: %s ", schedule.Tags), schedule.JobName, " for execution: ", err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName})

				return
			}

			err = worker.mailer.Send(*executionContext, schedule)

			if err != nil {
				logger.Error(fmt.Sprintf("Error notifying job: %s `", schedule.Tags), schedule.JobName, "` ID: `", executionContext.ExecutionID, "` execution status: `", executionContext.Status, "` to users: ", err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags, "job_name": schedule.JobName, "job_id": fmt.Sprint(executionContext.ExecutionID), "job_execution_status": string(executionContext.Status)})
				return
			}
		})

		if err != nil {
			logger.Error(fmt.Sprintf("Error adding cron job: %s", schedule.Tags), err.Error())
			raven.CaptureError(err, map[string]string{"job_tags": schedule.Tags})
			return
		}

		cronJob.Start()
		worker.inMemorySchedules[schedule.ID] = cronJob
	}
}

func (worker *worker) Run(tickerChan <-chan time.Time, signalsChan <-chan os.Signal) {
	for {
		select {
		case <-tickerChan:
			schedules, err := worker.scheduleRepository.GetAll()
			if err != nil {
				logger.Error("Error getting scheduled jobs from store: ", err.Error())
				raven.CaptureError(err, nil)
				continue
			}

			for _, schedule := range schedules {
				if schedule.Enabled {
					worker.enableScheduleIfItDoesNotExist(schedule)
				} else {
					worker.disableScheduleIfItExists(schedule.ID)
				}
			}
		case <-signalsChan:
			for id := range worker.inMemorySchedules {
				worker.disableScheduleIfItExists(id)
			}
			//TODO: wait for all active executions to complete
			return
		}
	}
}

func Start() error {
	fmt.Println("started scheduler")

	postgresClient := postgresql.NewClient()
	redisClient := redis.NewClient()

	executionContextStore := executionContextRepository.NewExecutionContextRepository(postgresClient)
	metadataStore := metadataRepository.NewMetadataRepository(redisClient)
	secretStore := secretRepository.NewSecretRepository(redisClient)
	scheduleStore := scheduleRepository.NewScheduleRepository(postgresClient)

	httpClient, err := http.NewClient()
	if err != nil {
		return err
	}
	kubeClient := kubernetes.NewKubernetesClient(httpClient)
	mailer := mail.New(config.MailServerHost(), config.MailServerPort())
	executionSvc := executionService.NewExecutionService(kubeClient, executionContextStore, metadataStore, secretStore)
	worker := NewWorker(executionSvc, executionContextStore, scheduleStore, mailer)
	ticker := time.NewTicker(time.Duration(config.ScheduledJobsFetchIntervalInMins()) * time.Minute)
	signalsChan := make(chan os.Signal, 1)

	worker.Run(ticker.C, signalsChan)

	_ = postgresClient.Close()
	return nil
}
