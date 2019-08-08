package repository

import (
	"github.com/pkg/errors"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/id"
	"proctor/internal/app/service/schedule/model"
	"time"
)

type ScheduleRepository interface {
	Insert(context model.Schedule) (uint64, error)
	Delete(id uint64) error
	GetByID(id uint64) (*model.Schedule, error)
	Disable(id uint64) error
	Enable(id uint64) error
	GetByUserEmail(userEmail string) ([]model.Schedule, error)
	GetByJobName(jobName string) ([]model.Schedule, error)
	GetAllEnabled() ([]model.Schedule, error)
	GetAll() ([]model.Schedule, error)
	GetEnabledByID(id uint64) (*model.Schedule, error)
	deleteAll() error
}

type scheduleRepository struct {
	postgresqlClient postgresql.Client
}

func NewScheduleRepository(client postgresql.Client) ScheduleRepository {
	return &scheduleRepository{
		postgresqlClient: client,
	}
}

func (repository *scheduleRepository) Enable(id uint64) error {
	sql := "UPDATE schedule SET enabled = :enabled, updated_at = :updated_at WHERE id = :id"
	context := model.Schedule{
		ID:        id,
		UpdatedAt: time.Now(),
		Enabled:   true,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	return err
}

func (repository *scheduleRepository) Disable(id uint64) error {
	sql := "UPDATE schedule SET enabled = :enabled, updated_at = :updated_at WHERE id = :id"
	context := model.Schedule{
		ID:        id,
		UpdatedAt: time.Now(),
		Enabled:   false,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	return err
}

func (repository *scheduleRepository) Insert(context model.Schedule) (uint64, error) {
	snowflakeID, _ := id.NextID()
	context.ID = snowflakeID
	sql := "INSERT INTO schedule (id, job_name, args,cron,notification_emails, user_email, \"group\", enabled) VALUES (:id, :job_name, :args, :cron, :notification_emails, :user_email, :group, :enabled)"
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	if err != nil {
		return 0, nil
	}
	return snowflakeID, nil
}

func (repository *scheduleRepository) Delete(id uint64) error {
	sql := "DELETE FROM schedule WHERE id = :id"
	schedule := model.Schedule{
		ID: id,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &schedule)
	return err
}

func (repository *scheduleRepository) GetByID(id uint64) (*model.Schedule, error) {
	sql := "SELECT id, job_name, args, cron, notification_emails, user_email,\"group\", enabled, created_at, updated_at FROM schedule WHERE id=$1 "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql, id)
	if err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return nil, errors.Errorf("Execution context with id %v is not found!", id)
	}

	return &schedules[0], nil
}

func (repository *scheduleRepository) GetByUserEmail(userEmail string) ([]model.Schedule, error) {
	sql := "SELECT id, job_name, args, cron, notification_emails, user_email, \"group\", enabled, created_at, updated_at FROM schedule WHERE user_email=$1 "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql, userEmail)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (repository *scheduleRepository) GetByJobName(jobName string) ([]model.Schedule, error) {
	sql := "SELECT id, job_name, args, cron, notification_emails, user_email, \"group\", enabled, created_at, updated_at FROM schedule WHERE job_name=$1 "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql, jobName)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (repository *scheduleRepository) GetAllEnabled() ([]model.Schedule, error) {
	sql := "SELECT id, job_name, args, cron, notification_emails, user_email, \"group\", enabled, created_at, updated_at FROM schedule WHERE enabled=$1 "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql, true)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (repository *scheduleRepository) GetAll() ([]model.Schedule, error) {
	sql := "SELECT id, job_name, args, cron, notification_emails, user_email, \"group\", enabled, created_at, updated_at FROM schedule "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (repository *scheduleRepository) GetEnabledByID(id uint64) (*model.Schedule, error) {
	sql := "SELECT id, job_name, args, cron, notification_emails, user_email, \"group\", enabled, created_at, updated_at FROM schedule WHERE enabled=$1 AND id=$2 "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql, true, id)
	if err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return nil, errors.Errorf("Execution context with id %v is not found!", id)
	}

	return &schedules[0], nil
}

func (repository *scheduleRepository) deleteAll() error {
	sql := "DELETE FROM schedule"
	schedules := model.Schedule{}
	_, err := repository.postgresqlClient.NamedExec(sql, schedules)
	return err
}
