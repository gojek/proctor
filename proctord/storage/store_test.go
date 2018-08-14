package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"testing"

	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/stretchr/testify/assert"
)

func TestJobsExecutionAuditLog(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	jobName := "any-job"
	imageName := "any-image"
	jobSubmittedForExecution := "any-submission"
	jobArgs := map[string]string{"key": "value"}
	jobSubmissionStatus := "any-status"
	jobExecutionStatus := "any-execution-status"

	var encodedJobArgs bytes.Buffer
	enc := gob.NewEncoder(&encodedJobArgs)
	err := enc.Encode(jobArgs)
	assert.NoError(t, err)

	data := postgres.JobsExecutionAuditLog{
		JobName:                  jobName,
		ImageName:                imageName,
		JobSubmittedForExecution: jobSubmittedForExecution,
		JobArgs:                  base64.StdEncoding.EncodeToString(encodedJobArgs.Bytes()),
		JobSubmissionStatus:      jobSubmissionStatus,
		JobExecutionStatus:       jobExecutionStatus,
	}

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_execution_audit_log (job_name, image_name, job_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :image_name, :job_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)",
		&data).
		Return(nil).
		Once()

	err = testStore.JobsExecutionAuditLog(jobSubmissionStatus, jobExecutionStatus, jobName, jobSubmittedForExecution, imageName, jobArgs)

	assert.NoError(t, err)
	mockPostgresClient.AssertExpectations(t)
}

func TestJobsExecutionAuditLogPostgresClientFailure(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	var encodedJobArgs bytes.Buffer
	enc := gob.NewEncoder(&encodedJobArgs)
	err := enc.Encode(map[string]string{})
	assert.NoError(t, err)

	data := postgres.JobsExecutionAuditLog{
		JobArgs: base64.StdEncoding.EncodeToString(encodedJobArgs.Bytes()),
	}

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_execution_audit_log (job_name, image_name, job_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :image_name, :job_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)",
		&data).
		Return(errors.New("error")).
		Once()

	err = testStore.JobsExecutionAuditLog("", "", "", "", "", map[string]string{})

	assert.Error(t, err)
	mockPostgresClient.AssertExpectations(t)
}

func TestJobsStatusGetWhenJobIsNotPresent(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)
	jobName := "any-job"

	dest := []postgres.JobsExecutionAuditLog{}

	mockPostgresClient.On("Select",
		&dest,
		"SELECT job_execution_status from jobs_execution_audit_log where job_name = $1",
		jobName).
		Return(nil).
		Once()

	testStore.GetJobStatus(jobName)
	mockPostgresClient.AssertExpectations(t)
}
