package storage

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gojektech/proctor/proctord/storage/postgres"
	uuid "github.com/satori/go.uuid"
)

type Store interface {
	AuditJobsExecution(*postgres.JobsExecutionAuditLog) error
	UpdateJobsExecutionAuditLog(string, string) error
	GetJobExecutionStatus(string) (string, error)
	InsertScheduledJob(string, string, string, string, string, map[string]string) (string, error)
	GetScheduledJobs() ([]postgres.JobsSchedule, error)
}

type store struct {
	postgresClient postgres.Client
}

func New(postgresClient postgres.Client) Store {
	return &store{
		postgresClient: postgresClient,
	}
}

func (store *store) AuditJobsExecution(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog) error {
	return store.postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, user_email, image_name, job_name_submitted_for_execution, job_args, job_submission_status, job_execution_status) VALUES (:job_name, :user_email, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)", &jobsExecutionAuditLog)
}

func (store *store) UpdateJobsExecutionAuditLog(jobExecutionID, jobExecutionStatus string) error {
	jobsExecutionAuditLog := postgres.JobsExecutionAuditLog{
		JobExecutionStatus: jobExecutionStatus,
		ExecutionID:        postgres.StringToSQLString(jobExecutionID),
		UpdatedAt:          time.Now(),
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

func (store *store) GetScheduledJobs() ([]postgres.JobsSchedule, error) {
	scheduledJobs := []postgres.JobsSchedule{}
	err := store.postgresClient.Select(&scheduledJobs, "SELECT id, name, args, time, notification_emails, enabled from jobs_schedule")
	return scheduledJobs, err
}
