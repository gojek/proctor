package execution

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
	"proctor/internal/pkg/model/execution"
)

type ExecutionCmdTestSuite struct {
	suite.Suite
	mockPrinter        *io.MockPrinter
	mockProctorDClient *daemon.MockClient
	testExecutionCmd   *cobra.Command
}

func (s *ExecutionCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testExecutionCmd = NewCmd(s.mockPrinter, s.mockProctorDClient, func(exitCode int) {})
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdUsage() {
	assert.Equal(s.T(), "execute", s.testExecutionCmd.Use)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdHelp() {
	assert.Equal(s.T(), "Execute a proc with given arguments", s.testExecutionCmd.Short)
	assert.Equal(s.T(), "To execute a proc, this command helps to communicate with `proctord` and streams to logs of proc in execution", s.testExecutionCmd.Long)
	assert.Equal(s.T(), "proctor execute proc-one SOME_VAR=foo ANOTHER_VAR=bar\nproctor execute proc-two ANY_VAR=baz", s.testExecutionCmd.Example)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmd() {
	args := []string{"say-hello-world", "SAMPLE_ARG_ONE=any", "SAMPLE_ARG_TWO=variable"}
	procArgs := make(map[string]string)
	procArgs["SAMPLE_ARG_ONE"] = "any"
	procArgs["SAMPLE_ARG_TWO"] = "variable"

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With Variables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "SAMPLE_ARG_ONE", "any"), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "SAMPLE_ARG_TWO", "variable"), color.Reset).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId:   uint64(42),
		ExecutionName: "Test",
	}

	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, nil).Once()
	s.mockPrinter.On("Println", "\nExecution Created", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Name", executionResult.ExecutionName), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionResult.ExecutionId).Return(nil).Once()

	s.mockPrinter.On("Println", "Execution completed.", color.FgGreen).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, args)

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForYAMLInput() {
	t := s.T()

	filename := "/tmp/yaml-input-test"
	testYAML := []byte("SAMPLE_ARG_ONE: any\nSAMPLE_ARG_TWO: variable")
	err := ioutil.WriteFile(filename, testYAML, 0644)
	defer os.Remove(filename)
	assert.NoError(t, err)

	args := []string{"say-hello-world", "-f", filename}
	procArgs := make(map[string]string)
	procArgs["SAMPLE_ARG_ONE"] = "any"
	procArgs["SAMPLE_ARG_TWO"] = "variable"

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With Variables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "SAMPLE_ARG_ONE", "any"), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "SAMPLE_ARG_TWO", "variable"), color.Reset).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId:   uint64(42),
		ExecutionName: "Test",
	}

	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, nil).Once()
	s.mockPrinter.On("Println", "\nExecution Created", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Name", executionResult.ExecutionName), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionResult.ExecutionId).Return(nil).Once()

	s.mockPrinter.On("Println", "Execution completed.", color.FgGreen).Once()

	s.testExecutionCmd.SetArgs(args)
	s.testExecutionCmd.Execute()

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForNoProcVariables() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId: uint64(42),
	}

	procArgs := make(map[string]string)
	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, nil).Once()

	s.mockPrinter.On("Println", "\nExecution Created", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Name", executionResult.ExecutionName), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionResult.ExecutionId).Return(nil).Once()

	s.mockPrinter.On("Println", "Execution completed.", color.FgGreen).Once()

	s.testExecutionCmd.SetArgs(args)
	s.testExecutionCmd.Execute()

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForIncorrectVariableFormat() {
	args := []string{"say-hello-world", "incorrect-format"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With Variables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "\nIncorrect variable format\n", "incorrect-format"), color.FgRed).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId: uint64(42),
	}

	procArgs := make(map[string]string)
	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, nil).Once()

	s.mockPrinter.On("Println", "\nExecution Created", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Name", executionResult.ExecutionName), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionResult.ExecutionId).Return(nil).Once()

	s.mockPrinter.On("Println", "Execution completed.", color.FgGreen).Once()

	s.testExecutionCmd.SetArgs(args)
	s.testExecutionCmd.Execute()

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForProctorDExecutionFailure() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId: uint64(42),
	}
	procArgs := make(map[string]string)
	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, errors.New("test error")).Once()

	s.mockPrinter.On("Println", mock.Anything, color.FgRed).Once()

	osExitFunc := func(exitCode int) {
		assert.Equal(s.T(), 1, exitCode)
	}
	testExecutionCmdOSExit := NewCmd(s.mockPrinter, s.mockProctorDClient, osExitFunc)
	testExecutionCmdOSExit.SetArgs(args)
	testExecutionCmdOSExit.Execute()

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForProctorDLogStreamingFailure() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId: uint64(42),
	}
	procArgs := make(map[string]string)
	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, nil).Once()

	s.mockPrinter.On("Println", "\nExecution Created", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Name", executionResult.ExecutionName), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionResult.ExecutionId).Return(errors.New("error")).Once()

	s.mockPrinter.On("Println", "Error while Streaming Log.", color.FgRed).Once()

	osExitFunc := func(exitCode int) {
		assert.Equal(s.T(), 1, exitCode)
	}
	testExecutionCmdOSExit := NewCmd(s.mockPrinter, s.mockProctorDClient, osExitFunc)
	testExecutionCmdOSExit.SetArgs(args)
	testExecutionCmdOSExit.Execute()

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForProctorDGetDefinitiveProcExecutionStatusError() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId: uint64(42),
	}

	procArgs := make(map[string]string)
	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, nil).Once()

	s.mockPrinter.On("Println", "\nExecution Created", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Name", executionResult.ExecutionName), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionResult.ExecutionId).Return(errors.New("error")).Once()

	s.mockPrinter.On("Println", "Error while Streaming Log.", color.FgRed).Once()

	osExitFunc := func(exitCode int) {
		assert.Equal(s.T(), 1, exitCode)
	}
	testExecutionCmdOSExit := NewCmd(s.mockPrinter, s.mockProctorDClient, osExitFunc)
	testExecutionCmdOSExit.SetArgs(args)
	testExecutionCmdOSExit.Execute()

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdForProctorDGetDefinitiveProcExecutionStatusFailure() {
	args := []string{"say-hello-world"}

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Executing Proc", "say-hello-world"), color.Reset).Once()
	s.mockPrinter.On("Println", "With No Variables", color.FgRed).Once()

	executionResult := &execution.ExecutionResult{
		ExecutionId: uint64(42),
	}
	procArgs := make(map[string]string)
	s.mockProctorDClient.On("ExecuteProc", "say-hello-world", procArgs).Return(executionResult, nil).Once()

	s.mockPrinter.On("Println", "\nExecution Created", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100v", "Name", executionResult.ExecutionName), color.FgGreen).Once()
	s.mockPrinter.On("Println", "\nStreaming logs", color.FgGreen).Once()

	s.mockProctorDClient.On("StreamProcLogs", executionResult.ExecutionId).Return(errors.New("error")).Once()

	s.mockPrinter.On("Println", "Error while Streaming Log.", color.FgRed).Once()

	osExitFunc := func(exitCode int) {
		assert.Equal(s.T(), 1, exitCode)
	}
	testExecutionCmdOSExit := NewCmd(s.mockPrinter, s.mockProctorDClient, osExitFunc)
	testExecutionCmdOSExit.SetArgs(args)
	testExecutionCmdOSExit.Execute()

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestExecutionCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionCmdTestSuite))
}
