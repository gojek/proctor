package list

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ListCmdTestSuite struct {
	suite.Suite
	mockPrinter *io.MockPrinter
	testListCmd *cobra.Command
}

func (s *ListCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.testListCmd = NewCmd(s.mockPrinter)
}

func (s *ListCmdTestSuite) TestListCmdUsage() {
	assert.Equal(s.T(), "list", s.testListCmd.Use)
}

func (s *ListCmdTestSuite) TestListCmdHelp() {
	assert.Equal(s.T(), "[Deprecated][Correct usage: `proctor list`]", s.testListCmd.Short)
}

func (s *ListCmdTestSuite) TestListCmdRun() {
	s.mockPrinter.On("Println", "[Deprecated] Correct usage: proctor list \n", color.FgRed).Once()

	s.testListCmd.Run(&cobra.Command{}, []string{})

	s.mockPrinter.AssertExpectations(s.T())
}

func TestListCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ListCmdTestSuite))
}
