package description

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/gojekfarm/proctor/engine"
	"github.com/gojekfarm/proctor/io"
	"github.com/gojekfarm/proctor/jobs"
	"github.com/gojekfarm/proctor/jobs/env"
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
	assert.Equal(s.T(), "Describe a job, list help for variables and constants", s.testDescribeCmd.Short)
	assert.Equal(s.T(), "Example: proctor job describe say-hello-world", s.testDescribeCmd.Long)
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

	doSomething := jobs.Metadata{
		Name:        "do-something",
		Description: "does something",
		EnvVars: env.Vars{
			Args:    []env.VarMetadata{arg},
			Secrets: []env.VarMetadata{secret},
		},
	}
	jobList := []jobs.Metadata{doSomething}

	s.mockProctorEngineClient.On("ListJobs").Return(jobList, nil).Once()

	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Job Name", doSomething.Name), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "Job Description", doSomething.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nVariables", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nConstants", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", secret.Name, secret.Description), color.Reset).Once()
	s.mockPrinter.On("Println", "\nFor executing a job, run:\nproctor job execute <job_name> <args_name>", color.FgGreen).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"do-something"})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRunProctorEngineClientFailure() {
	s.mockProctorEngineClient.On("ListJobs").Return([]jobs.Metadata{}, errors.New("error")).Once()
	s.mockPrinter.On("Println", "Error fetching list of jobs. Please check configuration and network connectivity", color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRunJobNotSupported() {
	s.mockProctorEngineClient.On("ListJobs").Return([]jobs.Metadata{}, nil).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("Proctor doesn't support job: %s", "any-job"), color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"any-job"})

	s.mockProctorEngineClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestDescribeCmdTestSuite(t *testing.T) {
	suite.Run(t, new(DescribeCmdTestSuite))
}
