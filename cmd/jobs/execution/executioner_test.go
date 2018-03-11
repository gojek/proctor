package execution

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/gojekfarm/proctor/engine"
	"github.com/gojekfarm/proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExecutionCmdTestSuite struct {
	suite.Suite
	mockPrinter             *io.MockPrinter
	mockProctorEngineClient *engine.MockClient
	testExecutionCmd        *cobra.Command
}

func (s *ExecutionCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorEngineClient = &engine.MockClient{}
	s.testExecutionCmd = NewCmd(s.mockPrinter, s.mockProctorEngineClient)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdUsage() {
	assert.Equal(s.T(), "execute", s.testExecutionCmd.Use)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdHelp() {
	assert.Equal(s.T(), "Execute a job with arguments given", s.testExecutionCmd.Short)
	assert.Equal(s.T(), "Example: proctor job execute say-hello-world SAMPLE_ARG_ONE=any SAMPLE_ARG_TWO=variable", s.testExecutionCmd.Long)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmd() {
	args := []string{"say-hello-world", "SAMPLE_ARG_ONE=any", "SAMPLE_ARG_TWO=variable"}
	jobArgs := make(map[string]string)
	jobArgs["SAMPLE_ARG_ONE"] = "any"
	jobArgs["SAMPLE_ARG_TWO"] = "variable"

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Job", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With Variables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "SAMPLE_ARG_ONE", "any"), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "SAMPLE_ARG_TWO", "variable"), color.Reset).Once()

	s.mockProctorEngineClient.On("ExecuteJob", "say-hello-world", jobArgs).Return("executed-job-name", nil).Once()

	s.mockPrinter.On("Println", "Job execution successful. \nStreaming logs:", color.FgGreen).Once()

	s.mockProctorEngineClient.On("StreamJobLogs", "executed-job-name").Return(nil).Once()
	s.mockPrinter.On("Println", "\nLog stream of job completed.", color.FgGreen).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, args)

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForIncorrectUsage() {
	s.mockPrinter.On("Println", "Incorrect usage of proctor job execute", color.FgRed).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, []string{})

	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForNoJobVariables() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Job", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	jobArgs := make(map[string]string)
	s.mockProctorEngineClient.On("ExecuteJob", "say-hello-world", jobArgs).Return("executed-job-name", nil).Once()

	s.mockPrinter.On("Println", "Job execution successful. \nStreaming logs:", color.FgGreen).Once()

	s.mockProctorEngineClient.On("StreamJobLogs", "executed-job-name").Return(nil).Once()
	s.mockPrinter.On("Println", "\nLog stream of job completed.", color.FgGreen).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, args)

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForIncorrectVariableFormat() {
	args := []string{"say-hello-world", "incorrect-format"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Job", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With Variables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "\nIncorrect variable format\n", "incorrect-format"), color.FgRed).Once()

	jobArgs := make(map[string]string)
	s.mockProctorEngineClient.On("ExecuteJob", "say-hello-world", jobArgs).Return("executed-job-name", nil).Once()

	s.mockPrinter.On("Println", "Job execution successful. \nStreaming logs:", color.FgGreen).Once()

	s.mockProctorEngineClient.On("StreamJobLogs", "executed-job-name").Return(nil).Once()
	s.mockPrinter.On("Println", "\nLog stream of job completed.", color.FgGreen).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, args)

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForProctorEngineExecutionFailure() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Job", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	jobArgs := make(map[string]string)
	s.mockProctorEngineClient.On("ExecuteJob", "say-hello-world", jobArgs).Return("", errors.New("error")).Once()

	s.mockPrinter.On("Println", "\nError executing job. Please check configuration and network connectivity", color.FgRed).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, args)

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForProctorEngineLogStreamingFailure() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Job", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	jobArgs := make(map[string]string)
	s.mockProctorEngineClient.On("ExecuteJob", "say-hello-world", jobArgs).Return("executed-job-name", nil).Once()

	s.mockPrinter.On("Println", "Job execution successful. \nStreaming logs:", color.FgGreen).Once()

	s.mockProctorEngineClient.On("StreamJobLogs", "executed-job-name").Return(errors.New("error")).Once()
	s.mockPrinter.On("Println", "\nError Streaming Logs", color.FgRed).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, args)

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestExecutionCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionCmdTestSuite))
}
