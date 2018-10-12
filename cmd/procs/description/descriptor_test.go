package description

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DescribeCmdTestSuite struct {
	suite.Suite
	mockPrinter     *io.MockPrinter
	testDescribeCmd *cobra.Command
}

func (s *DescribeCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.testDescribeCmd = NewCmd(s.mockPrinter)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdUsage() {
	assert.Equal(s.T(), "describe", s.testDescribeCmd.Use)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdHelp() {
	assert.Equal(s.T(), "[Deprecated][Correct usage: `proctor describe <proc>`]", s.testDescribeCmd.Short)
}

func (s *DescribeCmdTestSuite) TestDescribeCmdRun() {
	s.mockPrinter.On("Println", "[Deprecated] Correct usage:\tproctor describe <proc>", color.FgRed).Once()

	s.testDescribeCmd.Run(&cobra.Command{}, []string{"do-something"})

	s.mockPrinter.AssertExpectations(s.T())
}

func TestDescribeCmdTestSuite(t *testing.T) {
	suite.Run(t, new(DescribeCmdTestSuite))
}
