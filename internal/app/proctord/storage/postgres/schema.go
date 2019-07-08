package postgres

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	 "proctor/internal/app/proctord/logger"
	"time"
)

type JobsExecutionAuditLog struct {
	JobName             string         `db:"job_name"`
	UserEmail           string         `db:"user_email"`
	ImageName           string         `db:"image_name"`
	ExecutionID         sql.NullString `db:"job_name_submitted_for_execution"`
	JobArgs             string         `db:"job_args"`
	JobSubmissionStatus string         `db:"job_submission_status"`
	Errors              string         `db:"errors"`
	JobExecutionStatus  string         `db:"job_execution_status"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
}

func (j *JobsExecutionAuditLog) AddJobArgs(jobArgs map[string]string) {
	jsonEncodedArgs, err := json.Marshal(jobArgs)
	if err != nil {
		logger.Error("Error marshaling job args: ", err.Error())
		return
	}

	j.JobArgs = base64.StdEncoding.EncodeToString(jsonEncodedArgs)
}

func (j *JobsExecutionAuditLog) AddExecutionID(jobExecutionID string) {
	j.ExecutionID = StringToSQLString(jobExecutionID)
}

type JobsSchedule struct {
	ID                 string    `db:"id"`
	Name               string    `db:"name"`
	Args               string    `db:"args"`
	Tags               string    `db:"tags"`
	Time               string    `db:"time"`
	NotificationEmails string    `db:"notification_emails"`
	UserEmail          string    `db:"user_email"`
	Group              string    `db:"group_name"`
	Enabled            bool      `db:"enabled"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}
