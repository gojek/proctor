package main

import (
	"proctor/cli"
	"proctor/cli/version/github"
	"proctor/config"
	"proctor/daemon"
	"proctor/shared/io"
)

func main() {
	printer := io.GetPrinter()
	proctorConfigLoader := config.NewLoader()
	proctorDClient := daemon.NewClient(printer, proctorConfigLoader)
	githubClient := github.NewClient()

	cli.Execute(printer, proctorDClient, githubClient)
}
