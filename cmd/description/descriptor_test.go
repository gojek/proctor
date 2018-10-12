package description

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proc"
	"github.com/gojektech/proctor/proc/env"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DescribeCmdTestSuite struct {
	suite.Suite
	mockPrinter             *io.MockPrinter
	mockProctorEngineClient *daemon.MockClient
	testDescribeCmd         *cobra.Command
}

func (s *DescribeCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorEngineClient = &daemon.MockClient{}
	s.testDescribeCmd = NewCmd(s.mockPrinter, s.mockProctorEngineClient)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdUsage() {
	assert.Equal(s.T(), "describe", s.testDescribeCmd.Use)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdHelp() {
	assert.Equal(s.T(), "Describe a proc, list help for variables and constants", s.testDescribeCmd.Short)
	assert.Equal(s.T(), "In order to execute a proc, you need to provide certain variables. Describe command helps you with those variables and their meanings/convention/usage, etc.", s.testDescribeCmd.Long)
	assert.Equal(s.T(), "proctor describe proc-one\nproctor describe proc-two", s.testDescribeCmd.Example)
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

	anyProc := proc.Metadata{
		Name:         "any-proc",
		Description:  "does something",
		Contributors: "user@example.com",
		Organization: "org",
		EnvVars: env.Vars{
			Args:    []env.VarMetadata{arg},
			Secrets: []env.VarMetadata{secret},
		},
	}
	procList := []proc.Metadata{anyProc}

	s.mockProctorEngineClient.On("ListProcs").Return(procList, nil).Once()

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Proc Name", anyProc.Name), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Proc Description", anyProc.Description), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Contributors", anyProc.Contributors), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Organization", anyProc.Organization), color.Reset).Once()
	s.mockPrinter.On("Println", "\nVariables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nConstants", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", secret.Name, secret.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nFor executing a proc, run:\nproctor execute <proc_name> <args_name>", color.FgGreen).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"any-proc"})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdForIncorrectUsage() {
	s.mockPrinter.On("Println", "Incorrect command. See `proctor describe --help` for usage", color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{})

	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRunProctorEngineClientFailure() {
	s.mockProctorEngineClient.On("ListProcs").Return([]proc.Metadata{}, errors.New("error")).Once()
	s.mockPrinter.On("Println", utility.GenericDescribeCmdError, color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"any-proc"})

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

func (s *DescribeCmdTestSuite) TestDescribeCmdRunProcForUnauthorizedUser() {
	s.mockProctorEngineClient.On("ListProcs").Return([]proc.Metadata{}, errors.New(http.StatusText(http.StatusUnauthorized))).Once()
	s.mockPrinter.On("Println", utility.UnauthorizedError, color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"any-proc"})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestDescribeCmdTestSuite(t *testing.T) {
	suite.Run(t, new(DescribeCmdTestSuite))
}
