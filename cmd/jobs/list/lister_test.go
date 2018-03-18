package list

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/jobs"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ListCmdTestSuite struct {
	suite.Suite
	mockPrinter             *io.MockPrinter
	mockProctorEngineClient *engine.MockClient
	testListCmd             *cobra.Command
}

func (s *ListCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorEngineClient = &engine.MockClient{}
	s.testListCmd = NewCmd(s.mockPrinter, s.mockProctorEngineClient)
}

func (s *ListCmdTestSuite) TestListCmdUsage() {
	assert.Equal(s.T(), "list", s.testListCmd.Use)
}

func (s *ListCmdTestSuite) TestListCmdHelp() {
	assert.Equal(s.T(), "List jobs available with proctor for execution", s.testListCmd.Short)
	assert.Equal(s.T(), "Example: proctor job list", s.testListCmd.Long)
}

func (s *ListCmdTestSuite) TestListCmdRun() {
	jobOne := jobs.Metadata{
		Name:        "one",
		Description: "job one description",
	}
	jobTwo := jobs.Metadata{
		Name:        "two",
		Description: "job two description",
	}
	jobList := []jobs.Metadata{jobOne, jobTwo}

	s.mockProctorEngineClient.On("ListJobs").Return(jobList, nil).Once()

	s.mockPrinter.On("Println", "Proctor Jobs List:\n", color.FgGreen).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", jobOne.Name, jobOne.Description), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", jobTwo.Name, jobTwo.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nFor detailed information of jobs, run:\nproctor job describe <job_name>", color.FgGreen).Once()

	s.testListCmd.Run(&cobra.Command{}, []string{})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *ListCmdTestSuite) TestListCmdRunProctorEngineClientFailure() {
	s.mockProctorEngineClient.On("ListJobs").Return([]jobs.Metadata{}, errors.New("error")).Once()
	s.mockPrinter.On("Println", "Error fetching list of jobs. Please check configuration and network connectivity", color.FgRed).Once()

	s.testListCmd.Run(&cobra.Command{}, []string{})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestListCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ListCmdTestSuite))
}
