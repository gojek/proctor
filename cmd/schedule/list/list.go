package list

import (
	"fmt"

	"github.com/fatih/color"
	"proctor/daemon"
	"proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorDClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List scheduled jobs",
		Long:    "This command helps to list scheduled jobs",
		Example: fmt.Sprintf("proctor schedule list"),

		Run: func(cmd *cobra.Command, args []string) {
			scheduledProcs, err := proctorDClient.ListScheduledProcs()
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
				return
			}

			printer.Println(fmt.Sprintf("%-40s %-30s %-20s %s", "ID", "PROC NAME", "GROUP NAME", "TAGS"), color.FgGreen)
			for _, scheduledProc := range scheduledProcs {
				printer.Println(fmt.Sprintf("%-40s %-30s %-20s %s", scheduledProc.ID, scheduledProc.Name, scheduledProc.Group,scheduledProc.Tags), color.Reset)
			}
		},
	}
}
