package repository

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/schedule/model"
	"strconv"
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
	schedule := model.Schedule{
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

	actualSchedule, err := suite.repository.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, actualSchedule)

	assert.Equal(t, id, actualSchedule.ID)
	assert.NotNil(t, actualSchedule.CreatedAt)
	assert.NotNil(t, actualSchedule.UpdatedAt)
	assert.Equal(t, mapValue, actualSchedule.Args[mapKey])
	assert.Equal(t, schedule.JobName, actualSchedule.JobName)
	assert.Equal(t, schedule.UserEmail, actualSchedule.UserEmail)
	assert.Equal(t, schedule.Cron, actualSchedule.Cron)
	assert.Equal(t, schedule.Tags, actualSchedule.Tags)
	assert.Equal(t, schedule.NotificationEmails, actualSchedule.NotificationEmails)
	assert.Equal(t, schedule.Group, actualSchedule.Group)
	assert.Equal(t, schedule.Enabled, actualSchedule.Enabled)
}

func (suite *ScheduleTestSuite) TestScheduleRepository_Delete() {
	t := suite.T()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()
	schedule := model.Schedule{
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

	expectedSchedule, err := suite.repository.GetByID(id)
	assert.Error(t, err)
	assert.Nil(t, expectedSchedule)
}

func (suite *ScheduleTestSuite) TestScheduleRepository_GetAll() {
	t := suite.T()
	recordCount := 15
	err := populateSeedDataForTest(suite.repository, recordCount, map[string]string{})
	assert.NoError(t, err)

	schedules, err := suite.repository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, schedules)
	size := len(schedules)
	assert.Equal(t, recordCount, size)
	for _, schedule := range schedules {
		assert.NotNil(t, schedule)
		assertScheduleCompleteParam(t, schedule)
	}
}

func (suite *ScheduleTestSuite) TestScheduleRepository_GetByUserEmail() {
	t := suite.T()
	recordCount := 15
	err := populateSeedDataForTest(suite.repository, recordCount, map[string]string{})
	assert.NoError(t, err)

	suppliedEmail := "bimo.horizon@gojek.co.id"
	withEmailCount := 14
	err = populateSeedDataForTest(suite.repository, withEmailCount, map[string]string{"UserEmail": suppliedEmail})

	schedules, err := suite.repository.GetByUserEmail(suppliedEmail)
	assert.NoError(t, err)
	assert.NotNil(t, schedules)
	size := len(schedules)
	assert.Equal(t, withEmailCount, size)

	for _, schedule := range schedules {
		assert.Equal(t, suppliedEmail, schedule.UserEmail)
		assertScheduleCompleteParam(t, schedule)
	}
}

func (suite *ScheduleTestSuite) TestScheduleRepository_GetByJobName() {
	t := suite.T()
	recordCount := 15
	err := populateSeedDataForTest(suite.repository, recordCount, map[string]string{})
	assert.NoError(t, err)

	suppliedJobName := "bimo-awesome-job"
	withJobName := 14
	err = populateSeedDataForTest(suite.repository, withJobName, map[string]string{"JobName": suppliedJobName})

	schedules, err := suite.repository.GetByJobName(suppliedJobName)
	assert.NoError(t, err)
	assert.NotNil(t, schedules)
	size := len(schedules)
	assert.Equal(t, withJobName, size)

	for _, schedule := range schedules {
		assert.Equal(t, suppliedJobName, schedule.JobName)
		assertScheduleCompleteParam(t, schedule)
	}
}

func (suite *ScheduleTestSuite) TestScheduleRepository_GetAllEnabled() {
	t := suite.T()
	recordCount := 15
	err := populateSeedDataForTest(suite.repository, recordCount, map[string]string{"Enabled": "false"})
	assert.NoError(t, err)

	withJobName := 14
	err = populateSeedDataForTest(suite.repository, withJobName, map[string]string{"Enabled": "true"})

	schedules, err := suite.repository.GetAllEnabled()
	assert.NoError(t, err)
	assert.NotNil(t, schedules)
	size := len(schedules)
	assert.Equal(t, withJobName, size)

	for _, schedule := range schedules {
		assert.True(t, schedule.Enabled)
		assertScheduleCompleteParam(t, schedule)
	}
}

