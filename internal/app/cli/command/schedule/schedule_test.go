package schedule

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
	mockPrinter           *io.MockPrinter
	mockProctorDClient    *daemon.MockClient
	testScheduleCreateCmd *cobra.Command
}

func (s *ScheduleCreateCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testScheduleCreateCmd = NewCmd(s.mockPrinter, s.mockProctorDClient)
}

func (s *ScheduleCreateCmdTestSuite) TestScheduleCreateCmdHelp() {
	assert.Equal(s.T(), "Create scheduled jobs", s.testScheduleCreateCmd.Short)
	assert.Equal(s.T(), "This command helps to create scheduled jobs", s.testScheduleCreateCmd.Long)
	assert.Equal(s.T(), "proctor schedule run-sample -g my-group -t '0 2 * * *'  -n 'username@mail.com' -T 'sample,proctor' ARG_ONE1=foobar", s.testScheduleCreateCmd.Example)
}

func TestScheduleCreateCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleCreateCmdTestSuite))
}
