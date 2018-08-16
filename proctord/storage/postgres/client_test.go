package postgres

import (
	"fmt"
	"testing"

	"github.com/gojektech/proctor/proctord/config"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNamedExec(t *testing.T) {
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.PostgresDatabase(), config.PostgresUser(), config.PostgresPassword(), config.PostgresHost())

	db, err := sqlx.Connect("postgres", dataSourceName)
	assert.NoError(t, err)

	postgresClient := &client{db: db}
	defer postgresClient.db.Close()

	jobsExecutionAuditLog := &JobsExecutionAuditLog{
		JobName:                      "test-job-name",
		ImageName:                    "test-image-name",
		JobNameSubmittedForExecution: "test-submission-name",
		JobArgs:             "test-job-args",
		JobSubmissionStatus: "test-job-status",
		JobExecutionStatus:  "test-job-execution-status",
	}

	err = postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)", jobsExecutionAuditLog)
	assert.NoError(t, err)

	var persistedJobsExecutionAuditLog JobsExecutionAuditLog
	err = postgresClient.db.Get(&persistedJobsExecutionAuditLog, `SELECT job_name, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status FROM jobs_execution_audit_log WHERE job_name='test-job-name'`)
	assert.NoError(t, err)

	assert.Equal(t, jobsExecutionAuditLog.JobName, persistedJobsExecutionAuditLog.JobName)
	assert.Equal(t, jobsExecutionAuditLog.ImageName, persistedJobsExecutionAuditLog.ImageName)
	assert.Equal(t, jobsExecutionAuditLog.JobNameSubmittedForExecution, persistedJobsExecutionAuditLog.JobNameSubmittedForExecution)
	assert.Equal(t, jobsExecutionAuditLog.JobArgs, persistedJobsExecutionAuditLog.JobArgs)
	assert.Equal(t, jobsExecutionAuditLog.JobSubmissionStatus, persistedJobsExecutionAuditLog.JobSubmissionStatus)
	assert.Equal(t, jobsExecutionAuditLog.JobExecutionStatus, persistedJobsExecutionAuditLog.JobExecutionStatus)

	_, err = postgresClient.db.Exec("DELETE FROM jobs_execution_audit_log WHERE job_name='test-job-name'")
	assert.NoError(t, err)
}

func TestSelect(t *testing.T) {
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.PostgresDatabase(), config.PostgresUser(), config.PostgresPassword(), config.PostgresHost())

	db, err := sqlx.Connect("postgres", dataSourceName)
	assert.NoError(t, err)

	postgresClient := &client{db: db}
	defer postgresClient.db.Close()
	jobName := "test-job-name"

	jobsExecutionAuditLog := &JobsExecutionAuditLog{
		JobName:                      jobName,
		ImageName:                    "test-image-name",
		JobNameSubmittedForExecution: "test-submission-name",
		JobArgs:             "test-job-args",
		JobSubmissionStatus: "test-job-status",
		JobExecutionStatus:  "test-job-execution-status",
	}

	err = postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)", jobsExecutionAuditLog)
	assert.NoError(t, err)

	jobsExecutionAuditLogResult := []JobsExecutionAuditLog{}
	err = postgresClient.db.Select(&jobsExecutionAuditLogResult, "SELECT job_execution_status from jobs_execution_audit_log where job_name = $1", jobName)
	assert.NoError(t, err)

	assert.Equal(t, jobsExecutionAuditLog.JobExecutionStatus, jobsExecutionAuditLogResult[0].JobExecutionStatus)

	_, err = postgresClient.db.Exec("DELETE FROM jobs_execution_audit_log WHERE job_name='test-job-name'")
	assert.NoError(t, err)
}

func TestSelectForNoRows(t *testing.T) {
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.PostgresDatabase(), config.PostgresUser(), config.PostgresPassword(), config.PostgresHost())

	db, err := sqlx.Connect("postgres", dataSourceName)
	assert.NoError(t, err)

	postgresClient := &client{db: db}
	defer postgresClient.db.Close()
	jobName := "test-job-name"

	jobsExecutionAuditLogResult := []JobsExecutionAuditLog{}
	err = postgresClient.db.Select(&jobsExecutionAuditLogResult, "SELECT job_execution_status from jobs_execution_audit_log where job_name = $1", jobName)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(jobsExecutionAuditLogResult))

	assert.NoError(t, err)
}
