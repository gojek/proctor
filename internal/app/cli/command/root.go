package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"proctor/internal/app/cli/command/config"
	"proctor/internal/app/cli/command/config/view"
	"proctor/internal/app/cli/command/description"
	"proctor/internal/app/cli/command/execution"
	"proctor/internal/app/cli/command/list"
	"proctor/internal/app/cli/command/log"
	"proctor/internal/app/cli/command/schedule"
	scheduleDescribe "proctor/internal/app/cli/command/schedule/describe"
	scheduleList "proctor/internal/app/cli/command/schedule/list"
	"proctor/internal/app/cli/command/schedule/remove"
	"proctor/internal/app/cli/command/status"
	"proctor/internal/app/cli/command/version"
	"proctor/internal/app/cli/command/version/github"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
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

	logCmd := log.NewCmd(printer, proctorDClient, os.Exit)
	rootCmd.AddCommand(logCmd)

	statusCmd := status.NewCmd(printer, proctorDClient, os.Exit)
	rootCmd.AddCommand(statusCmd)

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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
