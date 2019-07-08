package describe

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/pkg/io"
)

func NewCmd(printer io.Printer, proctorDClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "describe",
		Short:   "Describe scheduled job",
		Long:    "This command helps to describe scheduled job",
		Example: fmt.Sprintf("proctor schedule describe D958FCCC-F2B3-49D1-B83A-4E70A2A775A0"),

		Run: func(cmd *cobra.Command, args []string) {
			jobID := args[0]
			scheduledProc, err := proctorDClient.DescribeScheduledProc(jobID)
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
				return
			}

			printer.Println(fmt.Sprintf("%-40s %-100s", "ID", scheduledProc.ID), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "PROC NAME", scheduledProc.Name), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "GROUP NAME", scheduledProc.Group), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "TAGS", scheduledProc.Tags), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "Time", scheduledProc.Time), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "Notifier", scheduledProc.NotificationEmails), color.Reset)

			printer.Println("\nArgs", color.FgMagenta)
			for k, v := range scheduledProc.Args {
				printer.Println(fmt.Sprintf("%-40s %-100s", k, v), color.Reset)
			}
		},
	}
}
