package sort

import (
	"github.com/stretchr/testify/assert"
	"proctor/shared/model/metadata"
	"testing"
)

func TestSorting(t *testing.T) {
	procOne := metadata.Metadata{Name: "one"}
	procTwo := metadata.Metadata{Name: "two"}
	procThree := metadata.Metadata{Name: "three"}
	procList := []metadata.Metadata{procThree, procTwo, procOne}
	expectedProcList := []metadata.Metadata{procOne, procThree, procTwo}

	Procs(procList)

	assert.Equal(t, expectedProcList, procList)
}
