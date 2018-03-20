package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/gojektech/proctor/proctord/storage/postgres"
)

type Store interface {
	JobsExecutionAuditLog(string, string, string, string, map[string]string) error
}

type store struct {
	postgresClient postgres.Client
}

func New(postgresClient postgres.Client) Store {
	return &store{
		postgresClient: postgresClient,
	}
}

func (store *store) JobsExecutionAuditLog(jobSubmissionStatus, jobName, jobSubmittedForExecution, imageName string, jobArgs map[string]string) error {
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
	}
	return store.postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, image_name, job_submitted_for_execution, job_args, job_submission_status) VALUES (:job_name, :image_name, :job_submitted_for_execution, :job_args, :job_submission_status)", &jobsExecutionAuditLog)
}
