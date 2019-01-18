package schedule

import (
	"bytes"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"encoding/json"
)

type GetScheduledJobsTestSuite struct {
	suite.Suite
	mockStore         *storage.MockStore

	testHandler Handler
}

func (suite *GetScheduledJobsTestSuite) SetupTest() {
	suite.mockStore = &storage.MockStore{}

	suite.testHandler = NewGetHandler(suite.mockStore)
}

func (s *GetScheduledJobsTestSuite) TestGetScheduledJobs() {
	t := s.T()

	req := httptest.NewRequest("GET", "/jobs/schedule", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	scheduledJobs := []postgres.JobsSchedule{}
	s.mockStore.On("GetScheduledJobs").Return(scheduledJobs, nil).Once()

	s.testHandler.GetScheduledJobs()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	expectedJobDetails, err := json.Marshal(scheduledJobs)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobDetails, responseRecorder.Body.Bytes())
}

func TestGetScheduledJobsTestSuite(t *testing.T) {
	suite.Run(t, new(GetScheduledJobsTestSuite))
}