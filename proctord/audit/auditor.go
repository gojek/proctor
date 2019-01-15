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
	AuditJobExecutionStatus(string) (string, error)
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
	userEmail := ctx.Value(utility.UserEmailContextKey).(string)

	if jobSubmissionStatus != utility.JobSubmissionSuccess {
		err := auditor.store.JobsExecutionAuditLog(jobSubmissionStatus, utility.JobFailed, "", userEmail, "", "", map[string]string{})
		if err != nil {
			logger.Error("Error auditing jobs execution", err)
		}
		return
	}
	jobName := ctx.Value(utility.JobNameContextKey).(string)
	JobNameSubmittedForExecution := ctx.Value(utility.JobNameSubmittedForExecutionContextKey).(string)
	imageName := ctx.Value(utility.ImageNameContextKey).(string)
	jobArgs := ctx.Value(utility.JobArgsContextKey).(map[string]string)

	err := auditor.store.JobsExecutionAuditLog(jobSubmissionStatus, utility.JobWaiting, jobName, userEmail, JobNameSubmittedForExecution, imageName, jobArgs)
	if err != nil {
		logger.Error("Error auditing jobs execution", err)
	}
}

func (auditor *auditor) AuditJobExecutionStatus(jobExecutionID string) (string, error) {
	status, err := auditor.kubeClient.JobExecutionStatus(jobExecutionID)
	if err != nil {
		logger.Error("Error getting job execution status", err)
		return "", err
	}

	err = auditor.store.UpdateJobsExecutionAuditLog(jobExecutionID, status)
	return status, err
}
