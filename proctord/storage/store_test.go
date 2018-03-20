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

func TestJobsExeutionAuditLog(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	jobName := "any-job"
	imageName := "any-image"
	jobSubmittedForExecution := "any-submission"
	jobArgs := map[string]string{"key": "value"}
	jobSubmissionStatus := "any-status"

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
	}

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_execution_audit_log (job_name, image_name, job_submitted_for_execution, job_args, job_submission_status) VALUES (:job_name, :image_name, :job_submitted_for_execution, :job_args, :job_submission_status)",
		&data).
		Return(nil).
		Once()

	err = testStore.JobsExecutionAuditLog(jobSubmissionStatus, jobName, jobSubmittedForExecution, imageName, jobArgs)

	assert.NoError(t, err)
	mockPostgresClient.AssertExpectations(t)
}

func TestJobsExeutionAuditLogPostgresClientFailure(t *testing.T) {
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
		"INSERT INTO jobs_execution_audit_log (job_name, image_name, job_submitted_for_execution, job_args, job_submission_status) VALUES (:job_name, :image_name, :job_submitted_for_execution, :job_args, :job_submission_status)",
		&data).
		Return(errors.New("error")).
		Once()

	err = testStore.JobsExecutionAuditLog("", "", "", "", map[string]string{})

	assert.Error(t, err)
	mockPostgresClient.AssertExpectations(t)
}
