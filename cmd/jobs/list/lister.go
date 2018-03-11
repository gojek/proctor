package list

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojekfarm/proctor/engine"
	"github.com/gojekfarm/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient engine.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List jobs available with proctor for execution",
		Long:  "Example: proctor job list",
		Run: func(cmd *cobra.Command, args []string) {
			jobList, err := proctorEngineClient.ListJobs()
			if err != nil {
				printer.Println("Error fetching list of jobs. Please check configuration and network connectivity", color.FgRed)
				return
			}

			printer.Println("Proctor Jobs List:\n", color.FgGreen)

			for _, job := range jobList {
				printer.Println(fmt.Sprintf("%-40s %-100s", job.Name, job.Description), color.Reset)
			}

			printer.Println("\nFor detailed information of jobs, run:\nproctor job describe <job_name>", color.FgGreen)
		},
	}
}
