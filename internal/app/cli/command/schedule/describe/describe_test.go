package describe

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
	"testing"
)

type ScheduleCreateCmdTestSuite struct {
	suite.Suite
	mockPrinter             *io.MockPrinter
	mockProctorDClient      *daemon.MockClient
	testScheduleDescribeCmd *cobra.Command
}

func (s *ScheduleCreateCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testScheduleDescribeCmd = NewCmd(s.mockPrinter, s.mockProctorDClient)
}

func (s *ScheduleCreateCmdTestSuite) TestScheduleCreateCmdHelp() {
	assert.Equal(s.T(), "Describe scheduled job", s.testScheduleDescribeCmd.Short)
	assert.Equal(s.T(), "This command helps to describe scheduled job", s.testScheduleDescribeCmd.Long)
	assert.Equal(s.T(), "proctor schedule describe 502376124721", s.testScheduleDescribeCmd.Example)
}

func TestScheduleCreateCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleCreateCmdTestSuite))
}
