package description

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proc"
	"github.com/gojektech/proctor/proc/env"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DescribeCmdTestSuite struct {
	suite.Suite
	mockPrinter             *io.MockPrinter
	mockProctorEngineClient *engine.MockClient
	testDescribeCmd         *cobra.Command
}

func (s *DescribeCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorEngineClient = &engine.MockClient{}
	s.testDescribeCmd = NewCmd(s.mockPrinter, s.mockProctorEngineClient)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdUsage() {
	assert.Equal(s.T(), "describe", s.testDescribeCmd.Use)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdHelp() {
	assert.Equal(s.T(), "Describe a proc, list help for variables and constants", s.testDescribeCmd.Short)
	assert.Equal(s.T(), "Example: proctor proc describe say-hello-world", s.testDescribeCmd.Long)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRun() {
	arg := env.VarMetadata{
		Name:        "arg-one",
		Description: "arg one description",
	}

	secret := env.VarMetadata{
		Name:        "secret-one",
		Description: "secret one description",
	}

	doSomething := proc.Metadata{
		Name:        "do-something",
		Description: "does something",
		EnvVars: env.Vars{
			Args:    []env.VarMetadata{arg},
			Secrets: []env.VarMetadata{secret},
		},
	}
	procList := []proc.Metadata{doSomething}

	s.mockProctorEngineClient.On("ListProcs").Return(procList, nil).Once()

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Proc Name", doSomething.Name), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Proc Description", doSomething.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nVariables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nConstants", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", secret.Name, secret.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nFor executing a proc, run:\nproctor proc execute <proc_name> <args_name>", color.FgGreen).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"do-something"})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRunProctorEngineClientFailure() {
	s.mockProctorEngineClient.On("ListProcs").Return([]proc.Metadata{}, errors.New("error")).Once()
	s.mockPrinter.On("Println", "Error fetching list of procs. Please check configuration and network connectivity", color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRunProcNotSupported() {
	s.mockProctorEngineClient.On("ListProcs").Return([]proc.Metadata{}, nil).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("Proctor doesn't support proc: %s", "any-proc"), color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"any-proc"})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestDescribeCmdTestSuite(t *testing.T) {
	suite.Run(t, new(DescribeCmdTestSuite))
}
