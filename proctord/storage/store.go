package storage

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"proctor/proctord/storage/postgres"
	"github.com/satori/go.uuid"
)

type Store interface {
	AuditJobsExecution(*postgres.JobsExecutionAuditLog) error
	UpdateJobsExecutionAuditLog(string, string) error
	GetJobExecutionStatus(string) (string, error)
	InsertScheduledJob(string, string, string, string, string, string, map[string]string) (string, error)
	GetScheduledJobs() ([]postgres.JobsSchedule, error)
	GetEnabledScheduledJobs() ([]postgres.JobsSchedule, error)
	GetScheduledJob(string) ([]postgres.JobsSchedule, error)
	RemoveScheduledJob(string) (int64, error)
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
	_, err := store.postgresClient.NamedExec("INSERT INTO jobs_execution_audit_log (job_name, user_email, image_name, job_name_submitted_for_execution, job_args, job_submission_status,"+
		" job_execution_status) VALUES (:job_name, :user_email, :image_name, :job_name_submitted_for_execution, :job_args, :job_submission_status, :job_execution_status)",
		&jobsExecutionAuditLog)
	return err
}

func (store *store) UpdateJobsExecutionAuditLog(jobExecutionID, jobExecutionStatus string) error {
	jobsExecutionAuditLog := postgres.JobsExecutionAuditLog{
		JobExecutionStatus: jobExecutionStatus,
		ExecutionID:        postgres.StringToSQLString(jobExecutionID),
		UpdatedAt:          time.Now(),
	}

	_, err := store.postgresClient.NamedExec("UPDATE jobs_execution_audit_log SET job_execution_status = :job_execution_status, updated_at = :updated_at where job_name_submitted_for_execution = "+
		":job_name_submitted_for_execution", &jobsExecutionAuditLog)
	return err
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

func (store *store) InsertScheduledJob(name, tags, time, notificationEmails, userEmail, groupName string, args map[string]string) (string, error) {
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
		Group:              groupName,
		Enabled:            true,
	}
	_, err = store.postgresClient.NamedExec("INSERT INTO jobs_schedule (id, name, tags, time, notification_emails, user_email, group_name, args, enabled) "+
		"VALUES (:id, :name, :tags, :time, :notification_emails, :user_email,  :group_name, :args, :enabled)", &jobsSchedule)
	return jobsSchedule.ID, err
}

func (store *store) GetScheduledJobs() ([]postgres.JobsSchedule, error) {
	scheduledJobs := []postgres.JobsSchedule{}
	err := store.postgresClient.Select(&scheduledJobs, "SELECT id, name, args, time, notification_emails, group_name, enabled from jobs_schedule")
	return scheduledJobs, err
}

func (store *store) GetEnabledScheduledJobs() ([]postgres.JobsSchedule, error) {
	scheduledJobs := []postgres.JobsSchedule{}
	err := store.postgresClient.Select(&scheduledJobs, "SELECT id, name, args, time, tags, notification_emails,group_name from jobs_schedule where enabled = 't'")
	return scheduledJobs, err
}

func (store *store) GetScheduledJob(jobID string) ([]postgres.JobsSchedule, error) {
	scheduledJob := []postgres.JobsSchedule{}
	err := store.postgresClient.Select(&scheduledJob, "SELECT id, name, args, time, tags, notification_emails,group_name from jobs_schedule where id = $1 and enabled = 't'", jobID)
	return scheduledJob, err
}

func (store *store) RemoveScheduledJob(jobID string) (int64, error) {
	job := postgres.JobsSchedule{
		ID:        jobID,
		UpdatedAt: time.Now(),
	}
	rowsAffected, err := store.postgresClient.NamedExec("UPDATE jobs_schedule set enabled = 'f', updated_at = :updated_at where id = :id and enabled = 't'", &job)
	return rowsAffected, err
}
