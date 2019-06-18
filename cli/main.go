package main

import (
	"proctor/cli/command"
	"proctor/cli/command/version/github"
	"proctor/cli/config"
	"proctor/cli/daemon"
	"proctor/shared/io"
)

func main() {
	printer := io.GetPrinter()
	proctorConfigLoader := config.NewLoader()
	proctorDClient := daemon.NewClient(printer, proctorConfigLoader)
	githubClient := github.NewClient()

	command.Execute(printer, proctorDClient, githubClient)
}
