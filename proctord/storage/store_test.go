package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"testing"

	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJobsExecutionAuditLog(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	jobName := "any-job"
	imageName := "any-image"
	userEmail := "mrproctor@example.com"
	jobNameSubmittedForExecution := "any-submission"
	jobArgs := map[string]string{"key": "value"}
	jobSubmissionStatus := "any-status"
	jobExecutionStatus := "any-execution-status"

	var encodedJobArgs bytes.Buffer
	enc := gob.NewEncoder(&encodedJobArgs)
	err := enc.Encode(jobArgs)
	assert.NoError(t, err)

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_execution_audit_log (job_name, user_email, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status, created_at, updated_at) VALUES (:job_name, :user_email, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status, :created_at, :updated_at)",
		mock.Anything).
		Run(func(args mock.Arguments) {
			data := args.Get(1).(*postgres.JobsExecutionAuditLog)

			assert.Equal(t, jobName, data.JobName)
			assert.Equal(t, userEmail, data.UserEmail)
			assert.Equal(t, imageName, data.ImageName)
			assert.Equal(t, postgres.StringToSQLString(jobNameSubmittedForExecution), data.JobNameSubmittedForExecution)
			assert.Equal(t, base64.StdEncoding.EncodeToString(encodedJobArgs.Bytes()), data.JobArgs)
			assert.Equal(t, jobSubmissionStatus, data.JobSubmissionStatus)
			assert.Equal(t, jobExecutionStatus, data.JobExecutionStatus)
		}).
		Return(nil).
		Once()

	err = testStore.JobsExecutionAuditLog(jobSubmissionStatus, jobExecutionStatus, jobName, userEmail, jobNameSubmittedForExecution, imageName, jobArgs)

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

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_execution_audit_log (job_name, user_email, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status, created_at, updated_at) VALUES (:job_name, :user_email, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status, :created_at, :updated_at)",
		mock.Anything).
		Return(errors.New("error")).
		Once()

	err = testStore.JobsExecutionAuditLog("", "", "", "", "", "", map[string]string{})

	assert.Error(t, err)
	mockPostgresClient.AssertExpectations(t)
}

func TestUpdateJobsExecutionAuditLog(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	jobNameSubmittedForExecution := "any-submission"
	jobExecutionStatus := "updated-status"

	mockPostgresClient.On("NamedExec",
		"UPDATE jobs_execution_audit_log SET job_execution_status = :job_execution_status, updated_at = :updated_at where job_name_submitted_for_execution = :job_name_submitted_for_execution",
		mock.Anything).
		Run(func(args mock.Arguments) {
			data := args.Get(1).(*postgres.JobsExecutionAuditLog)

			assert.Equal(t, postgres.StringToSQLString(jobNameSubmittedForExecution), data.JobNameSubmittedForExecution)
			assert.Equal(t, jobExecutionStatus, data.JobExecutionStatus)
		}).
		Return(nil).
		Once()

	err := testStore.UpdateJobsExecutionAuditLog(jobNameSubmittedForExecution, jobExecutionStatus)

	assert.NoError(t, err)
	mockPostgresClient.AssertExpectations(t)
}

func TestGetJobsStatusWhenJobIsPresent(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)
	jobName := "any-job"

	dest := []postgres.JobsExecutionAuditLog{}

	mockPostgresClient.On("Select",
		&dest,
		"SELECT job_execution_status from jobs_execution_audit_log where job_name_submitted_for_execution = $1",
		jobName).
		Return(nil).
		Run(func(args mock.Arguments) {
			jobsExecutionAuditLogResult := args.Get(0).(*[]postgres.JobsExecutionAuditLog)
			*jobsExecutionAuditLogResult = append(*jobsExecutionAuditLogResult, postgres.JobsExecutionAuditLog{
				JobExecutionStatus: utility.JobSucceeded,
			})
		}).
		Once()

	status, err := testStore.GetJobExecutionStatus(jobName)
	assert.NoError(t, err)

	assert.Equal(t, utility.JobSucceeded, status)

	mockPostgresClient.AssertExpectations(t)
}

func TestGetJobsStatusWhenJobIsNotPresent(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)
	jobName := "any-job"

	dest := []postgres.JobsExecutionAuditLog{}

	mockPostgresClient.On("Select",
		&dest,
		"SELECT job_execution_status from jobs_execution_audit_log where job_name_submitted_for_execution = $1",
		jobName).
		Return(nil).
		Once()

	status, err := testStore.GetJobExecutionStatus(jobName)
	assert.NoError(t, err)

	assert.Equal(t, "", status)

	mockPostgresClient.AssertExpectations(t)
}

func TestGetJobsStatusWhenError(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)
	jobName := "any-job"

	dest := []postgres.JobsExecutionAuditLog{}

	mockPostgresClient.On("Select",
		&dest,
		"SELECT job_execution_status from jobs_execution_audit_log where job_name_submitted_for_execution = $1",
		jobName).
		Return(errors.New("error")).
		Once()

	_, err := testStore.GetJobExecutionStatus(jobName)
	assert.Error(t, err, "error")
}

func TestJobsScheduleInsertionSuccessfull(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_schedule (id, name, tags, time, notification_emails, user_email, args) VALUES (:id, :name, :tags, :time, :notification_emails, :user_email, :args)",
		mock.Anything).
		Return(nil).
		Once()

	scheduledJobID, err := testStore.InsertScheduledJob("job-name", "tag-one,tag-two", "* * 3 * *", "foo@bar.com,bar@foo.com", "ms@proctor.com", map[string]string{})

	assert.NoError(t, err)

	_, err = uuid.FromString(scheduledJobID)
	assert.NoError(t, err)

	mockPostgresClient.AssertExpectations(t)
}

func TestJobsScheduleInsertionFailed(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_schedule (id, name, tags, time, notification_emails, user_email, args) VALUES (:id, :name, :tags, :time, :notification_emails, :user_email, :args)",
		mock.Anything).
		Return(errors.New("any-error")).
		Once()

	_, err := testStore.InsertScheduledJob("job-name", "tag-one", "* * 3 * *", "foo@bar.com", "ms@proctor.com", map[string]string{})

	assert.Error(t, err)

	mockPostgresClient.AssertExpectations(t)
}
