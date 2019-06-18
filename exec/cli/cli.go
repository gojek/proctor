package main

import (
	"proctor/cmd"
	"proctor/cmd/version/github"
	"proctor/config"
	"proctor/daemon"
	"proctor/io"
)

func main() {
	printer := io.GetPrinter()
	proctorConfigLoader := config.NewLoader()
	proctorDClient := daemon.NewClient(printer, proctorConfigLoader)
	githubClient := github.NewClient()

	cmd.Execute(printer, proctorDClient, githubClient)
}
