package sort

import (
	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"sort"
)

func Procs(procList []metadata.Metadata) {
	sort.Slice(procList, func(i, j int) bool {
		if procList[i].Name < procList[j].Name {
			return true
		}
		return false
	})
}
