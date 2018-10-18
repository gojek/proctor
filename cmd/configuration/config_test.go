package configuration

import (
	"testing"

	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigCmdTestSuite struct {
	suite.Suite
	mockPrinter             *io.MockPrinter
	mockProctorEngineClient *daemon.MockClient
	testConfigCmd           *cobra.Command
}

func (s *ConfigCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorEngineClient = &daemon.MockClient{}
	s.testConfigCmd = NewCmd(s.mockPrinter, s.mockProctorEngineClient)
}

func (s *ConfigCmdTestSuite) TestConfigCmdUsage() {
	assert.Equal(s.T(), "config", s.testConfigCmd.Use)
}

func (s *ConfigCmdTestSuite) TestConfigCmdHelp() {
	assert.Equal(s.T(), "configure proctor with arguments given", s.testConfigCmd.Short)
	assert.Equal(s.T(), "To configure a proctor, this command helps configuring proctor by storing emailId and accessToken locally", s.testConfigCmd.Long)
	assert.Equal(s.T(), "proctor config set EMAIL_ID=someone@somewhere.com ACCESS_TOKEN=XXXXXXXXXX", s.testConfigCmd.Example)
}

func TestExecutionCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigCmdTestSuite))
}
