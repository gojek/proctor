package schedule

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gojektech/proctor/proctord/audit"
	"github.com/gojektech/proctor/proctord/jobs/execution"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/mail"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/robfig/cron"
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
			logger.Error("Error deserializing job args: ", err.Error())
			return
		}

		cronJob := cron.New()
		err = cronJob.AddFunc(scheduledJob.Time, func() {
			jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{}
			jobsExecutionAuditLog.UserEmail = utility.WorkerEmail

			jobExecutionID, err := worker.executioner.Execute(jobsExecutionAuditLog, scheduledJob.Name, jobArgs)
			if err != nil {
				logger.Error("Error submitting job: ", scheduledJob.Name, " for execution: ", err.Error())

				jobsExecutionAuditLog.Errors = fmt.Sprintf("Error executing job: %s", err.Error())
				jobsExecutionAuditLog.JobSubmissionStatus = utility.JobSubmissionServerError
				worker.auditor.JobsExecution(jobsExecutionAuditLog)
				return
			}

			worker.auditor.JobsExecution(jobsExecutionAuditLog)

			jobExecutionStatus, err := worker.auditor.JobsExecutionStatus(jobExecutionID)
			if err != nil {
				logger.Error("Error fetching execution status for job: ", jobExecutionID, ". Error: ", err.Error())
				return
			}

			recipients := strings.Split(scheduledJob.NotificationEmails, ",")
			err = worker.mailer.Send(scheduledJob.Name, jobExecutionID, jobExecutionStatus, jobArgs, recipients)

			if err != nil {
				logger.Error("Error notifying job: `", scheduledJob.Name, "` ID: `", jobExecutionID, "` execution status: `", jobExecutionStatus, "` to users: ", err.Error())
				return
			}
		})

		if err != nil {
			logger.Error("Error adding cron job: ", err.Error())
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
