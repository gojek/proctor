package status

import (
	"fmt"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/pkg/io"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorDClient daemon.Client, osExitFunc func(int)) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Get status of an execution context",
		Long:    "To get status of an execution context, this command retrieve status from previous execution",
		Example: "proctor status 123",
		Args:    cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			executionIDParam := args[0]
			executionID, err := strconv.ParseUint(executionIDParam, 10, 64)
			if executionIDParam == "" || err != nil {
				printer.Println("No valid execution context id provided as argument", color.FgRed)
				return
			}

			printer.Println("Getting status", color.FgGreen)
			printer.Println(fmt.Sprintf("%-40s %-100v", "ID", executionID), color.FgGreen)

			executionContextStatus, err := proctorDClient.GetExecutionContextStatus(executionID)
			if err != nil {
				printer.Println(fmt.Sprintf("%-40s %-100v", "Error while Getting Status:", err.Error()), color.FgRed)
				osExitFunc(1)
				return
			}

			printer.Println(fmt.Sprintf("%-40s %-100v", "Job Name", executionContextStatus.JobName), color.FgGreen)
			printer.Println(fmt.Sprintf("%-40s %-100v", "Status", executionContextStatus.Status), color.FgGreen)
			printer.Println(fmt.Sprintf("%-40s %-100v", "Updated At", executionContextStatus.UpdatedAt), color.FgGreen)
			printer.Println("Execution completed.", color.FgGreen)
		},
	}
}
