package main

import (
	"proctor/cmd"
	"proctor/config"
	"proctor/daemon"
	"proctor/io"
	"proctor/cmd/version/github"
)

func main() {
	printer := io.GetPrinter()
	proctorConfigLoader := config.NewLoader()
	proctorDClient := daemon.NewClient(printer, proctorConfigLoader)
	githubClient := github.NewClient()

	cmd.Execute(printer, proctorDClient, githubClient)
}
