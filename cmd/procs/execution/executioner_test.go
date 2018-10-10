package execution

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExecutionCmdTestSuite struct {
	suite.Suite
	mockPrinter      *io.MockPrinter
	testExecutionCmd *cobra.Command
}

func (s *ExecutionCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.testExecutionCmd = NewCmd(s.mockPrinter)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdUsage() {
	assert.Equal(s.T(), "execute", s.testExecutionCmd.Use)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmdHelp() {
	assert.Equal(s.T(), "[Deprecated][Correct usage: `proctor execute <proc> [args]`]", s.testExecutionCmd.Short)
}

func (s *ExecutionCmdTestSuite) TestExecutionCmd() {
	s.mockPrinter.On("Println", "[Deprecated] Correct usage:\tproctor execute <proc> [args]", color.FgRed).Once()

	s.testExecutionCmd.Run(&cobra.Command{}, []string{})

	s.mockPrinter.AssertExpectations(s.T())
}

func TestExecutionCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionCmdTestSuite))
}
