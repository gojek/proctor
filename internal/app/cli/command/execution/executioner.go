package execution

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"proctor/internal/app/cli/daemon"
	utilArgs "proctor/internal/app/cli/utility/args"
	utilFile "proctor/internal/app/cli/utility/file"
	utilIO "proctor/internal/app/cli/utility/io"
)

func NewCmd(printer utilIO.Printer, proctorDClient daemon.Client, osExitFunc func(int)) *cobra.Command {
	executionCmd := &cobra.Command{
		Use:     "execute",
		Short:   "Execute a proc with given arguments",
		Long:    "To execute a proc, this command helps to communicate with `proctord` and streams to logs of proc in execution",
		Example: "proctor execute proc-one SOME_VAR=foo ANOTHER_VAR=bar\nproctor execute proc-two ANY_VAR=baz",
		Args:    cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			procName := args[0]
			printer.Println(fmt.Sprintf("%-40s %-100s", "Executing Proc", procName), color.Reset)

			filename, err := cmd.Flags().GetString("filename")
			if err != nil && !strings.Contains(err.Error(), "flag accessed but not defined") {
				printer.Println(err.Error(), color.FgRed)
			}

			procArgs := make(map[string]string)
			if filename != "" {
				parseErr := utilFile.ParseYAML(filename, procArgs)
				if err != nil {
					printer.Println(parseErr.Error(), color.FgRed)
				}
			}
			if len(procArgs) > 1 || len(args) > 1 {
				printer.Println("With Variables", color.FgMagenta)
				for _, v := range args[1:] {
					utilArgs.ParseArg(printer, procArgs, v)
				}

				for field, value := range procArgs {
					printer.Println(fmt.Sprintf("%-40s %-100s", field, value), color.Reset)
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
	var Filename string

	executionCmd.Flags().StringVarP(&Filename, "filename", "f", "", "Filename")
	executionCmd.MarkFlagFilename("filename")

	return executionCmd
}
