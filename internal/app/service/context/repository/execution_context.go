package repository

import (
	"github.com/jmoiron/sqlx/types"
	"proctor/internal/app/service/context/model"
	"time"
)

type ExecutionContextRepository interface {
	Insert(context model.ExecutionContext) error
	UpdateJobOutput(executionId string, status types.GzippedText) error
	UpdateStatus(executionId string, status string) error
	Delete(executionId string) error
	GetById(executionId string) (model.ExecutionContext, error)
	GetByEmail(userEmail string) ([]model.ExecutionContext, error)
	GetByJobName(jobName string) ([]model.ExecutionContext, error)
	GetByStatus(status string) ([]model.ExecutionContext, error)
	GetByCreationTime(start time.Time, end time.Time) ([]model.ExecutionContext, error)
}

type executionContextRepository struct {

}
