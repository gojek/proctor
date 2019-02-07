package cmd

import (
	"fmt"
	"github.com/gojektech/proctor/cmd/schedule/remove"
	"os"

	"github.com/gojektech/proctor/cmd/config"
	"github.com/gojektech/proctor/cmd/config/view"
	"github.com/gojektech/proctor/cmd/description"
	"github.com/gojektech/proctor/cmd/execution"
	"github.com/gojektech/proctor/cmd/list"
	"github.com/gojektech/proctor/cmd/schedule"
	schedule_list "github.com/gojektech/proctor/cmd/schedule/list"
	schedule_describe "github.com/gojektech/proctor/cmd/schedule/describe"
	"github.com/gojektech/proctor/cmd/version"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"

	"github.com/spf13/cobra"
	"github.com/gojektech/proctor/utility/sort"
)

var (
	rootCmd = &cobra.Command{
		Use:   "proctor",
		Short: "A command-line interface to run procs",
		Long:  `A command-line interface to run procs`,
	}
)

func Execute(printer io.Printer, proctorDClient daemon.Client, sorter sort.Sorter) {
	versionCmd := version.NewCmd(printer)
	rootCmd.AddCommand(versionCmd)

	descriptionCmd := description.NewCmd(printer, proctorDClient)
	rootCmd.AddCommand(descriptionCmd)

	//TODO: Test execution.NewCmd is given os.Exit function as params
	executionCmd := execution.NewCmd(printer, proctorDClient, os.Exit)
	rootCmd.AddCommand(executionCmd)

	listCmd := list.NewCmd(printer, proctorDClient, sorter)
	rootCmd.AddCommand(listCmd)

	configCmd := config.NewCmd(printer)
	configShowCmd := view.NewCmd(printer)
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)

	scheduleCmd := schedule.NewCmd(printer, proctorDClient)
	rootCmd.AddCommand(scheduleCmd)
	scheduleListCmd := schedule_list.NewCmd(printer, proctorDClient)
	scheduleCmd.AddCommand(scheduleListCmd)
	scheduleDescribeCmd := schedule_describe.NewCmd(printer, proctorDClient)
	scheduleCmd.AddCommand(scheduleDescribeCmd)
	scheduleRemoveCmd := remove.NewCmd(printer, proctorDClient)
	scheduleCmd.AddCommand(scheduleRemoveCmd)

	var Time, NotifyEmails, Tags, Group string

	scheduleCmd.PersistentFlags().StringVarP(&Time, "time", "t", "", "Schedule time")
	scheduleCmd.MarkFlagRequired("time")
	scheduleCmd.PersistentFlags().StringVarP(&Group, "group", "g", "", "Group Name")
	scheduleCmd.MarkFlagRequired("group")
	scheduleCmd.PersistentFlags().StringVarP(&NotifyEmails, "notify", "n", "", "Notifier Email ID's")
	scheduleCmd.MarkFlagRequired("notify")
	scheduleCmd.PersistentFlags().StringVarP(&Tags, "tags", "T", "", "Tags")
	scheduleCmd.MarkFlagRequired("tags")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
