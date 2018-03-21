package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/gojektech/proctor/proctord/storage/postgres"
)

type Store interface {
	JobsExecutionAuditLog(string, string, string, string, string, map[string]string) error
	UpdateJobsExecutionAuditLog(string, string) error
}

type store struct {
	postgresClient postgres.Client
}

func New(postgresClient postgres.Client) Store {
	return &store{
		postgresClient: postgresClient,
	}
}

func (store *store) JobsExecutionAuditLog(jobSubmissionStatus, jobExecutionStatus, jobName, jobSubmittedForExecution, imageName string, jobArgs map[string]string) error {
	var encodedJobArgs bytes.Buffer
	enc := gob.NewEncoder(&encodedJobArgs)
	err := enc.Encode(jobArgs)
	if err != nil {
		return err
	}

	jobsExecutionAuditLog := postgres.JobsExecutionAuditLog{
		JobName:                  jobName,
		ImageName:                imageName,
		JobSubmittedForExecution: jobSubmittedForExecution,
		JobArgs:                  base64.StdEncoding.EncodeToString(encodedJobArgs.Bytes()),
		JobSubmissionStatus:      jobSubmissionStatus,
		JobExecutionStatus:       jobExecutionStatus,
	}
	return store.postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, image_name, job_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :image_name, :job_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)", &jobsExecutionAuditLog)
}

func (store *store) UpdateJobsExecutionAuditLog(jobSubmittedForExecution, status string) error {
	jobsExecutionAuditLog := postgres.JobsExecutionAuditLog{
		JobExecutionStatus:       status,
		JobSubmittedForExecution: jobSubmittedForExecution,
	}

	return store.postgresClient.NamedExec("UPDATE jobs_execution_audit_log SET job_execution_status = :job_execution_status where job_submitted_for_execution = :job_submitted_for_execution", &jobsExecutionAuditLog)
}
