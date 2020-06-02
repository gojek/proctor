package log

import (
	"errors"
	"fmt"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LogCmdTestSuite struct {
	suite.Suite
	mockPrinter        *io.MockPrinter
	mockProctorDClient *daemon.MockClient
	testLogCmd         *cobra.Command
}

func (s *LogCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testLogCmd = NewCmd(s.mockPrinter, s.mockProctorDClient, func(exitCode int) {})
}

func (s *LogCmdTestSuite) TestLogCmdUsage() {
	assert.Equal(s.T(), "logs", s.testLogCmd.Use)
}

func (s *LogCmdTestSuite) TestLogCmdHelp() {
	assert.Equal(s.T(), "Get logs of an execution context", s.testLogCmd.Short)
	assert.Equal(s.T(), "To get a log of execution context, this command helps retrieve logs from previous execution", s.testLogCmd.Long)
	assert.Equal(s.T(), "proctor logs 123", s.testLogCmd.Example)
}

func (s *LogCmdTestSuite) TestLogCmd() {
	executionID := uint64(42)

	s.mockPrinter.On("Println", "Getting logs", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionID), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionID).Return(nil).Once()
	s.mockPrinter.On("Println", "Execution completed.", color.FgGreen).Once()

	s.testLogCmd.Run(&cobra.Command{}, []string{"42"})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *LogCmdTestSuite) TestLogCmdInvalidExecutionIDError() {
	t := s.T()

	s.mockPrinter.On("Println", "No valid execution context id provided as argument", color.FgRed).Once()

	s.testLogCmd.Run(&cobra.Command{}, []string{"foo"})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
	s.mockPrinter.AssertNotCalled(t, "Println", "Execution completed.", color.FgGreen)
}

func (s *LogCmdTestSuite) TestLogCmdInvalidStreamProcLogsError() {
	t := s.T()

	executionID := uint64(42)

	s.mockPrinter.On("Println", "Getting logs", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionID), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionID).Return(errors.New("test")).Once()
	s.mockPrinter.On("Println", "Error while Streaming Log.", color.FgRed).Once()
	s.testLogCmd.Run(&cobra.Command{}, []string{"42"})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
	s.mockPrinter.AssertNotCalled(t, "Println", "Execution completed.", color.FgGreen)
}

func TestLogCmdTestSuite(t *testing.T) {
	suite.Run(t, new(LogCmdTestSuite))
}
