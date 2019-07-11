package repository

import (
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/id"
	"time"
)

type ExecutionContextRepository interface {
	Insert(context *model.ExecutionContext) (uint64, error)
	UpdateJobOutput(executionId uint64, output types.GzippedText) error
	UpdateStatus(executionId uint64, status string) error
	Delete(executionId uint64) error
	GetById(executionId uint64) (*model.ExecutionContext, error)
	GetByEmail(userEmail string) ([]model.ExecutionContext, error)
	GetByJobName(jobName string) ([]model.ExecutionContext, error)
	GetByStatus(status string) ([]model.ExecutionContext, error)
	deleteAll() error
}

type executionContextRepository struct {
	postgresqlClient postgresql.Client
}

func NewExecutionContextRepository(client postgresql.Client) ExecutionContextRepository {
	return &executionContextRepository{
		postgresqlClient: client,
	}
}

func (repository *executionContextRepository) Insert(context *model.ExecutionContext) (uint64, error) {
	snowflakeId, _ := id.NextId()
	context.ExecutionID = snowflakeId
	sql := "INSERT INTO execution_context (id, job_name, user_email, image_tag, args, output, status) VALUES (:id, :job_name, :user_email, :image_tag, :args, :output, :status)"
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	if err != nil {
		return 0, nil
	}
	return snowflakeId, nil
}

func (repository *executionContextRepository) UpdateJobOutput(executionId uint64, output types.GzippedText) error {
	sql := "UPDATE execution_context SET output = :output, updated_at = :updated_at WHERE id = :id"
	context := model.ExecutionContext{
		ExecutionID: executionId,
		UpdatedAt:   time.Now(),
		Output:      output,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	return err
}

func (repository *executionContextRepository) UpdateStatus(executionId uint64, status string) error {
	sql := "UPDATE execution_context SET status = :status, updated_at = :updated_at WHERE id = :id"
	context := model.ExecutionContext{
		ExecutionID: executionId,
		UpdatedAt:   time.Now(),
		Status:      status,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	return err
}

func (repository *executionContextRepository) Delete(executionId uint64) error {
	sql := "DELETE FROM execution_context WHERE id = :id"
	context := model.ExecutionContext{
		ExecutionID: executionId,
	}
	_, err := repository.postgresqlClient.NamedExec(sql, &context)
	return err
}

func (repository *executionContextRepository) GetById(executionId uint64) (*model.ExecutionContext, error) {
	sql := "SELECT id, job_name, user_email, image_tag, args, output, status, created_at, updated_at FROM execution_context WHERE id=$1 "
	var contexts []model.ExecutionContext
	err := repository.postgresqlClient.Select(&contexts, sql, executionId)
	if err != nil {
		return nil, err
	}

	if len(contexts) == 0 {
		return nil, errors.Errorf("Execution context with id %v is not found!", executionId)
	}

	return &contexts[0], nil
}

func (repository *executionContextRepository) GetByEmail(userEmail string) ([]model.ExecutionContext, error) {
	sql := "SELECT id, job_name, user_email, image_tag, args, output, status, created_at, updated_at FROM execution_context WHERE user_email=$1 "
	var contexts []model.ExecutionContext
	err := repository.postgresqlClient.Select(&contexts, sql, userEmail)
	if err != nil {
		return nil, err
	}

	return contexts, nil
}

func (repository *executionContextRepository) GetByJobName(jobName string) ([]model.ExecutionContext, error) {
	sql := "SELECT id, job_name, user_email, image_tag, args, output, status, created_at, updated_at FROM execution_context WHERE job_name=$1 "
	var contexts []model.ExecutionContext
	err := repository.postgresqlClient.Select(&contexts, sql, jobName)
	if err != nil {
		return nil, err
	}

	return contexts, nil
}

func (repository *executionContextRepository) GetByStatus(status string) ([]model.ExecutionContext, error) {
	sql := "SELECT id, job_name, user_email, image_tag, args, output, status, created_at, updated_at FROM execution_context WHERE status=$1 "
	var contexts []model.ExecutionContext
	err := repository.postgresqlClient.Select(&contexts, sql, status)
	if err != nil {
		return nil, err
	}

	return contexts, nil
}

func (repository *executionContextRepository) deleteAll() error {
	sql := "DELETE FROM execution_context"
	context := model.ExecutionContext{}
	_, err := repository.postgresqlClient.NamedExec(sql, context)
	return err
}
