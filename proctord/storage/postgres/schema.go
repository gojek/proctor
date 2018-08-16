package postgres

import "database/sql"

type JobsExecutionAuditLog struct {
	JobName                      string         `db:"job_name"`
	ImageName                    string         `db:"image_name"`
	JobNameSubmittedForExecution sql.NullString `db:"job_name_submitted_for_execution"`
	JobArgs                      string         `db:"job_args"`
	JobSubmissionStatus          string         `db:"job_submission_status"`
	JobExecutionStatus           string         `db:"job_execution_status"`
}
