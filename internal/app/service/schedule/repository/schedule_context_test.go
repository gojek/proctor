package repository

import (
	executionModel "proctor/internal/app/service/execution/model"
	executionRepository "proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/id"
	"proctor/internal/app/service/schedule/model"
	"testing"

	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	postgresqlClient           postgresql.Client
	repository                 ScheduleContextRepository
	executionContextRepository executionRepository.ExecutionContextRepository
	scheduleRepository         ScheduleRepository
}

func (context *testContext) setUp(t *testing.T) {
	context.postgresqlClient = postgresql.NewClient()
	context.repository = NewScheduleContextRepository(context.postgresqlClient)
	context.executionContextRepository = executionRepository.NewExecutionContextRepository(context.postgresqlClient)
	context.scheduleRepository = NewScheduleRepository(context.postgresqlClient)
	err := context.repository.deleteAll()
	assert.NoError(t, err)
	err = context.executionContextRepository.DeleteAll()
	assert.NoError(t, err)
	err = context.scheduleRepository.deleteAll()
	assert.NoError(t, err)
	fake.Seed(0)
}

func (context *testContext) tearDown() {
	context.postgresqlClient.Close()
}
func (context *testContext) instance() *testContext {
	return context
}

func newContext() context {
	ctx := &testContext{}
	return ctx
}

func TestScheduleContextRepository_InsertConstraintFailed(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	scheduleId := fake.Uint64()
	executionContextId := fake.Uint64()

	scheduleContext := model.ScheduleContext{
		ScheduleId:         scheduleId,
		ExecutionContextId: executionContextId,
	}

	updatedContext, err := ctx.instance().repository.Insert(scheduleContext)
	assert.Error(t, err)
	assert.Nil(t, updatedContext)

	ctx.tearDown()
}

func TestScheduleContextRepository_InsertSuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	schedule := model.Schedule{
		JobName:   fake.BuzzWord(),
		UserEmail: fake.Email(),
		Args: map[string]string{
			fake.FirstName(): fake.LastName(),
		},
		Cron:               "5 * * * *",
		Tags:               fake.BeerMalt(),
		NotificationEmails: fake.Email(),
		Group:              fake.HackerIngverb(),
		Enabled:            fake.Bool(),
	}

	scheduleId, err := ctx.instance().scheduleRepository.Insert(schedule)
	assert.NoError(t, err)
	assert.NotNil(t, scheduleId)

	fake.Seed(0)
	executionContext := executionModel.ExecutionContext{
		JobName:   fake.BuzzWord(),
		UserEmail: fake.Email(),
		ImageTag:  fake.BeerStyle(),
		Args: map[string]string{
			fake.FirstName(): fake.LastName(),
		},
		Status: status.Received,
	}

	executionContextId, err := ctx.instance().executionContextRepository.Insert(executionContext)
	assert.NoError(t, err)
	assert.NotNil(t, executionContextId)

	scheduleContext := model.ScheduleContext{
		ScheduleId:         scheduleId,
		ExecutionContextId: executionContextId,
	}

	updatedContext, err := ctx.instance().repository.Insert(scheduleContext)
	assert.NoError(t, err)
	assert.NotNil(t, updatedContext)

	expectedContext, err := ctx.instance().repository.GetByID(updatedContext.ID)
	assert.NoError(t, err)
	assert.NotNil(t, expectedContext)

	assert.Equal(t, updatedContext.ID, expectedContext.ID)
	assert.NotNil(t, expectedContext.CreatedAt)
	assert.NotNil(t, expectedContext.UpdatedAt)

	ctx.tearDown()
}

func TestScheduleContextRepository_GetContextAndSchedule(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	schedule := model.Schedule{
		JobName:   fake.BuzzWord(),
		UserEmail: fake.Email(),
		Args: map[string]string{
			fake.FirstName(): fake.LastName(),
		},
		Cron:               "5 * * * *",
		Tags:               fake.BeerMalt(),
		NotificationEmails: fake.Email(),
		Group:              fake.HackerIngverb(),
		Enabled:            fake.Bool(),
	}

	scheduleId, err := ctx.instance().scheduleRepository.Insert(schedule)
	assert.NoError(t, err)
	assert.NotNil(t, scheduleId)

	contextCount := 4
	for i := 1; i <= contextCount; i++ {
		fake.Seed(0)
		executionID, _ := id.NextID()
		executionContext := executionModel.ExecutionContext{
			ExecutionID: executionID,
			JobName:     fake.BuzzWord(),
			UserEmail:   fake.Email(),
			ImageTag:    fake.BeerStyle(),
			Args: map[string]string{
				fake.FirstName(): fake.LastName(),
			},
			Status: status.Received,
		}

		persistedExecutionID, err := ctx.instance().executionContextRepository.Insert(executionContext)
		assert.NoError(t, err)
		assert.NotNil(t, persistedExecutionID)

		scheduleContext := model.ScheduleContext{
			ScheduleId:         scheduleId,
			ExecutionContextId: persistedExecutionID,
		}

		updatedContext, err := ctx.instance().repository.Insert(scheduleContext)
		assert.NoError(t, err)
		assert.NotNil(t, updatedContext)
	}

	contexts, err := ctx.instance().repository.GetContextByScheduleId(scheduleId)
	assert.NoError(t, err)
	assert.NotEmpty(t, contexts)
	assert.Equal(t, contextCount, len(contexts))

	firstExecutionId := contexts[0].ExecutionID
	assert.NotNil(t, firstExecutionId)

	scheduleFromContext, err := ctx.instance().repository.GetScheduleByContextId(firstExecutionId)
	assert.NoError(t, err)
	assert.NotNil(t, scheduleFromContext)
	assert.Equal(t, scheduleId, scheduleFromContext.ID)

	ctx.tearDown()
}
