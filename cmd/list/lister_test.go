package list

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proc"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ListCmdTestSuite struct {
	suite.Suite
	mockPrinter             *io.MockPrinter
	mockProctorEngineClient *daemon.MockClient
	testListCmd             *cobra.Command
}

func (s *ListCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorEngineClient = &daemon.MockClient{}
	s.testListCmd = NewCmd(s.mockPrinter, s.mockProctorEngineClient)
}

func (s *ListCmdTestSuite) TestListCmdUsage() {
	assert.Equal(s.T(), "list", s.testListCmd.Use)
}

func (s *ListCmdTestSuite) TestListCmdHelp() {
	assert.Equal(s.T(), "List procs available with proctor for execution", s.testListCmd.Short)
	assert.Equal(s.T(), "List procs available with proctor for execution", s.testListCmd.Long)
	assert.Equal(s.T(), "proctor list", s.testListCmd.Example)
}

func (s *ListCmdTestSuite) TestListCmdRun() {
	procOne := proc.Metadata{
		Name:        "one",
		Description: "proc one description",
	}
	procTwo := proc.Metadata{
		Name:        "two",
		Description: "proc two description",
	}
	procList := []proc.Metadata{procOne, procTwo}

	s.mockProctorEngineClient.On("ListProcs").Return(procList, nil).Once()

	s.mockPrinter.On("Println", "List of Procs:\n", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", procOne.Name, procOne.Description), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", procTwo.Name, procTwo.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nFor detailed information of procs, run:\nproctor describe <proc_name>", color.FgGreen).Once()

	s.testListCmd.Run(&cobra.Command{}, []string{})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ListCmdTestSuite) TestListCmdRunProctorEngineClientFailure() {
	s.mockProctorEngineClient.On("ListProcs").Return([]proc.Metadata{}, errors.New("error")).Once()
	s.mockPrinter.On("Println", utility.GenericListCmdError, color.FgRed).Once()

	s.testListCmd.Run(&cobra.Command{}, []string{})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ListCmdTestSuite) TestListCmdRunProctorEngineClientForUnauthorizedUser() {
	s.mockProctorEngineClient.On("ListProcs").Return([]proc.Metadata{}, errors.New(http.StatusText(http.StatusUnauthorized))).Once()
	s.mockPrinter.On("Println", utility.UnauthorizedError, color.FgRed).Once()

	s.testListCmd.Run(&cobra.Command{}, []string{})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestListCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ListCmdTestSuite))
}
