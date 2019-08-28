package postgresql

import (
	"fmt"
	"proctor/internal/app/service/infra/id"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	executionContextModel "proctor/internal/app/service/execution/model"
	executionContextStatus "proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/config"
)

func TestNamedExec(t *testing.T) {
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.PostgresDatabase(), config.PostgresUser(), config.PostgresPassword(), config.PostgresHost())

	db, err := sqlx.Connect("postgres", dataSourceName)
	assert.NoError(t, err)

	postgresClient := &client{db: db}
	defer postgresClient.db.Close()

	executionContext := &executionContextModel.ExecutionContext{
		JobName:     "test-job-name",
		ImageTag:    "test-image-name",
		ExecutionID: uint64(1),
		Args:        map[string]string{"foo": "bar"},
		Status:      executionContextStatus.Finished,
	}

	_, err = postgresClient.NamedExec("INSERT INTO execution_context (id, job_name, image_tag, args, status) VALUES (:id, :job_name, :image_tag, :args, :status)", executionContext)
	assert.NoError(t, err)

	var persistedExecutionContext executionContextModel.ExecutionContext
	err = postgresClient.db.Get(&persistedExecutionContext, `SELECT id, job_name, image_tag, args, status FROM execution_context WHERE job_name='test-job-name'`)
	assert.NoError(t, err)

	assert.Equal(t, executionContext.JobName, persistedExecutionContext.JobName)
	assert.Equal(t, executionContext.ImageTag, persistedExecutionContext.ImageTag)
	assert.Equal(t, executionContext.ExecutionID, persistedExecutionContext.ExecutionID)
	assert.Equal(t, executionContext.Args, persistedExecutionContext.Args)
	assert.Equal(t, executionContext.Status, persistedExecutionContext.Status)

	_, err = postgresClient.db.Exec("DELETE FROM execution_context WHERE job_name='test-job-name'")
	assert.NoError(t, err)
}

func TestSelect(t *testing.T) {
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.PostgresDatabase(), config.PostgresUser(), config.PostgresPassword(), config.PostgresHost())

	db, err := sqlx.Connect("postgres", dataSourceName)
	assert.NoError(t, err)

	postgresClient := &client{db: db}
	defer postgresClient.db.Close()
	jobName := "test-job-name"

	snowflakeID, _ := id.NextID()
	executionContext := &executionContextModel.ExecutionContext{
		ExecutionID: snowflakeID,
		JobName:     jobName,
		ImageTag:    "test-image-name",
		Args:        map[string]string{"foo": "bar"},
		Status:      executionContextStatus.Finished,
	}

	_, err = postgresClient.NamedExec("INSERT INTO execution_context (id,job_name, image_tag, args, status) VALUES (:id, :job_name, :image_tag, :args, :status)", executionContext)
	assert.NoError(t, err)

	executionContextResult := []executionContextModel.ExecutionContext{}
	err = postgresClient.Select(&executionContextResult, "SELECT status from execution_context where job_name = $1", jobName)
	assert.NoError(t, err)

	assert.Equal(t, executionContext.Status, executionContextResult[0].Status)

	_, err = postgresClient.db.Exec("DELETE FROM execution_context WHERE job_name='test-job-name'")
	assert.NoError(t, err)
}

func TestSelectForNoRows(t *testing.T) {
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.PostgresDatabase(), config.PostgresUser(), config.PostgresPassword(), config.PostgresHost())

	db, err := sqlx.Connect("postgres", dataSourceName)
	assert.NoError(t, err)

	postgresClient := &client{db: db}
	defer postgresClient.db.Close()
	jobName := "test-job-name"

	executionContextResult := []executionContextModel.ExecutionContext{}
	err = postgresClient.db.Select(&executionContextResult, "SELECT status from execution_context where job_name = $1", jobName)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(executionContextResult))

	assert.NoError(t, err)
}
