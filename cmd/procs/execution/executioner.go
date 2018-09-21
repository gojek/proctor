package execution

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "execute",
		Short: "Execute a proc with arguments given",
		Long:  `Example: proctor proc execute say-hello-world SAMPLE_ARG_ONE=any SAMPLE_ARG_TWO=variable`,

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				printer.Println("Incorrect usage of proctor proc execute", color.FgRed)
				return
			}

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

			executedProcName, err := proctorEngineClient.ExecuteProc(procName, procArgs)
			if err != nil {
				if err.Error() == http.StatusText(http.StatusUnauthorized) {
					printer.Println(utility.UnauthorizedError, color.FgRed)
					return
				}
				printer.Println(utility.GenericProcCmdError, color.FgRed)
				return
			}

			printer.Println("Proc execution successful. \nStreaming logs:", color.FgGreen)
			err = proctorEngineClient.StreamProcLogs(executedProcName)
			if err != nil {
				printer.Println("Error Streaming Logs", color.FgRed)
				return
			}

			printer.Println("Log stream of proc completed.", color.FgGreen)
		},
	}
}
