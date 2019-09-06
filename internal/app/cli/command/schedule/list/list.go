package list

import (
	"fmt"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"

	"github.com/fatih/color"
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
				printer.Println(fmt.Sprintf("%-40d %-30s %-20s %s", scheduledProc.ID, scheduledProc.Name, scheduledProc.Group, scheduledProc.Tags), color.Reset)
			}
		},
	}
}
