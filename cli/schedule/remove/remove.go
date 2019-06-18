package remove

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"proctor/daemon"
	"proctor/shared/io"
)

func NewCmd(printer io.Printer, proctorDClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "remove",
		Short:   "Remove scheduled job",
		Long:    "This command helps to remove scheduled job",
		Example: fmt.Sprintf("proctor schedule remove D958FCCC-F2B3-49D1-B83A-4E70A2A775A0"),

		Run: func(cmd *cobra.Command, args []string) {
			jobID := args[0]
			err := proctorDClient.RemoveScheduledProc(jobID)
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
				return
			}
			printer.Println(fmt.Sprintf("Sucessfully removed the scheduled job ID: %s", jobID))
		},
	}
}
