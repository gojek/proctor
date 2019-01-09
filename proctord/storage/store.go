package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"time"

	"github.com/gojektech/proctor/proctord/storage/postgres"
	uuid "github.com/satori/go.uuid"
)

type Store interface {
	JobsExecutionAuditLog(string, string, string, string, string, string, map[string]string) error
	UpdateJobsExecutionAuditLog(string, string) error
	GetJobExecutionStatus(string) (string, error)
	InsertScheduledJob(string, string, string, string, string, map[string]string) (string, error)
}

type store struct {
	postgresClient postgres.Client
}

func New(postgresClient postgres.Client) Store {
	return &store{
		postgresClient: postgresClient,
	}
}

func (store *store) JobsExecutionAuditLog(jobSubmissionStatus, jobExecutionStatus, jobName, userEmail, JobNameSubmittedForExecution, imageName string, jobArgs map[string]string) error {
	var encodedJobArgs bytes.Buffer
	enc := gob.NewEncoder(&encodedJobArgs)
	err := enc.Encode(jobArgs)
	if err != nil {
		return err
	}

	jobsExecutionAuditLog := postgres.JobsExecutionAuditLog{
		JobName:                      jobName,
		UserEmail:                    userEmail,
		ImageName:                    imageName,
		JobNameSubmittedForExecution: postgres.StringToSQLString(JobNameSubmittedForExecution),
		JobArgs:             base64.StdEncoding.EncodeToString(encodedJobArgs.Bytes()),
		JobSubmissionStatus: jobSubmissionStatus,
		JobExecutionStatus:  jobExecutionStatus,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	return store.postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, user_email, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status, created_at, updated_at) VALUES (:job_name, :user_email, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status, :created_at, :updated_at)", &jobsExecutionAuditLog)
}

func (store *store) UpdateJobsExecutionAuditLog(JobNameSubmittedForExecution, jobExecutionStatus string) error {
	jobsExecutionAuditLog := postgres.JobsExecutionAuditLog{
		JobExecutionStatus:           jobExecutionStatus,
		JobNameSubmittedForExecution: postgres.StringToSQLString(JobNameSubmittedForExecution),
		UpdatedAt:                    time.Now(),
	}

	return store.postgresClient.NamedExec("UPDATE jobs_execution_audit_log SET job_execution_status = :job_execution_status, updated_at = :updated_at where job_name_submitted_for_execution = :job_name_submitted_for_execution", &jobsExecutionAuditLog)
}

func (store *store) GetJobExecutionStatus(JobNameSubmittedForExecution string) (string, error) {
	jobsExecutionAuditLogResult := []postgres.JobsExecutionAuditLog{}
	err := store.postgresClient.Select(&jobsExecutionAuditLogResult, "SELECT job_execution_status from jobs_execution_audit_log where job_name_submitted_for_execution = $1", JobNameSubmittedForExecution)
	if err != nil {
		return "", err
	}

	if len(jobsExecutionAuditLogResult) == 0 {
		return "", nil
	}

	return jobsExecutionAuditLogResult[0].JobExecutionStatus, nil
}

func (store *store) InsertScheduledJob(name, tags, time, notificationEmails, userEmail string, args map[string]string) (string, error) {
	jsonEncodedArgs, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	jobsSchedule := postgres.JobsSchedule{
		ID:                 uuid.NewV4().String(),
		Name:               name,
		Args:               base64.StdEncoding.EncodeToString(jsonEncodedArgs),
		Tags:               tags,
		Time:               time,
		NotificationEmails: notificationEmails,
		UserEmail:          userEmail,
		Enabled:            true,
	}
	return jobsSchedule.ID, store.postgresClient.NamedExec("INSERT INTO jobs_schedule (id, name, tags, time, notification_emails, user_email, args, enabled) VALUES (:id, :name, :tags, :time, :notification_emails, :user_email, :args, :enabled)", &jobsSchedule)
}
