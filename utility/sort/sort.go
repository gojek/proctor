package sort

import (
	"proctor/proctord/jobs/metadata"
	"sort"
)

func Procs(procList []metadata.Metadata) {
	sort.Slice(procList, func(i, j int) bool {
		return procList[i].Name < procList[j].Name
	})
}
