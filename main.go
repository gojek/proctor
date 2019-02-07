package main

import (
	"github.com/gojektech/proctor/cmd"
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/utility/sort"
)

func main() {
	printer := io.GetPrinter()
	proctorConfigLoader := config.NewLoader()
	proctorDClient := daemon.NewClient(printer, proctorConfigLoader)
	sorter := sort.GetSorter()
	cmd.Execute(printer, proctorDClient, sorter)
}
