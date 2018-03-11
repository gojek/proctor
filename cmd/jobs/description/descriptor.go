package description

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojekfarm/proctor/engine"
	"github.com/gojekfarm/proctor/io"
	"github.com/gojekfarm/proctor/jobs"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient engine.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Describe a job, list help for variables and constants",
		Long:  `Example: proctor job describe say-hello-world`,
		Run: func(cmd *cobra.Command, args []string) {
			jobList, err := proctorEngineClient.ListJobs()
			if err != nil {
				printer.Println("Error fetching list of jobs. Please check configuration and network connectivity", color.FgRed)
				return
			}

			desiredJob := jobs.Metadata{}
			for _, job := range jobList {
				if args[0] == job.Name {
					desiredJob = job
				}
			}
			if len(desiredJob.Name) == 0 {
				printer.Println(fmt.Sprintf("Proctor doesn't support job: %s", args[0]), color.FgRed)
				return
			}

			printer.Println(fmt.Sprintf("%-40s %-100s", "Job Name", desiredJob.Name), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "Job Description", desiredJob.Description), color.Reset)

			printer.Println("\nVariables", color.FgMagenta)
			for _, arg := range desiredJob.EnvVars.Args {
				printer.Println(fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset)
			}

			printer.Println("\nConstants", color.FgMagenta)
			for _, secret := range desiredJob.EnvVars.Secrets {
				printer.Println(fmt.Sprintf("%-40s %-100s", secret.Name, secret.Description), color.Reset)
			}

			printer.Println("\nFor executing a job, run:\nproctor job execute <job_name> <args_name>", color.FgGreen)
		},
	}
}
