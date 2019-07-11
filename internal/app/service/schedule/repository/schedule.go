package repository

import (
	"github.com/pkg/errors"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/id"
	"proctor/internal/app/service/schedule/model"
)

type ScheduleRepository interface {
	Insert(context *model.Schedule) (uint64, error)
	Delete(id uint64) error
	GetById(id uint64) (*model.Schedule, error)
	GetByUserEmail(userEmail string) ([]model.Schedule, error)
	GetAllEnabled() ([]model.Schedule, error)
	GetAll() ([]model.Schedule, error)
	GetEnabledById(id uint64) (*model.Schedule, error)
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

func (repository *scheduleRepository) Insert(context *model.Schedule) (uint64, error) {
	snowflakeId, _ := id.NextId()
	context.ID = snowflakeId
	sql := "INSERT INTO schedule (id, job_name, args,cron,notification_emails, user_email, \"group\", enabled) VALUES (:id, :job_name, :args, :cron, :notification_emails, :user_email, :group, :enabled)"
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	if err != nil {
		return 0, nil
	}
	return snowflakeId, nil
}

func (repository *scheduleRepository) Delete(id uint64) error {
	sql := "DELETE FROM schedule WHERE id = :id"
	schedule := model.Schedule{
		ID: id,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &schedule)
	return err
}

func (repository *scheduleRepository) GetById(id uint64) (*model.Schedule, error) {
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

func (repository *scheduleRepository) GetEnabledById(id uint64) (*model.Schedule, error) {
	sql := "SELECT id, job_name, args, cron, notification_emails, user_email, \"group\", enabled, created_at, updated_at FROM schedule WHERE enabled=$1 AND id=$2 "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql, true, id)
	if err != nil {
		return nil, err
	}

	return &schedules[0], nil
}

func (repository *scheduleRepository) deleteAll() error {
	sql := "DELETE FROM schedule"
	schedules := model.Schedule{}
	_, err := repository.postgresqlClient.NamedExec(sql, schedules)
	return err
}
