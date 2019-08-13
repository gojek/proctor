package main

import (
	"proctor/internal/app/cli/command"
	"proctor/internal/app/cli/command/version/github"
	"proctor/internal/app/cli/config"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
)

func main() {
	printer := io.GetPrinter()
	proctorConfigLoader := config.NewLoader()
	proctorDClient := daemon.NewClient(printer, proctorConfigLoader)
	githubClient := github.NewClient()

	command.Execute(printer, proctorDClient, githubClient)
}