func (suite *ScheduleTestSuite) TestScheduleRepository_GetEnabledByID() {
	t := suite.T()
	recordCount := 15
	err := populateSeedDataForTest(suite.repository, recordCount, map[string]string{})
	assert.NoError(t, err)

	mapKey := fake.FirstName()
	mapValue := fake.LastName()
	schedule := model.Schedule{
		JobName:   fake.BuzzWord(),
		UserEmail: fake.Email(),
		Args: map[string]string{
			mapKey: mapValue,
		},
		Cron:               "5 * * * *",
		Tags:               fake.BeerMalt(),
		NotificationEmails: fake.Email(),
		Group:              fake.HackerIngverb(),
		Enabled:            true,
	}

	id, err := suite.repository.Insert(schedule)
	assert.NotNil(t, id)
	assert.NoError(t, err)

	actualSchedule, err := suite.repository.GetEnabledByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, actualSchedule)
	assert.True(t, actualSchedule.Enabled)

	willNotExistsID := uint64(17777717)
	unexpectedSchedule, err := suite.repository.GetEnabledByID(willNotExistsID)
	assert.Error(t, err)
	assert.Nil(t, unexpectedSchedule)

	assert.Equal(t, id, actualSchedule.ID)
	assert.NotNil(t, actualSchedule.CreatedAt)
	assert.NotNil(t, actualSchedule.UpdatedAt)
	assert.Equal(t, mapValue, actualSchedule.Args[mapKey])
	assert.Equal(t, schedule.JobName, actualSchedule.JobName)
	assert.Equal(t, schedule.UserEmail, actualSchedule.UserEmail)
	assert.Equal(t, schedule.Cron, actualSchedule.Cron)
	assert.Equal(t, schedule.Tags, actualSchedule.Tags)
	assert.Equal(t, schedule.NotificationEmails, actualSchedule.NotificationEmails)
	assert.Equal(t, schedule.Group, actualSchedule.Group)
	assert.Equal(t, schedule.Enabled, actualSchedule.Enabled)
}

func (suite *ScheduleTestSuite) TestScheduleRepository_EnableDisable() {
	t := suite.T()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()
	schedule := model.Schedule{
		JobName:   fake.BuzzWord(),
		UserEmail: fake.Email(),
		Args: map[string]string{
			mapKey: mapValue,
		},
		Cron:               "5 * * * *",
		Tags:               fake.BeerMalt(),
		NotificationEmails: fake.Email(),
		Group:              fake.HackerIngverb(),
		Enabled:            true,
	}

	id, err := suite.repository.Insert(schedule)
	assert.NotNil(t, id)
	assert.NoError(t, err)

	expectedSchedule, err := suite.repository.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, expectedSchedule)
	assert.True(t, expectedSchedule.Enabled)

	err = suite.repository.Disable(id)
	assert.NoError(t, err)

	expectedSchedule, err = suite.repository.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, expectedSchedule)
	assert.False(t, expectedSchedule.Enabled)

	err = suite.repository.Enable(id)
	assert.NoError(t, err)

	expectedSchedule, err = suite.repository.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, expectedSchedule)
	assert.True(t, expectedSchedule.Enabled)

}

func assertScheduleCompleteParam(t *testing.T, schedule model.Schedule) {
	assert.NotNil(t, schedule)
	assert.NotNil(t, schedule.CreatedAt)
	assert.NotNil(t, schedule.UpdatedAt)
	assert.NotEmpty(t, schedule.Args)
	assert.NotEmpty(t, schedule.JobName)
	assert.NotEmpty(t, schedule.UserEmail)
	assert.NotEmpty(t, schedule.Cron)
	assert.NotEmpty(t, schedule.Tags)
	assert.NotEmpty(t, schedule.NotificationEmails)
	assert.NotEmpty(t, schedule.Group)
	assert.NotNil(t, schedule.Enabled)
}

func populateSeedDataForTest(repository ScheduleRepository, count int, seedField map[string]string) error {
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

		var enabled = fake.Bool()
		if val, ok := seedField["Enabled"]; ok {
			enabled, _ = strconv.ParseBool(val)
		}

		schedule := model.Schedule{
			JobName:   jobName,
			UserEmail: email,
			Args: map[string]string{
				fake.FirstName(): fake.LastName(),
			},
			Cron:               "5 * * * *",
			Tags:               fake.BeerMalt(),
			NotificationEmails: fake.Email(),
			Group:              fake.HackerIngverb(),
			Enabled:            enabled,
		}

		_, err := repository.Insert(schedule)

		if err != nil {
			return err
		}
	}
	return nil
}

func (suite *ScheduleTestSuite) TearDownSuite() {
	postgresqlClient.Close()
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleTestSuite))
}
