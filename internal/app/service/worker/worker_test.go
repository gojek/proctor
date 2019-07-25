package worker

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	executionContextModel "proctor/internal/app/service/execution/model"
	executionContextRepository "proctor/internal/app/service/execution/repository"
	executionService "proctor/internal/app/service/execution/service"
	"proctor/internal/app/service/infra/mail"
	scheduleModel "proctor/internal/app/service/schedule/model"
	scheduleRepository "proctor/internal/app/service/schedule/repository"
	"proctor/internal/pkg/constant"
)

type WorkerTestSuite struct {
	suite.Suite
	mockExecutionService           *executionService.MockExecutionService
	mockExecutionContextRepository *executionContextRepository.MockExecutionContextRepository
	mockScheduleRepository         *scheduleRepository.MockScheduleRepository
	mockMailer                     *mail.MockMailer
	worker                         Worker
}

func (suite *WorkerTestSuite) SetupTest() {
	suite.mockExecutionService = &executionService.MockExecutionService{}
	suite.mockExecutionContextRepository = &executionContextRepository.MockExecutionContextRepository{}
	suite.mockScheduleRepository = &scheduleRepository.MockScheduleRepository{}
	suite.mockMailer = &mail.MockMailer{}
	suite.worker = NewWorker(
		suite.mockExecutionService,
		suite.mockExecutionContextRepository,
		suite.mockScheduleRepository,
		suite.mockMailer,
	)
}

func (suite *WorkerTestSuite) TestEnableRun() {
	t := suite.T()

	jobArgs := map[string]string{"abc": "def"}
	enabledJob := "some-job-one"
	disabledJob := "some-job-two"
	userEmail := "foo@bar.com"
	notificationEmails := "foo@bar.com,goo@bar.com"
	executionContext := executionContextModel.ExecutionContext{
		ExecutionID: uint64(1),
		JobName:     enabledJob,
		Name:        "test",
		UserEmail:   constant.WorkerEmail,
		ImageTag:    "test",
		Args:        jobArgs,
	}
	scheduledJobs := []scheduleModel.Schedule{
		{
			ID:                 uint64(1),
			Enabled:            true,
			Cron:               "*/1 * * * * *",
			JobName:            enabledJob,
			Args:               jobArgs,
			NotificationEmails: notificationEmails,
			UserEmail:          userEmail,
			Tags:               "test",
		},
		{
			ID:                 uint64(2),
			Enabled:            false,
			Cron:               "*/1 * * * * *",
			JobName:            disabledJob,
			Args:               jobArgs,
			NotificationEmails: notificationEmails,
			UserEmail:          userEmail,
			Tags:               "test",
		},
	}

	tickerChan := make(chan time.Time)
	signalsChan := make(chan os.Signal, 1)
	scheduledJobExecutedChan := make(chan bool)

	suite.mockScheduleRepository.On("GetAll").Return(scheduledJobs, nil)

	suite.mockExecutionService.On("Execute", enabledJob, constant.WorkerEmail, jobArgs).Return(&executionContext, "test", nil)
	defer suite.mockExecutionService.AssertExpectations(t)
	defer suite.mockExecutionService.AssertNotCalled(t, "Execute", disabledJob, constant.WorkerEmail, jobArgs)
	suite.mockMailer.On("Send", executionContext, scheduledJobs[0]).Return(nil).Run(
		func(args mock.Arguments) {
			scheduledJobExecutedChan <- true
		},
	)
	defer suite.mockMailer.AssertExpectations(t)

	go suite.worker.Run(tickerChan, signalsChan)

	tickerChan <- time.Now()

	<-scheduledJobExecutedChan
	signalsChan <- syscall.SIGTERM
}

func (suite *WorkerTestSuite) TestDisableRun() {
	t := suite.T()

	jobArgs := map[string]string{"abc": "def"}
	enabledJob := "some-job-one"
	userEmail := "foo@bar.com"
	notificationEmails := "foo@bar.com,goo@bar.com"
	executionContext := executionContextModel.ExecutionContext{
		ExecutionID: uint64(1),
		JobName:     enabledJob,
		Name:        "test",
		UserEmail:   constant.WorkerEmail,
		ImageTag:    "test",
		Args:        jobArgs,
	}
	disableExecutionContext := executionContext
	enabledScheduledJobs := []scheduleModel.Schedule{
		{
			ID:                 uint64(1),
			Enabled:            true,
			Cron:               "*/1 * * * * *",
			JobName:            enabledJob,
			Args:               jobArgs,
			NotificationEmails: notificationEmails,
			UserEmail:          userEmail,
			Tags:               "test",
		},
	}
	disabledScheduledJobs := []scheduleModel.Schedule{
		{
			ID:                 uint64(1),
			Enabled:            false,
			Cron:               "*/1 * * * * *",
			JobName:            enabledJob,
			Args:               jobArgs,
			NotificationEmails: notificationEmails,
			UserEmail:          userEmail,
			Tags:               "test",
		},
	}

	tickerChan := make(chan time.Time)
	signalsChan := make(chan os.Signal, 1)
	toggledOffEnabledJobChan := make(chan bool)

	suite.mockScheduleRepository.On("GetAll").Return(enabledScheduledJobs, nil).Once().Run(
		func(args mock.Arguments) {
			toggledOffEnabledJobChan <- true
		},
	)
	suite.mockScheduleRepository.On("GetAll").Return(disabledScheduledJobs, nil).Run(
		func(args mock.Arguments) {
			toggledOffEnabledJobChan <- true
		},
	)
	defer suite.mockScheduleRepository.AssertExpectations(t)

	suite.mockExecutionService.On("Execute", enabledJob, constant.WorkerEmail, jobArgs).Return(&disableExecutionContext, "test", nil)
	defer suite.mockExecutionService.AssertExpectations(t)

	suite.mockMailer.On("Send", disableExecutionContext, enabledScheduledJobs[0]).Return(nil).Run(
		func(args mock.Arguments) {
			toggledOffEnabledJobChan <- true
		},
	)
	defer suite.mockMailer.AssertExpectations(t)

	go suite.worker.Run(tickerChan, signalsChan)

	tickerChan <- time.Now()

	<-toggledOffEnabledJobChan

	//Wait for 2 seconds to ensure disabled job isn't executed again
	time.Sleep(2 * time.Second)
	signalsChan <- syscall.SIGTERM
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}
