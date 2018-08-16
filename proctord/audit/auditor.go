package audit

import (
	"context"

	"github.com/gojektech/proctor/proctord/kubernetes"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/utility"
)

type Auditor interface {
	AuditJobsExecution(context.Context)
}

type auditor struct {
	store      storage.Store
	kubeClient kubernetes.Client
}

func New(store storage.Store, kubeClient kubernetes.Client) Auditor {
	return &auditor{
		store:      store,
		kubeClient: kubeClient,
	}
}

func (auditor *auditor) AuditJobsExecution(ctx context.Context) {
	jobSubmissionStatus := ctx.Value(utility.JobSubmissionStatusContextKey).(string)

	if jobSubmissionStatus != utility.JobSubmissionSuccess {
		err := auditor.store.JobsExecutionAuditLog(jobSubmissionStatus, utility.JobFailed, "", "", "", map[string]string{})
		if err != nil {
			logger.Error("Error auditing jobs execution", err)
		}
		return
	}
	jobName := ctx.Value(utility.JobNameContextKey).(string)
	JobNameSubmittedForExecution := ctx.Value(utility.JobNameSubmittedForExecutionContextKey).(string)
	imageName := ctx.Value(utility.ImageNameContextKey).(string)
	jobArgs := ctx.Value(utility.JobArgsContextKey).(map[string]string)

	err := auditor.store.JobsExecutionAuditLog(jobSubmissionStatus, utility.JobWaiting, jobName, JobNameSubmittedForExecution, imageName, jobArgs)
	if err != nil {
		logger.Error("Error auditing jobs execution", err)
	}

	go auditor.auditJobExecutionStatus(JobNameSubmittedForExecution)
}

func (auditor *auditor) auditJobExecutionStatus(JobNameSubmittedForExecution string) {
	status, err := auditor.kubeClient.JobExecutionStatus(JobNameSubmittedForExecution)
	if err != nil {
		logger.Error("Error getting job execution status", err)
	}

	err = auditor.store.UpdateJobsExecutionAuditLog(JobNameSubmittedForExecution, status)
	if err != nil {
		logger.Error("Error auditing jobs execution", err)
	}
}
