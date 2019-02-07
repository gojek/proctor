package sort

import (
	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"sort"
)

type Sorter interface {
	Sort(procList []metadata.Metadata)
}

var sorterInstance Sorter

type commandSorter struct{}

func GetSorter() Sorter {
	if sorterInstance == nil {
		sorterInstance = &commandSorter{}
	}
	return sorterInstance
}

func (c *commandSorter) Sort(procList []metadata.Metadata) {
	sort.Slice(procList, func(i, j int) bool {
		if procList[i].Name < procList[j].Name {
			return true
		}
		return false
	})
}
