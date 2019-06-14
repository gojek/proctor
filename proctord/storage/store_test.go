package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"proctor/proctord/storage/postgres"
	"proctor/proctord/utility"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestJobsExecutionAuditLog(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	jobExecutionAuditLog := &postgres.JobsExecutionAuditLog{
		JobName:   "sample-job",
		ImageName: "any-image",
		UserEmail: "mrproctor@example.com",
	}

	var encodedJobArgs bytes.Buffer
	enc := gob.NewEncoder(&encodedJobArgs)
	err := enc.Encode(map[string]string{})
	assert.NoError(t, err)

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_execution_audit_log (job_name, user_email, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :user_email, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)", mock.Anything).Run(func(args mock.Arguments) {
	}).Return(int64(1), nil).Once()

	err = testStore.AuditJobsExecution(jobExecutionAuditLog)

	assert.NoError(t, err)
	mockPostgresClient.AssertExpectations(t)
}

func TestJobsExecutionAuditLogPostgresClientFailure(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	jobExecutionAuditLog := &postgres.JobsExecutionAuditLog{
		JobName: "sample-job",
	}

	var encodedJobArgs bytes.Buffer
	enc := gob.NewEncoder(&encodedJobArgs)
	err := enc.Encode(map[string]string{})
	assert.NoError(t, err)

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_execution_audit_log (job_name, user_email, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :user_email, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)",
		mock.Anything).
		Return(int64(0), errors.New("error")).
		Once()

	err = testStore.AuditJobsExecution(jobExecutionAuditLog)

	assert.Error(t, err)
	mockPostgresClient.AssertExpectations(t)
}

func TestUpdateJobsExecutionAuditLog(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	executionID := "any-submission"
	jobExecutionStatus := "updated-status"

	mockPostgresClient.On("NamedExec",
		"UPDATE jobs_execution_audit_log SET job_execution_status = :job_execution_status, updated_at = :updated_at where job_name_submitted_for_execution = :job_name_submitted_for_execution",
		mock.Anything).
		Run(func(args mock.Arguments) {
			data := args.Get(1).(*postgres.JobsExecutionAuditLog)

			assert.Equal(t, postgres.StringToSQLString(executionID), data.ExecutionID)
			assert.Equal(t, jobExecutionStatus, data.JobExecutionStatus)
		}).
		Return(int64(1), nil).
		Once()

	err := testStore.UpdateJobsExecutionAuditLog(executionID, jobExecutionStatus)

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
	postgresClient := postgres.NewClient()
	testStore := New(postgresClient)

	scheduledJobID, err := testStore.InsertScheduledJob("job-name", "tag-one", "* * 3 * *", "foo@bar.com", "ms@proctor.com","group1", map[string]string{})
	assert.NoError(t, err)
	_, err = uuid.FromString(scheduledJobID)
	assert.NoError(t, err)

	_, err = postgresClient.GetDB().Exec("truncate table jobs_schedule;")
	assert.NoError(t, err)
}

func TestJobsScheduleInsertionFailed(t *testing.T) {
	mockPostgresClient := &postgres.ClientMock{}
	testStore := New(mockPostgresClient)

	jobName := "job-name"
	tag := "tag-one1"
	time := "* * 3 * *"
	notificationEmail := "foo@bar.com"
	userEmail := "ms@proctor.com"
	groupName := "group1"

	mockPostgresClient.On("NamedExec",
		"INSERT INTO jobs_schedule (id, name, tags, time, notification_emails, user_email, group_name, args, enabled) "+
	"VALUES (:id, :name, :tags, :time, :notification_emails, :user_email,  :group_name, :args, :enabled)",
		mock.AnythingOfType("*postgres.JobsSchedule")).Run(func(args mock.Arguments) {
	}).Return(int64(0), errors.New("any-error")).
		Once()

	_, err := testStore.InsertScheduledJob(jobName, tag, time, notificationEmail, userEmail,groupName, map[string]string{})

	assert.Error(t, err)

	mockPostgresClient.AssertExpectations(t)
}

func TestGetScheduledJobByID(t *testing.T) {
	postgresClient := postgres.NewClient()
	testStore := New(postgresClient)

	jobID, err := testStore.InsertScheduledJob("job-name", "tag-one", "* * 3 * *", "foo@bar.com", "ms@proctor.com","group1", map[string]string{})
	assert.NoError(t, err)

	resultJob, err := testStore.GetScheduledJob(jobID)
	assert.NoError(t, err)
	assert.Equal(t, "job-name", resultJob[0].Name)
	assert.Equal(t, "tag-one", resultJob[0].Tags)
	assert.Equal(t, "* * 3 * *", resultJob[0].Time)

	_, err = postgresClient.GetDB().Exec("truncate table jobs_schedule;")
	assert.NoError(t, err)
}

func TestGetScheduledJobByIDReturnErrorIfIDnotFound(t *testing.T) {
	postgresClient := postgres.NewClient()
	testStore := New(postgresClient)

	resultJob, err := testStore.GetScheduledJob("86A7963B-3621-492D-8D6C-33076242256B")
	assert.NoError(t, err)
	assert.Equal(t, []postgres.JobsSchedule{}, resultJob)
}

func TestRemoveScheduledJobByID(t *testing.T) {
	postgresClient := postgres.NewClient()
	testStore := New(postgresClient)

	jobID, err := testStore.InsertScheduledJob("job-name", "tag-one", "* * 3 * *", "foo@bar.com", "ms@proctor.com","group1", map[string]string{})
	assert.NoError(t, err)

	removedJobsCount, err := testStore.RemoveScheduledJob(jobID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), removedJobsCount)

	_, err = postgresClient.GetDB().Exec("truncate table jobs_schedule;")
	assert.NoError(t, err)
}

func TestRemoveScheduledJobByIDReturnErrorIfIDnotFound(t *testing.T) {
	postgresClient := postgres.NewClient()
	testStore := New(postgresClient)

	removedJobsCount, err := testStore.RemoveScheduledJob("86A7963B-3621-492D-8D6C-33076242256B")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), removedJobsCount)
}

func TestRemoveScheduledJobByIDReturnErrorIfIDIsInvalid(t *testing.T) {
	postgresClient := postgres.NewClient()
	testStore := New(postgresClient)

	removedJobsCount, err := testStore.RemoveScheduledJob("86A7963B")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input syntax")
	assert.Equal(t, int64(0), removedJobsCount)
}
