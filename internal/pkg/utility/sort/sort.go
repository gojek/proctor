package sort

import (
	"proctor/internal/pkg/model/metadata"
	"sort"
)

func Procs(procList []metadata.Metadata) {
	sort.Slice(procList, func(i, j int) bool {
		return procList[i].Name < procList[j].Name
	})
}
