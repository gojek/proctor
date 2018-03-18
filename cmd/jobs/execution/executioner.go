package execution

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient engine.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "execute",
		Short: "Execute a job with arguments given",
		Long:  `Example: proctor job execute say-hello-world SAMPLE_ARG_ONE=any SAMPLE_ARG_TWO=variable`,

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				printer.Println("Incorrect usage of proctor job execute", color.FgRed)
				return
			}

			jobName := args[0]
			printer.Println(fmt.Sprintf("%-40s %-100s", "Executing Job", jobName), color.Reset)

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

			executedJobName, err := proctorEngineClient.ExecuteJob(jobName, jobArgs)
			if err != nil {
				printer.Println("\nError executing job. Please check configuration and network connectivity", color.FgRed)
				return
			}

			printer.Println("Job execution successful. \nStreaming logs:", color.FgGreen)
			err = proctorEngineClient.StreamJobLogs(executedJobName)
			if err != nil {
				printer.Println("\nError Streaming Logs", color.FgRed)
				return
			}

			printer.Println("\nLog stream of job completed.", color.FgGreen)
		},
	}
}
