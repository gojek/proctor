package audit

import (
	"github.com/getsentry/raven-go"
	"proctor/proctord/kubernetes"
	"proctor/proctord/logger"
	"proctor/proctord/storage"
	"proctor/proctord/storage/postgres"
	"proctor/shared/constant"
)

type Auditor interface {
	JobsExecutionAndStatus(*postgres.JobsExecutionAuditLog)
	JobsExecution(*postgres.JobsExecutionAuditLog)
	JobsExecutionStatus(string) (string, error)
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

func (auditor *auditor) JobsExecutionAndStatus(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog) {
	auditor.JobsExecution(jobsExecutionAuditLog)

	if jobsExecutionAuditLog.JobSubmissionStatus == constant.JobSubmissionSuccess &&
		jobsExecutionAuditLog.ExecutionID.Valid {
		_, _ = auditor.JobsExecutionStatus(jobsExecutionAuditLog.ExecutionID.String)
	}
}

func (auditor *auditor) JobsExecution(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog) {
	err := auditor.store.AuditJobsExecution(jobsExecutionAuditLog)
	if err != nil {
		logger.Error("Error auditing jobs execution", err)
		raven.CaptureError(err, nil)
	}
}

func (auditor *auditor) JobsExecutionStatus(jobExecutionID string) (string, error) {
	status, err := auditor.kubeClient.JobExecutionStatus(jobExecutionID)
	if err != nil {
		logger.Error("Error getting job execution status", err)
		raven.CaptureError(err, nil)
		return "", err
	}

	err = auditor.store.UpdateJobsExecutionAuditLog(jobExecutionID, status)
	if err != nil {
		logger.Error("Error updating job status", err)
		raven.CaptureError(err, nil)
		return "", err
	}

	return status, err
}
