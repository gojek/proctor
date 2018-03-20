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
		JobName:                  "test-job-name",
		ImageName:                "test-image-name",
		JobSubmittedForExecution: "test-submission-name",
		JobArgs:                  "test-job-args",
		JobSubmissionStatus:      "test-job-status",
	}

	err = postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, image_name, job_submitted_for_execution, job_args, job_submission_status) VALUES (:job_name, :image_name, :job_submitted_for_execution, :job_args, :job_submission_status)", jobsExecutionAuditLog)
	assert.NoError(t, err)

	var persistedJobsExecutionAuditLog JobsExecutionAuditLog
	err = postgresClient.db.Get(&persistedJobsExecutionAuditLog, `SELECT job_name, image_name, job_submitted_for_execution, job_args, job_submission_status FROM jobs_execution_audit_log WHERE job_name='test-job-name'`)
	assert.NoError(t, err)

	assert.Equal(t, jobsExecutionAuditLog.JobName, persistedJobsExecutionAuditLog.JobName)
	assert.Equal(t, jobsExecutionAuditLog.ImageName, persistedJobsExecutionAuditLog.ImageName)
	assert.Equal(t, jobsExecutionAuditLog.JobSubmittedForExecution, persistedJobsExecutionAuditLog.JobSubmittedForExecution)
	assert.Equal(t, jobsExecutionAuditLog.JobArgs, persistedJobsExecutionAuditLog.JobArgs)
	assert.Equal(t, jobsExecutionAuditLog.JobSubmissionStatus, persistedJobsExecutionAuditLog.JobSubmissionStatus)

	_, err = postgresClient.db.Exec("DELETE FROM jobs_execution_audit_log WHERE job_name='test-job-name'")
	assert.NoError(t, err)
}
