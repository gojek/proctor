package schedule

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron"
	"proctor/proctord/audit"
	"proctor/proctord/jobs/execution"
	"proctor/proctord/logger"
	"proctor/proctord/mail"
	"proctor/proctord/storage"
	"proctor/proctord/storage/postgres"
	"proctor/shared/constant"
	"proctor/shared/utility"
)

type worker struct {
	store                 storage.Store
	executioner           execution.Executioner
	auditor               audit.Auditor
	mailer                mail.Mailer
	inMemoryScheduledJobs map[string]*cron.Cron
}

type Worker interface {
	Run(<-chan time.Time, <-chan os.Signal)
}

func NewWorker(store storage.Store, executioner execution.Executioner, auditor audit.Auditor, mailer mail.Mailer) Worker {
	return &worker{
		store:                 store,
		executioner:           executioner,
		auditor:               auditor,
		mailer:                mailer,
		inMemoryScheduledJobs: make(map[string]*cron.Cron),
	}
}

func (worker *worker) disableScheduledJobIfItExists(scheduledJobID string) {
	if scheduledCronJob, ok := worker.inMemoryScheduledJobs[scheduledJobID]; ok {
		scheduledCronJob.Stop()
		delete(worker.inMemoryScheduledJobs, scheduledJobID)
	}
}

func (worker *worker) enableScheduledJobIfItDoesNotExist(scheduledJob postgres.JobsSchedule) {
	if _, ok := worker.inMemoryScheduledJobs[scheduledJob.ID]; !ok {
		jobArgs, err := utility.DeserializeMap(scheduledJob.Args)
		if err != nil {
			logger.Error(fmt.Sprintf("Error deserializing job args: %s ", scheduledJob.Tags), scheduledJob.Name, err.Error())
			raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name})
			return
		}

		cronJob := cron.New()
		err = cronJob.AddFunc(scheduledJob.Time, func() {
			jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{}
			jobsExecutionAuditLog.UserEmail = constant.WorkerEmail

			jobExecutionID, err := worker.executioner.Execute(jobsExecutionAuditLog, scheduledJob.Name, jobArgs)
			if err != nil {
				logger.Error(fmt.Sprintf("Error submitting job: %s ", scheduledJob.Tags), scheduledJob.Name, " for execution: ", err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name})

				jobsExecutionAuditLog.Errors = fmt.Sprintf("Error executing job: %s", err.Error())
				jobsExecutionAuditLog.JobSubmissionStatus = constant.JobSubmissionServerError
				worker.auditor.JobsExecution(jobsExecutionAuditLog)
				return
			}

			worker.auditor.JobsExecution(jobsExecutionAuditLog)

			jobExecutionStatus, err := worker.auditor.JobsExecutionStatus(jobExecutionID)
			if err != nil {
				logger.Error(fmt.Sprintf("Error fetching execution status for job: %s ", scheduledJob.Tags), jobExecutionID, ". Error: ", err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name})

				return
			}

			recipients := strings.Split(scheduledJob.NotificationEmails, ",")
			err = worker.mailer.Send(scheduledJob.Name, jobExecutionID, jobExecutionStatus, jobArgs, recipients)

			if err != nil {
				logger.Error(fmt.Sprintf("Error notifying job: %s `", scheduledJob.Tags), scheduledJob.Name, "` ID: `", jobExecutionID, "` execution status: `", jobExecutionStatus, "` to users: ", err.Error())
				raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags, "job_name": scheduledJob.Name, "job_id": jobExecutionID, "job_execution_status": jobExecutionStatus})
				return
			}
		})

		if err != nil {
			logger.Error(fmt.Sprintf("Error adding cron job: %s", scheduledJob.Tags), err.Error())
			raven.CaptureError(err, map[string]string{"job_tags": scheduledJob.Tags})
			return
		}

		cronJob.Start()
		worker.inMemoryScheduledJobs[scheduledJob.ID] = cronJob
	}
}

func (worker *worker) Run(tickerChan <-chan time.Time, signalsChan <-chan os.Signal) {
	for {
		select {
		case <-tickerChan:
			scheduledJobs, err := worker.store.GetScheduledJobs()
			if err != nil {
				logger.Error("Error getting scheduled jobs from store: ", err.Error())
				raven.CaptureError(err, nil)
				continue
			}

			for _, scheduledJob := range scheduledJobs {
				if scheduledJob.Enabled {
					worker.enableScheduledJobIfItDoesNotExist(scheduledJob)
				} else {
					worker.disableScheduledJobIfItExists(scheduledJob.ID)
				}
			}
		case <-signalsChan:
			for id, _ := range worker.inMemoryScheduledJobs {
				worker.disableScheduledJobIfItExists(id)
			}
			//TODO: wait for all active executions to complete
			return
		}
	}
}
