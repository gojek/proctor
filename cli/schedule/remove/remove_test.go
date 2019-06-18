package remove

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"proctor/daemon"
	"proctor/shared/io"
	"testing"
)

type ScheduleCreateCmdTestSuite struct {
	suite.Suite
	mockPrinter           *io.MockPrinter
	mockProctorDClient    *daemon.MockClient
	testScheduleRemoveCmd *cobra.Command
}

func (s *ScheduleCreateCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon.MockClient{}
	s.testScheduleRemoveCmd = NewCmd(s.mockPrinter, s.mockProctorDClient)
}

func (s *ScheduleCreateCmdTestSuite) TestScheduleCreateCmdHelp() {
	assert.Equal(s.T(), "Remove scheduled job", s.testScheduleRemoveCmd.Short)
	assert.Equal(s.T(), "This command helps to remove scheduled job", s.testScheduleRemoveCmd.Long)
	assert.Equal(s.T(), "proctor schedule remove D958FCCC-F2B3-49D1-B83A-4E70A2A775A0", s.testScheduleRemoveCmd.Example)
}

func TestScheduleCreateCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleCreateCmdTestSuite))
}
