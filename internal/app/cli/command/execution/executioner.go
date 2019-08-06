package execution

import (
	"fmt"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/pkg/io"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"proctor/internal/pkg/constant"
)

func NewCmd(printer io.Printer, proctorDClient daemon.Client, osExitFunc func(int)) *cobra.Command {
	return &cobra.Command{
		Use:     "execute",
		Short:   "Execute a proc with given arguments",
		Long:    "To execute a proc, this command helps communicate with `proctord` and streams to logs of proc in execution",
		Example: "proctor execute proc-one SOME_VAR=foo ANOTHER_VAR=bar\nproctor execute proc-two ANY_VAR=baz",
		Args:    cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			procName := args[0]
			printer.Println(fmt.Sprintf("%-40s %-100s", "Executing Proc", procName), color.Reset)

			procArgs := make(map[string]string)
			if len(args) > 1 {
				printer.Println("With Variables", color.FgMagenta)
				for _, v := range args[1:] {
					arg := strings.Split(v, "=")

					if len(arg) < 2 {
						printer.Println(fmt.Sprintf("%-40s %-100s", "\nIncorrect variable format\n", v), color.FgRed)
						continue
					}

					combinedArgValue := strings.Join(arg[1:], "=")
					procArgs[arg[0]] = combinedArgValue

					printer.Println(fmt.Sprintf("%-40s %-100s", arg[0], combinedArgValue), color.Reset)
				}
			} else {
				printer.Println("With No Variables", color.FgRed)
			}

			executionResult, err := proctorDClient.ExecuteProc(procName, procArgs)
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
				print()
				osExitFunc(1)
				return
			}

			printer.Println("Proc submitted for execution. \nStreaming logs:", color.FgGreen)
			err = proctorDClient.StreamProcLogs(executionResult.ExecutionId)
			if err != nil {
				printer.Println("Error Streaming Logs", color.FgRed)
				osExitFunc(1)
				return
			}

			printer.Println("Log stream of proc completed.", color.FgGreen)

			procExecutionStatus, err := proctorDClient.GetDefinitiveProcExecutionStatus(executionResult.ExecutionId)
			if err != nil {
				printer.Println("Error Fetching Proc execution status", color.FgRed)
				osExitFunc(1)
				return
			}

			if procExecutionStatus != constant.JobSucceeded {
				printer.Println("Proc execution failed", color.FgRed)
				osExitFunc(1)
				return
			}

			printer.Println("Proc execution successful", color.FgGreen)
		},
	}
}
