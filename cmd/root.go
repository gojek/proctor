package cmd

import (
	"fmt"
	"os"

	"github.com/gojektech/proctor/cmd/procs"
	"github.com/gojektech/proctor/cmd/version"
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:  "proctor",
		Long: `A command-line interface to interact with proctord, the heart of Proctor: An Automation Orchestrator`,
	}
)

func Execute(printer io.Printer, proctorEngineClient engine.Client) {
	cobra.OnInitialize(config.InitConfig)

	versionCmd := version.NewCmd(printer)
	rootCmd.AddCommand(versionCmd)

	procCmd := procs.NewCmd(printer, proctorEngineClient)
	rootCmd.AddCommand(procCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
