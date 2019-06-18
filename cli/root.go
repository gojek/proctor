package cli

import (
	"fmt"
	"os"
	"proctor/cli/schedule/remove"
	"proctor/shared/io"

	"github.com/spf13/cobra"
	"proctor/cli/config"
	"proctor/cli/config/view"
	"proctor/cli/description"
	"proctor/cli/execution"
	"proctor/cli/list"
	"proctor/cli/schedule"
	scheduleDescribe "proctor/cli/schedule/describe"
	scheduleList "proctor/cli/schedule/list"
	"proctor/cli/version"
	"proctor/cli/version/github"
	"proctor/daemon"
)

var (
	rootCmd = &cobra.Command{
		Use:   "proctor",
		Short: "A command-line interface to run procs",
		Long:  `A command-line interface to run procs`,
	}
)

func Execute(printer io.Printer, proctorDClient daemon.Client, githubClient github.LatestReleaseFetcher) {
	versionCmd := version.NewCmd(printer, githubClient)
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
	scheduleListCmd := scheduleList.NewCmd(printer, proctorDClient)
	scheduleCmd.AddCommand(scheduleListCmd)
	scheduleDescribeCmd := scheduleDescribe.NewCmd(printer, proctorDClient)
	scheduleCmd.AddCommand(scheduleDescribeCmd)
	scheduleRemoveCmd := remove.NewCmd(printer, proctorDClient)
	scheduleCmd.AddCommand(scheduleRemoveCmd)

	var Time, NotifyEmails, Tags, Group string

	scheduleCmd.PersistentFlags().StringVarP(&Time, "time", "t", "", "Schedule time")
	_ = scheduleCmd.MarkFlagRequired("time")
	scheduleCmd.PersistentFlags().StringVarP(&Group, "group", "g", "", "Group Name")
	_ = scheduleCmd.MarkFlagRequired("group")
	scheduleCmd.PersistentFlags().StringVarP(&NotifyEmails, "notify", "n", "", "Notifier Email ID's")
	_ = scheduleCmd.MarkFlagRequired("notify")
	scheduleCmd.PersistentFlags().StringVarP(&Tags, "tags", "T", "", "Tags")
	_ = scheduleCmd.MarkFlagRequired("tags")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
