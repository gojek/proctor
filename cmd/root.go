package cmd

import (
	"fmt"
	"github.com/gojektech/proctor/cmd/config"
	"github.com/gojektech/proctor/cmd/config/view"
	"github.com/gojektech/proctor/cmd/description"
	"github.com/gojektech/proctor/cmd/execution"
	"github.com/gojektech/proctor/cmd/list"
	"github.com/gojektech/proctor/cmd/schedule"
	"github.com/gojektech/proctor/cmd/version"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "proctor",
		Short: "A command-line interface to run procs",
		Long:  `A command-line interface to run procs`,
	}
)

func Execute(printer io.Printer, proctorDClient daemon.Client) {
	versionCmd := version.NewCmd(printer)
	rootCmd.AddCommand(versionCmd)

	descriptionCmd := description.NewCmd(printer, proctorDClient)
	rootCmd.AddCommand(descriptionCmd)

	//TODO: Test execution.NewCmd is given os.Exit function as params
	executionCmd := execution.NewCmd(printer, proctorDClient, os.Exit)
	rootCmd.AddCommand(executionCmd)

	listCmd := list.NewCmd(printer, proctorDClient)
	rootCmd.AddCommand(listCmd)

	configCmd := config.NewCmd(printer)
	configShowCmd := view.NewCmd(printer)
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)

	scheduleCmd := schedule.NewCmd(printer, proctorDClient)
	rootCmd.AddCommand(scheduleCmd)

	var Time, NotifyEmails, Tags string

	scheduleCmd.PersistentFlags().StringVarP(&Time, "time", "t", "", "Schedule time")
	scheduleCmd.MarkFlagRequired("time")
	scheduleCmd.PersistentFlags().StringVarP(&NotifyEmails, "notify", "n", "", "Notifier Email ID's")
	scheduleCmd.MarkFlagRequired("notify")
	scheduleCmd.PersistentFlags().StringVarP(&Tags, "tags", "T", "", "Tags")
	scheduleCmd.MarkFlagRequired("tags")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
