package repository

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/schedule/model"
	"testing"
)

type ScheduleTestSuite struct {
	suite.Suite
	repository ScheduleRepository
}

var postgresqlClient postgresql.Client

func (suite *ScheduleTestSuite) SetupSuite() {
	postgresqlClient = postgresql.NewClient()
}

func (suite *ScheduleTestSuite) SetupTest() {
	t := suite.T()
	suite.repository = NewScheduleRepository(postgresqlClient)
	err := suite.repository.deleteAll()
	assert.NoError(t, err)
	fake.Seed(0)
}

func (suite *ScheduleTestSuite) TestScheduleRepository_Insert() {
	t := suite.T()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()
	schedule := &model.Schedule{
		JobName:   fake.BuzzWord(),
		UserEmail: fake.Email(),
		Args: map[string]string{
			mapKey: mapValue,
		},
		Cron:               "5 * * * *",
		Tags:               fake.BeerMalt(),
		NotificationEmails: fake.Email(),
		Group:              fake.HackerIngverb(),
		Enabled:            fake.Bool(),
	}

	id, err := suite.repository.Insert(schedule)
	assert.NotNil(t, id)
	assert.NoError(t, err)

	expectedSchedule, err := suite.repository.GetById(id)
	assert.NoError(t, err)
	assert.NotNil(t, expectedSchedule)

	assert.Equal(t, id, expectedSchedule.ID)
	assert.NotNil(t, expectedSchedule.CreatedAt)
	assert.NotNil(t, expectedSchedule.UpdatedAt)
	assert.Equal(t, expectedSchedule.Args[mapKey], mapValue)
}

func (suite *ScheduleTestSuite) TestScheduleRepository_Delete() {
	t := suite.T()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()
	schedule := &model.Schedule{
		JobName:   fake.BuzzWord(),
		UserEmail: fake.Email(),
		Args: map[string]string{
			mapKey: mapValue,
		},
		Cron:               "5 * * * *",
		Tags:               fake.BeerMalt(),
		NotificationEmails: fake.Email(),
		Group:              fake.HackerIngverb(),
		Enabled:            fake.Bool(),
	}

	id, err := suite.repository.Insert(schedule)
	assert.NotNil(t, id)
	assert.NoError(t, err)

	err = suite.repository.Delete(id)
	assert.NoError(t, err)

	expectedSchedule, err := suite.repository.GetById(id)
	assert.Error(t, err)
	assert.Nil(t, expectedSchedule)
}

func (suite *ScheduleTestSuite) TearDownSuite() {
	postgresqlClient.Close()
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleTestSuite))
}
