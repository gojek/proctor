package main

import (
	"github.com/fatih/color"
	"github.com/gojektech/proctor/cmd"
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
)

func main() {
	printer := io.GetPrinter()
	proctorConfig, err := config.LoadConfig()
	if err != nil {
		printer.Println("Encountered error while loading config, exiting.", color.FgRed)
	}
	proctorEngineClient := daemon.NewClient(proctorConfig)

	cmd.Execute(printer, proctorEngineClient)
}
