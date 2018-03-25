package main

import (
	"github.com/gojektech/proctor/cmd"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
)

func main() {
	printer := io.NewPrinter()
	proctorEngineClient := daemon.NewClient()

	cmd.Execute(printer, proctorEngineClient)
}
