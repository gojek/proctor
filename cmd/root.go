package cmd

import (
	"fmt"
	"os"

	"github.com/gojektech/proctor/cmd/description"
	"github.com/gojektech/proctor/cmd/execution"
	"github.com/gojektech/proctor/cmd/list"
	"github.com/gojektech/proctor/cmd/procs"
	"github.com/gojektech/proctor/cmd/version"
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "proctor",
		Short: "A command-line interface to run procs",
		Long:  `A command-line interface to interact with proctord, the heart of Proctor: An Automation Orchestrator`,
	}
)

func Execute(printer io.Printer, proctorEngineClient daemon.Client) {
	cobra.OnInitialize(config.InitConfig)

	versionCmd := version.NewCmd(printer)
	rootCmd.AddCommand(versionCmd)

	procCmd := procs.NewCmd(printer)
	rootCmd.AddCommand(procCmd)

	descriptionCmd := description.NewCmd(printer, proctorEngineClient)
	rootCmd.AddCommand(descriptionCmd)

	executionCmd := execution.NewCmd(printer, proctorEngineClient)
	rootCmd.AddCommand(executionCmd)

	listCmd := list.NewCmd(printer, proctorEngineClient)
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
