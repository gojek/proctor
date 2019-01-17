package audit

import (
	"github.com/gojektech/proctor/proctord/kubernetes"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"
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

	if jobsExecutionAuditLog.JobSubmissionStatus == utility.JobSubmissionSuccess &&
		jobsExecutionAuditLog.ExecutionID.Valid {
		auditor.JobsExecutionStatus(jobsExecutionAuditLog.ExecutionID.String)
	}
}

func (auditor *auditor) JobsExecution(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog) {
	err := auditor.store.AuditJobsExecution(jobsExecutionAuditLog)
	if err != nil {
		logger.Error("Error auditing jobs execution", err)
	}
}

func (auditor *auditor) JobsExecutionStatus(jobExecutionID string) (string, error) {
	status, err := auditor.kubeClient.JobExecutionStatus(jobExecutionID)
	if err != nil {
		logger.Error("Error getting job execution status", err)
		return "", err
	}

	err = auditor.store.UpdateJobsExecutionAuditLog(jobExecutionID, status)
	return status, err
}
