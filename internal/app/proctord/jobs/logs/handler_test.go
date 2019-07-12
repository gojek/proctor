package logs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"proctor/internal/app/service/infra/kubernetes"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"proctor/internal/pkg/constant"
	"proctor/internal/pkg/utility"
)

type LoggerTestSuite struct {
	suite.Suite
	testLogger     Logger
	mockKubeClient *kubernetes.MockKubernetesClient
}

func (suite *LoggerTestSuite) SetupTest() {
	suite.mockKubeClient = &kubernetes.MockKubernetesClient{}
	suite.testLogger = NewLogger(suite.mockKubeClient)
}

type logsHandlerServer struct {
	*httptest.Server
}

var logsHandlerDialer = websocket.Dialer{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	logsHandlerPath       = "/jobs/logs"
	logsHandlerRawQuery   = "job_name=sample"
	logsHandlerRequestURI = logsHandlerPath
)

func (suite *LoggerTestSuite) newServer() *logsHandlerServer {
	var s logsHandlerServer
	s.Server = httptest.NewServer(suite.testLogger.Stream())
	s.Server.URL += logsHandlerRequestURI
	s.URL = makeWsProto(s.Server.URL)
	return &s
}

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

func (suite *LoggerTestSuite) TestLoggerStream() {
	t := suite.T()

	s := suite.newServer()
	defer s.Close()

	buffer := utility.NewBuffer()
	_, _ = buffer.Write([]byte("first line\nsecond line\n"))
	suite.mockKubeClient.On("StreamJobLogs", "sample").Return(buffer, nil).Once()

	c, _, err := websocket.DefaultDialer.Dial(s.URL+"?"+logsHandlerRawQuery, nil)
	assert.NoError(t, err)
	defer c.Close()

	_, firstMessage, err := c.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "first line", string(firstMessage))

	_, secondMessage, err := c.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "second line", string(secondMessage))

	_, finalMessage, err := c.ReadMessage()
	assert.Error(t, err)
	assert.Equal(t, "", string(finalMessage))
	assert.Equal(t, "websocket: close 1000 (normal): All logs are read", err.Error())

	suite.mockKubeClient.AssertExpectations(t)
	assert.True(t, buffer.WasClosed())
}

func (suite *LoggerTestSuite) TestLoggerStreamConnectionUpgradeFailure() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/jobs/logs", &utility.Buffer{})
	responseRecorder := httptest.NewRecorder()

	suite.testLogger.Stream()(responseRecorder, req)

	suite.mockKubeClient.AssertNotCalled(t, "StreamJobLogs", mock.Anything)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "Bad Request\n"+constant.ClientError, responseRecorder.Body.String())
}

func (suite *LoggerTestSuite) TestLoggerStreamForNoJobName() {
	t := suite.T()

	s := suite.newServer()
	defer s.Close()

	c, _, err := websocket.DefaultDialer.Dial(s.URL, nil)
	assert.NoError(t, err)
	defer c.Close()

	suite.mockKubeClient.AssertNotCalled(t, "StreamJobLogs", mock.Anything)

	_, finalMessage, err := c.ReadMessage()
	assert.Error(t, err)
	assert.Equal(t, "", string(finalMessage))
	assert.Equal(t, "websocket: close 1000 (normal): No job name provided while requesting for logs", err.Error())
}

func (suite *LoggerTestSuite) TestLoggerStreamKubeClientFailure() {
	t := suite.T()

	s := suite.newServer()
	defer s.Close()

	suite.mockKubeClient.On("StreamJobLogs", "sample").Return(&utility.Buffer{}, errors.New("error")).Once()

	c, _, err := websocket.DefaultDialer.Dial(s.URL+"?"+logsHandlerRawQuery, nil)
	assert.NoError(t, err)
	defer c.Close()

	_, finalMessage, err := c.ReadMessage()
	assert.Error(t, err)
	assert.Equal(t, "", string(finalMessage))
	assert.Equal(t, "websocket: close 1000 (normal): Something went wrong", err.Error())

	suite.mockKubeClient.AssertExpectations(t)
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}
