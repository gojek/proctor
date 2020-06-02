package schedule

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
	"strings"
)

func NewCmd(printer io.Printer, proctorDClient daemon.Client) *cobra.Command {
	scheduleCmd := &cobra.Command{
		Use:     "schedule",
		Short:   "Create scheduled jobs",
		Long:    "This command helps to create scheduled jobs",
		Example: fmt.Sprintf("proctor schedule run-sample -g my-group -c '0 2 * * *'  -n 'username@mail.com' -T 'sample,proctor' ARG_ONE1=foobar"),
		Args:    cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			procName := args[0]
			printer.Println(fmt.Sprintf("%-40s %-100s", "Creating Scheduled Job", procName), color.Reset)
			cron, err := cmd.Flags().GetString("cron")
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
			}

			notificationEmails, err := cmd.Flags().GetString("notify")
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
			}

			tags, err := cmd.Flags().GetString("tags")
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
			}

			group, err := cmd.Flags().GetString("group")
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
			}

			jobArgs := make(map[string]string)
			if len(args) > 1 {
				printer.Println("With Variables", color.FgMagenta)
				for _, v := range args[1:] {
					arg := strings.Split(v, "=")

					if len(arg) < 2 {
						printer.Println(fmt.Sprintf("%-40s %-100s", "\nIncorrect variable format\n", v), color.FgRed)
						continue
					}

					combinedArgValue := strings.Join(arg[1:], "=")
					jobArgs[arg[0]] = combinedArgValue

					printer.Println(fmt.Sprintf("%-40s %-100s", arg[0], combinedArgValue), color.Reset)
				}
			} else {
				printer.Println("With No Variables", color.FgRed)
			}

			scheduledJobID, err := proctorDClient.ScheduleJob(procName, tags, cron, notificationEmails, group, jobArgs)
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
				print()
				return
			}
			printer.Println(fmt.Sprintf("Scheduled Job UUID : %d", scheduledJobID), color.FgGreen)
		},
	}

	var Cron, NotifyEmails, Tags, Group string

	scheduleCmd.PersistentFlags().StringVarP(&Cron, "cron", "c", "", "Schedule cron")
	_ = scheduleCmd.MarkFlagRequired("cron")
	scheduleCmd.PersistentFlags().StringVarP(&Group, "group", "g", "", "Group Name")
	_ = scheduleCmd.MarkFlagRequired("group")
	scheduleCmd.PersistentFlags().StringVarP(&NotifyEmails, "notify", "n", "", "Notifier Email ID's")
	_ = scheduleCmd.MarkFlagRequired("notify")
	scheduleCmd.PersistentFlags().StringVarP(&Tags, "tags", "T", "", "Tags")
	_ = scheduleCmd.MarkFlagRequired("tags")

	return scheduleCmd
}
