package status

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"proctor/internal/app/cli/daemon"
	"proctor/internal/pkg/constant"
	"proctor/internal/pkg/io"
	modelExecution "proctor/internal/pkg/model/execution"
)

type StatusCmdTestSuite struct {
	suite.Suite
	mockPrinter        *io.MockPrinter
	mockProctorDClient *daemon.MockClient
	testStatusCmd      *cobra.Command
}

func (s *StatusCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testStatusCmd = NewCmd(s.mockPrinter, s.mockProctorDClient, func(exitCode int) {})
}

func (s *StatusCmdTestSuite) TestStatusCmdUsage() {
	assert.Equal(s.T(), "status", s.testStatusCmd.Use)
}

func (s *StatusCmdTestSuite) TestStatusCmdHelp() {
	assert.Equal(s.T(), "Get status of an execution context", s.testStatusCmd.Short)
	assert.Equal(s.T(), "To get status of an execution context, this command retrieve status from previous execution", s.testStatusCmd.Long)
	assert.Equal(s.T(), "proctor status 123", s.testStatusCmd.Example)
}

func (s *StatusCmdTestSuite) TestStatusCmd() {
	executionID := uint64(42)

	s.mockPrinter.On("Println", "Getting status", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionID), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Job Name", "foo"), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Status", constant.JobSucceeded), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Updated At", ""), color.FgGreen).Once()

	executionResult := &modelExecution.ExecutionResult{
		ExecutionId:   uint64(0),
		JobName:       "foo",
		ExecutionName: "",
		ImageTag:      "",
		CreatedAt:     "",
		UpdatedAt:     "",
		Status:        constant.JobSucceeded,
	}
	s.mockProctorDClient.On("GetExecutionContextStatus", executionID).Return(executionResult, nil).Once()
	s.mockPrinter.On("Println", "Execution completed.", color.FgGreen).Once()

	s.testStatusCmd.Run(&cobra.Command{}, []string{"42"})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *StatusCmdTestSuite) TestStatusCmdInvalidExecutionIDError() {
	t := s.T()

	s.mockPrinter.On("Println", "No valid execution context id provided as argument", color.FgRed).Once()

	s.testStatusCmd.Run(&cobra.Command{}, []string{"foo"})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
	s.mockPrinter.AssertNotCalled(t, "Println", "Execution completed.", color.FgGreen)
}

func (s *StatusCmdTestSuite) TestStatusCmdGetExecutionStatusError() {
	t := s.T()

	executionID := uint64(42)

	s.mockPrinter.On("Println", "Getting status", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionID), color.FgGreen).Once()

	s.mockProctorDClient.On("GetExecutionContextStatus", executionID).Return(&modelExecution.ExecutionResult{}, errors.New("test")).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Error while Getting Status:", "test"), color.FgRed).Once()
	s.testStatusCmd.Run(&cobra.Command{}, []string{"42"})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
	s.mockPrinter.AssertNotCalled(t, "Println", "Execution completed.", color.FgGreen)
	s.mockPrinter.AssertNotCalled(t, "Println", fmt.Sprintf("%-40s %-100v", "Status", constant.JobSucceeded), color.FgGreen)
}

func TestStatusCmdTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCmdTestSuite))
}
