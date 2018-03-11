package main

import (
	"github.com/gojekfarm/proctor/cmd"
	"github.com/gojekfarm/proctor/engine"
	"github.com/gojekfarm/proctor/io"
)

func main() {
	printer := io.NewPrinter()
	proctorEngineClient := engine.NewClient()

	cmd.Execute(printer, proctorEngineClient)
}
