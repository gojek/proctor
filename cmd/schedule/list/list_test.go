package list

import (
	"proctor/daemon"
	"proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"testing"
	"github.com/stretchr/testify/assert"
)

type ScheduleCreateCmdTestSuite struct {
	suite.Suite
	mockPrinter        *io.MockPrinter
	mockProctorDClient *daemon.MockClient
	testScheduleListCmd   *cobra.Command
}

func (s *ScheduleCreateCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testScheduleListCmd = NewCmd(s.mockPrinter, s.mockProctorDClient)
}

func (s *ScheduleCreateCmdTestSuite) TestScheduleCreateCmdHelp() {
	assert.Equal(s.T(), "List scheduled jobs", s.testScheduleListCmd.Short)
	assert.Equal(s.T(), "This command helps to list scheduled jobs", s.testScheduleListCmd.Long)
	assert.Equal(s.T(), "proctor schedule list", s.testScheduleListCmd.Example)
}

func TestScheduleCreateCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleCreateCmdTestSuite))
}
