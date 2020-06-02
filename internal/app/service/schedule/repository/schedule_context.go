package repository

import (
	"github.com/pkg/errors"
	executionModel "proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/id"
	"proctor/internal/app/service/schedule/model"
)

type ScheduleContextRepository interface {
	Insert(context model.ScheduleContext) (*model.ScheduleContext, error)
	Delete(id uint64) error
	GetByID(id uint64) (*model.ScheduleContext, error)
	GetContextByScheduleId(scheduleId uint64) ([]executionModel.ExecutionContext, error)
	GetScheduleByContextId(contextId uint64) (*model.Schedule, error)
	deleteAll() error
}

type scheduleContextRepository struct {
	postgresqlClient postgresql.Client
}

func NewScheduleContextRepository(client postgresql.Client) ScheduleContextRepository {
	return &scheduleContextRepository{
		postgresqlClient: client,
	}
}

func (repository *scheduleContextRepository) Insert(context model.ScheduleContext) (*model.ScheduleContext, error) {
	snowflakeID, _ := id.NextID()
	context.ID = snowflakeID
	sql := "INSERT INTO schedule_context (id,schedule_id, execution_context_id) VALUES (:id, :schedule_id, :execution_context_id)"
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	if err != nil {
		return nil, err
	}
	return &context, nil
}

func (repository *scheduleContextRepository) Delete(id uint64) error {
	sql := "DELETE FROM schedule_context WHERE id = :id"
	schedule := model.Schedule{
		ID: id,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &schedule)
	return err
}

func (repository *scheduleContextRepository) GetByID(id uint64) (*model.ScheduleContext, error) {
	sql := "SELECT id, schedule_id, execution_context_id, created_at, updated_at FROM schedule_context WHERE id=$1 "
	var schedules []model.ScheduleContext
	err := repository.postgresqlClient.Select(&schedules, sql, id)
	if err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return nil, errors.Errorf("Execution context with id %v is not found!", id)
	}

	return &schedules[0], nil
}

func (repository *scheduleContextRepository) GetContextByScheduleId(scheduleId uint64) ([]executionModel.ExecutionContext, error) {
	sql := "SELECT e.id, e.job_name, e.name, e.user_email, e.image_tag, e.args, e.output, e.status, e.created_at, e.updated_at FROM execution_context e INNER JOIN schedule_context sc ON e.id = sc.execution_context_id WHERE sc.schedule_id=$1"
	var contexts []executionModel.ExecutionContext
	err := repository.postgresqlClient.Select(&contexts, sql, scheduleId)
	if err != nil {
		return nil, err
	}

	return contexts, nil
}

func (repository *scheduleContextRepository) GetScheduleByContextId(contextId uint64) (*model.Schedule, error) {
	sql := "SELECT s.id, s.job_name, s.args, s.cron, s.notification_emails, s.user_email,s.\"group\", s.enabled, s.created_at, s.updated_at  FROM schedule s INNER JOIN schedule_context sc ON s.id = sc.schedule_id WHERE sc.execution_context_id=$1 "
	var schedules []model.Schedule
	err := repository.postgresqlClient.Select(&schedules, sql, contextId)
	if err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return nil, errors.Errorf("Execution context with id %v is not found!", contextId)
	}

	return &schedules[0], nil
}

func (repository *scheduleContextRepository) deleteAll() error {
	sql := "DELETE FROM schedule_context"
	schedules := model.ScheduleContext{}
	_, err := repository.postgresqlClient.NamedExec(sql, schedules)
	return err
}
