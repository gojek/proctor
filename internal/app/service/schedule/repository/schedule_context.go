package repository

import (
	executionModel "proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/schedule/model"
)

type ScheduleContextRepository interface {
	Insert(context *model.ScheduleContext) (uint64, error)
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

func (repository *scheduleContextRepository) Insert(context *model.ScheduleContext) (uint64, error) {
	panic("implement me")
}

func (repository *scheduleContextRepository) Delete(id uint64) error {
	panic("implement me")
}

func (repository *scheduleContextRepository) GetByID(id uint64) (*model.ScheduleContext, error) {
	panic("implement me")
}

func (repository *scheduleContextRepository) GetContextByScheduleId(scheduleId uint64) ([]executionModel.ExecutionContext, error) {
	panic("implement me")
}

func (repository *scheduleContextRepository) GetScheduleByContextId(contextId uint64) (*model.Schedule, error) {
	panic("implement me")
}

func (repository *scheduleContextRepository) deleteAll() error {
	sql := "DELETE FROM schedule_context"
	schedules := model.ScheduleContext{}
	_, err := repository.postgresqlClient.NamedExec(sql, schedules)
	return err
}
