package sort

import "github.com/stretchr/testify/suite"
import (
	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SortTestSuite struct {
	suite.Suite
	sorter *commandSorter
}

func (s *SortTestSuite) setupTest() {
	s.sorter = new(commandSorter)
}

func (s *SortTestSuite) TestSorting() {
	procOne := metadata.Metadata{
		Name:        "one",
		Description: "proc one description",
	}

	procTwo := metadata.Metadata{
		Name:        "two",
		Description: "proc two description",
	}

	procList := []metadata.Metadata{procTwo, procOne}
	expectedProcList := []metadata.Metadata{procOne, procTwo}

	s.sorter.Sort(procList)

	assert.Equal(s.T(), expectedProcList, procList)
}

func TestSortTestSuite(t *testing.T) {
	suite.Run(t, new(SortTestSuite))
}
