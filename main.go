package main

import (
	"github.com/gojektech/proctor/cmd"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
)

func main() {
	printer := io.NewPrinter()
	proctorEngineClient := engine.NewClient()

	cmd.Execute(printer, proctorEngineClient)
}
