package description

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proc"
	"github.com/gojektech/proctor/proc/env"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DescribeCmdTestSuite struct {
	suite.Suite
	mockPrinter        *io.MockPrinter
	mockProctorDClient *daemon.MockClient
	testDescribeCmd    *cobra.Command
}

func (s *DescribeCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testDescribeCmd = NewCmd(s.mockPrinter, s.mockProctorDClient)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdUsage() {
	assert.Equal(s.T(), "describe", s.testDescribeCmd.Use)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdHelp() {
	assert.Equal(s.T(), "Help on executing a proc", s.testDescribeCmd.Short)
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
		Name:             "do-something",
		Description:      "does something",
		Contributors:     "user@example.com",
		Organization:     "org",
		AuthorizedGroups: []string{"group_one", "group_two"},
		EnvVars: env.Vars{
			Args:    []env.VarMetadata{arg},
			Secrets: []env.VarMetadata{secret},
		},
	}
	procList := []proc.Metadata{anyProc}

	s.mockProctorDClient.On("ListProcs").Return(procList, nil).Once()

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Description", anyProc.Description), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Contributors", anyProc.Contributors), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Organization", anyProc.Organization), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s [%s]", "Authorized Groups", strings.Join(anyProc.AuthorizedGroups, ", ")), color.Reset).Once()
	s.mockPrinter.On("Println", "\nArgs", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("\nTo %s, run:\nproctor execute %s ARG_ONE=foo ARG_TWO=bar", anyProc.Name, anyProc.Name), color.FgGreen).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{anyProc.Name})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdForIncorrectUsage() {
	s.mockPrinter.On("Println", "Incorrect command. See `proctor describe --help` for usage", color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{})

	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRunProctorDClientFailure() {
	s.mockProctorDClient.On("ListProcs").Return([]proc.Metadata{}, errors.New("test error")).Once()
	s.mockPrinter.On("Println", "test error", color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"do-something"})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRunProcNotSupported() {
	s.mockProctorDClient.On("ListProcs").Return([]proc.Metadata{}, nil).Once()
	testProcName := "do-something"
	s.mockPrinter.On("Println", fmt.Sprintf("Proctor doesn't support Proc `%s`\nRun `proctor list` to view supported Procs", testProcName), color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{testProcName})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestDescribeCmdTestSuite(t *testing.T) {
	suite.Run(t, new(DescribeCmdTestSuite))
}
