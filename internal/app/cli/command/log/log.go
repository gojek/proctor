package log

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
		Use:     "logs",
		Short:   "Get logs of an execution context",
		Long:    "To get a log of execution context, this command helps retrieve logs from previous execution",
		Example: "proctor logs 123",
		Args:    cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			executionIDParam := args[0]
			executionID, err := strconv.ParseUint(executionIDParam, 10, 64)
			if executionIDParam == "" || err != nil {
				printer.Println("No valid execution context id provided as argument", color.FgRed)
				return
			}

			printer.Println("Getting logs", color.FgGreen)
			printer.Println(fmt.Sprintf("%-40s %-100v", "ID", executionID), color.FgGreen)

			printer.Println("\nStreaming logs", color.FgGreen)
			err = proctorDClient.StreamProcLogs(executionID)
			if err != nil {
				printer.Println("Error while Streaming Log.", color.FgRed)
				osExitFunc(1)
				return
			}

			printer.Println("Execution completed.", color.FgGreen)
		},
	}
}
