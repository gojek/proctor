package description

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proc"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient engine.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Describe a proc, list help for variables and constants",
		Long:  `Example: proctor proc describe say-hello-world`,
		Run: func(cmd *cobra.Command, args []string) {
			procList, err := proctorEngineClient.ListProcs()
			if err != nil {
				printer.Println("Error fetching list of procs. Please check configuration and network connectivity", color.FgRed)
				return
			}

			desiredProc := proc.Metadata{}
			for _, proc := range procList {
				if args[0] == proc.Name {
					desiredProc = proc
				}
			}
			if len(desiredProc.Name) == 0 {
				printer.Println(fmt.Sprintf("Proctor doesn't support proc: %s", args[0]), color.FgRed)
				return
			}

			printer.Println(fmt.Sprintf("%-40s %-100s", "Proc Name", desiredProc.Name), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "Proc Description", desiredProc.Description), color.Reset)

			printer.Println("\nVariables", color.FgMagenta)
			for _, arg := range desiredProc.EnvVars.Args {
				printer.Println(fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset)
			}

			printer.Println("\nConstants", color.FgMagenta)
			for _, secret := range desiredProc.EnvVars.Secrets {
				printer.Println(fmt.Sprintf("%-40s %-100s", secret.Name, secret.Description), color.Reset)
			}

			printer.Println("\nFor executing a proc, run:\nproctor proc execute <proc_name> <args_name>", color.FgGreen)
		},
	}
}
