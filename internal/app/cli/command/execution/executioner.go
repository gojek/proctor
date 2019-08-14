package execution

import (
	"fmt"
	"proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorDClient daemon.Client, osExitFunc func(int)) *cobra.Command {
	return &cobra.Command{
		Use:     "execute",
		Short:   "Execute a proc with given arguments",
		Long:    "To execute a proc, this command helps to communicate with `proctord` and streams to logs of proc in execution",
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

			printer.Println("\nExecution Created", color.FgGreen)
			printer.Println(fmt.Sprintf("%-40s %-100v", "ID", executionResult.ExecutionId), color.FgGreen)
			printer.Println(fmt.Sprintf("%-40s %-100s", "Name", executionResult.ExecutionName), color.FgGreen)

			printer.Println("\nStreaming logs", color.FgGreen)
			err = proctorDClient.StreamProcLogs(executionResult.ExecutionId)
			if err != nil {
				printer.Println("Error while Streaming Log.", color.FgRed)
				osExitFunc(1)
				return
			}

			printer.Println("Execution completed.", color.FgGreen)
		},
	}
}
