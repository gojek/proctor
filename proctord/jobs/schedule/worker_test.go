package schedule

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gojektech/proctor/proctord/audit"
	"github.com/gojektech/proctor/proctord/jobs/execution"
	"github.com/gojektech/proctor/proctord/mail"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WorkerTestSuite struct {
	suite.Suite
	mockStore       *storage.MockStore
	mockExecutioner *execution.MockExecutioner
	mockAuditor     *audit.MockAuditor
	mockMailer      *mail.MockMailer
	testWorker      Worker
}

func (suite *WorkerTestSuite) SetupTest() {
	suite.mockStore = &storage.MockStore{}
	suite.mockExecutioner = &execution.MockExecutioner{}
	suite.mockAuditor = &audit.MockAuditor{}
	suite.mockMailer = &mail.MockMailer{}

	suite.testWorker = NewWorker(suite.mockStore, suite.mockExecutioner, suite.mockAuditor, suite.mockMailer)
}

func (suite *WorkerTestSuite) TestCronEnablingForScheduledJobs() {
	t := suite.T()

	jobArgs := map[string]string{"abc": "def"}
	jsonEncodedJobArgs, err := json.Marshal(jobArgs)
	assert.NoError(t, err)

	enabledJob := "some-job-one"
	disabledJob := "some-job-two"
	notificationEmails := "foo@bar.com,goo@bar.com"
	scheduledJobs := []postgres.JobsSchedule{
		postgres.JobsSchedule{
			ID:                 "some-uuid-one",
			Enabled:            true,
			Time:               "*/1 * * * * *",
			Name:               enabledJob,
			Args:               base64.StdEncoding.EncodeToString(jsonEncodedJobArgs),
			NotificationEmails: notificationEmails,
		},
		postgres.JobsSchedule{
			ID:                 "some-uuid-two",
			Enabled:            false,
			Time:               "*/1 * * * * *",
			Name:               disabledJob,
			Args:               base64.StdEncoding.EncodeToString(jsonEncodedJobArgs),
			NotificationEmails: notificationEmails,
		},
	}

	tickerChan := make(chan time.Time)
	signalsChan := make(chan os.Signal, 1)
	scheduledJobExecutedChan := make(chan bool)

	suite.mockStore.On("GetScheduledJobs").Return(scheduledJobs, nil)

	jobExecutionID := "job-execution-id"
	suite.mockExecutioner.On("Execute", mock.Anything, enabledJob, jobArgs).Return(jobExecutionID, nil)

	jobExecutionStatus := utility.JobSucceeded
	suite.mockAuditor.On("JobsExecution", mock.Anything).Return()
	suite.mockAuditor.On("JobsExecutionStatus", jobExecutionID).Return(jobExecutionStatus, nil)

	expectedRecipients := strings.Split(notificationEmails, ",")
	suite.mockMailer.On("Send", enabledJob, jobExecutionID, jobExecutionStatus, jobArgs, expectedRecipients).Return(nil).Run(
		func(args mock.Arguments) {
			scheduledJobExecutedChan <- true
		},
	)

	go suite.testWorker.Run(tickerChan, signalsChan)

	tickerChan <- time.Now()

	<-scheduledJobExecutedChan
	signalsChan <- syscall.SIGTERM
	suite.mockExecutioner.AssertExpectations(t)
	suite.mockAuditor.AssertExpectations(t)
	suite.mockMailer.AssertExpectations(t)
	suite.mockExecutioner.AssertNotCalled(t, "Execute", disabledJob, jobArgs)
}

func (suite *WorkerTestSuite) TestCronForDisablingEnabledScheduledJobs() {
	t := suite.T()

	jobArgs := map[string]string{"abc": "def"}
	jsonEncodedJobArgs, err := json.Marshal(jobArgs)
	assert.NoError(t, err)

	jobName := "some-job-one"
	notificationEmails := "foo@bar.com,goo@bar.com"
	scheduledJobs := []postgres.JobsSchedule{
		postgres.JobsSchedule{
			ID:                 "some-uuid-one",
			Enabled:            true,
			Time:               "*/1 * * * * *",
			Name:               jobName,
			Args:               base64.StdEncoding.EncodeToString(jsonEncodedJobArgs),
			NotificationEmails: notificationEmails,
		},
	}

	disabledScheduledJobs := []postgres.JobsSchedule{
		postgres.JobsSchedule{
			ID:      "some-uuid-one",
			Enabled: false,
			Time:    "*/1 * * * * *",
			Name:    jobName,
			Args:    base64.StdEncoding.EncodeToString(jsonEncodedJobArgs),
		},
	}

	tickerChan := make(chan time.Time)
	signalsChan := make(chan os.Signal, 1)
	toggledOffEnabledJobChan := make(chan bool)

	suite.mockStore.On("GetScheduledJobs").Return(scheduledJobs, nil).Once()
	suite.mockStore.On("GetScheduledJobs").Return(disabledScheduledJobs, nil).Run(
		func(args mock.Arguments) {
			toggledOffEnabledJobChan <- true
		},
	)

	jobExecutionID := "job-execution-id"
	suite.mockExecutioner.On("Execute", mock.Anything, jobName, jobArgs).Return(jobExecutionID, nil)

	suite.mockAuditor.On("JobsExecution", mock.Anything).Return()
	jobExecutionStatus := utility.JobSucceeded
	suite.mockAuditor.On("JobsExecutionStatus", jobExecutionID).Return(jobExecutionStatus, nil)

	expectedRecipients := strings.Split(notificationEmails, ",")
	suite.mockMailer.On("Send", jobName, jobExecutionID, jobExecutionStatus, jobArgs, expectedRecipients).Return(nil).Run(
		func(args mock.Arguments) {
			toggledOffEnabledJobChan <- true
		},
	)

	go suite.testWorker.Run(tickerChan, signalsChan)

	tickerChan <- time.Now()

	<-toggledOffEnabledJobChan
	suite.mockExecutioner.AssertExpectations(t)
	suite.mockAuditor.AssertExpectations(t)
	suite.mockMailer.AssertExpectations(t)

	//Wait for 2 seconds to ensure disabled job isn't executed again
	time.Sleep(2 * time.Second)
	signalsChan <- syscall.SIGTERM
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}
