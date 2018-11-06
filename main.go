package main

import (
	"github.com/gojektech/proctor/cmd"
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proctord/logger"
)

func main() {
	printer := io.NewPrinter()
	proctorConfig, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(err)
	}
	proctorEngineClient := daemon.NewClient(proctorConfig)

	cmd.Execute(printer, proctorEngineClient)
}
