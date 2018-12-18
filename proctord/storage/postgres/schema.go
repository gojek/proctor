package postgres

import (
	"database/sql"
	"time"
)

type JobsExecutionAuditLog struct {
	JobName                      string         `db:"job_name"`
	UserEmail                    string         `db:"user_email"`
	ImageName                    string         `db:"image_name"`
	JobNameSubmittedForExecution sql.NullString `db:"job_name_submitted_for_execution"`
	JobArgs                      string         `db:"job_args"`
	JobSubmissionStatus          string         `db:"job_submission_status"`
	JobExecutionStatus           string         `db:"job_execution_status"`
	CreatedAt                    time.Time      `db:"created_at"`
	UpdatedAt                    time.Time      `db:"updated_at"`
}
