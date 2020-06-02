package repository

import (
	"testing"

	fake "github.com/brianvoe/gofakeit"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/id"
)

func TestExecutionContextRepository_Insert(t *testing.T) {
	postgresqlClient := postgresql.NewClient()
	defer postgresqlClient.Close()
	repository := NewExecutionContextRepository(postgresqlClient)
	defer repository.DeleteAll()

	fake.Seed(0)
	mapKey := fake.FirstName()
	mapValue := fake.LastName()

	executionID, _ := id.NextID()
	context := model.ExecutionContext{
		ExecutionID: executionID,
		JobName:     fake.BuzzWord(),
		UserEmail:   fake.Email(),
		ImageTag:    fake.BeerStyle(),
		Args: map[string]string{
			mapKey: mapValue,
		},
		Status: status.Received,
	}

	id, err := repository.Insert(context)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	expectedContext, err := repository.GetById(id)
	assert.Nil(t, err)
	assert.NotNil(t, expectedContext)
	assert.Equal(t, id, expectedContext.ExecutionID)
	assert.NotNil(t, expectedContext.CreatedAt)
	assert.NotNil(t, expectedContext.UpdatedAt)
	assert.Equal(t, expectedContext.Args[mapKey], mapValue)
}

func TestExecutionContextRepository_Delete(t *testing.T) {
	postgresqlClient := postgresql.NewClient()
	defer postgresqlClient.Close()
	repository := NewExecutionContextRepository(postgresqlClient)
	defer repository.DeleteAll()

	fake.Seed(0)
	executionID, _ := id.NextID()
	context := model.ExecutionContext{
		ExecutionID: executionID,
		JobName:     fake.BuzzWord(),
		UserEmail:   fake.Email(),
		ImageTag:    fake.BeerStyle(),
		Args: map[string]string{
			fake.FirstName(): fake.LastName(),
		},
		Status: status.Received,
	}

	id, err := repository.Insert(context)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	err = repository.Delete(id)
	assert.Nil(t, err)

	expectedContext, err := repository.GetById(id)
	assert.NotNil(t, err)
	assert.Nil(t, expectedContext)
}

func TestExecutionContextRepository_UpdateStatus(t *testing.T) {
	postgresqlClient := postgresql.NewClient()
	defer postgresqlClient.Close()
	repository := NewExecutionContextRepository(postgresqlClient)
	defer repository.DeleteAll()

	fake.Seed(0)
	executionID, _ := id.NextID()
	context := model.ExecutionContext{
		ExecutionID: executionID,
		JobName:     fake.BuzzWord(),
		UserEmail:   fake.Email(),
		ImageTag:    fake.BeerStyle(),
		Args: map[string]string{
			fake.FirstName(): fake.LastName(),
		},
		Status: status.Received,
	}

	id, err := repository.Insert(context)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	newStatus := status.Created
	err = repository.UpdateStatus(id, newStatus)
	assert.Nil(t, err)

	expectedContext, err := repository.GetById(id)
	assert.NotNil(t, expectedContext)
	assert.Nil(t, err)
	assert.Equal(t, newStatus, expectedContext.Status)
}

func TestExecutionContextRepository_UpdateJobOutput(t *testing.T) {
	postgresqlClient := postgresql.NewClient()
	defer postgresqlClient.Close()
	repository := NewExecutionContextRepository(postgresqlClient)
	defer repository.DeleteAll()

	fake.Seed(0)
	executionID, _ := id.NextID()
	context := model.ExecutionContext{
		ExecutionID: executionID,
		JobName:     fake.BuzzWord(),
		UserEmail:   fake.Email(),
		ImageTag:    fake.BeerStyle(),
		Args: map[string]string{
			fake.FirstName(): fake.LastName(),
		},
		Status: status.Received,
	}

	id, err := repository.Insert(context)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	newLog := `
        This ain't a log for the broken-hearted
		No silent prayer for the faith-departed
		I ain't gonna be just a face in the crowd
		You're gonna hear my voice
		When I shout it out loud
		It's my log
		It's now or never
		I ain't gonna log forever
		I just want to log while I'm alive
		(It's my log)
		My heart is like an open highway
		Like Frankie said
		I did it my way
		I just want to log while I'm alive
		It's my log
	`

	newOutput := types.GzippedText(newLog)

	err = repository.UpdateJobOutput(id, newOutput)
	assert.Nil(t, err)

	expectedContext, err := repository.GetById(id)
	assert.NotNil(t, expectedContext)
	assert.Nil(t, err)
	expectedLog := string(expectedContext.Output)
	assert.Equal(t, newLog, expectedLog)
}

func populateSeedDataForTest(repository ExecutionContextRepository, count int, seedField map[string]string) error {
	for i := 0; i < count; i++ {
		fake.Seed(0)
		var jobName = fake.BuzzWord()
		if val, ok := seedField["JobName"]; ok {
			jobName = val
		}

		var email = fake.Email()
		if val, ok := seedField["UserEmail"]; ok {
			email = val
		}

		var defaultStatus = status.RequirementNotMet
		if val, ok := seedField["Status"]; ok {
			defaultStatus = status.ExecutionStatus(val)
		}

		executionID, _ := id.NextID()
		context := model.ExecutionContext{
			ExecutionID: executionID,
			JobName:     jobName,
			Name:        fake.HackerAdjective(),
			UserEmail:   email,
			ImageTag:    fake.BeerStyle(),
			Args: map[string]string{
				fake.FirstName(): fake.LastName(),
			},
			Status: defaultStatus,
		}

		_, err := repository.Insert(context)

		if err != nil {
			return err
		}
	}
	return nil
}

func TestExecutionContextRepository_GetByEmail(t *testing.T) {
	postgresqlClient := postgresql.NewClient()
	defer postgresqlClient.Close()
	repository := NewExecutionContextRepository(postgresqlClient)
	defer repository.DeleteAll()

	recordCount := 15
	userEmail := "bimo.horizon@go-pay.co.id"
	err := populateSeedDataForTest(repository, recordCount, map[string]string{"UserEmail": userEmail})
	assert.Nil(t, err)

	contexts, err := repository.GetByEmail(userEmail)
	assert.Nil(t, err)
	assert.NotEmpty(t, contexts)
	assert.Equal(t, recordCount, len(contexts))
}

func TestExecutionContextRepository_GetByJobName(t *testing.T) {
	postgresqlClient := postgresql.NewClient()
	defer postgresqlClient.Close()
	repository := NewExecutionContextRepository(postgresqlClient)
	defer repository.DeleteAll()

	recordCount := 15
	jobName := "some_job_that_only_exists_in_your_past"
	err := populateSeedDataForTest(repository, recordCount, map[string]string{"JobName": jobName})
	assert.Nil(t, err)

	contexts, err := repository.GetByJobName(jobName)
	assert.Nil(t, err)
	assert.NotEmpty(t, contexts)
	assert.Equal(t, recordCount, len(contexts))
}

func TestExecutionContextRepository_GetByStatus(t *testing.T) {
	postgresqlClient := postgresql.NewClient()
	defer postgresqlClient.Close()
	repository := NewExecutionContextRepository(postgresqlClient)
	defer repository.DeleteAll()

	recordCount := 15
	status := "well_execution_status_here_must_be_cool"
	err := populateSeedDataForTest(repository, recordCount, map[string]string{"Status": status})
	assert.Nil(t, err)

	contexts, err := repository.GetByStatus(status)
	assert.Nil(t, err)
	assert.NotEmpty(t, contexts)
	assert.Equal(t, recordCount, len(contexts))
}
