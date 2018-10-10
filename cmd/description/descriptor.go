package description

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/gojektech/proctor/proc"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "describe",
		Short:   "Describe a proc, list help for variables and constants",
		Long:    "In order to execute a proc, you need to provide certain variables. Describe command helps you with those variables and their meanings/convention/usage, etc.",
		Example: "proctor describe proc-one\nproctor describe proc-two",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				printer.Println("Incorrect command. See `proctor describe --help` for usage", color.FgRed)
				return
			}

			procList, err := proctorEngineClient.ListProcs()
			if err != nil {
				if err.Error() == http.StatusText(http.StatusUnauthorized) {
					printer.Println(utility.UnauthorizedError, color.FgRed)
					return
				}

				printer.Println(utility.GenericDescribeCmdError, color.FgRed)
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
			printer.Println(fmt.Sprintf("%-40s %-100s", "Contributors", desiredProc.Contributors), color.Reset)
			printer.Println(fmt.Sprintf("%-40s %-100s", "Organization", desiredProc.Organization), color.Reset)

			printer.Println("\nVariables", color.FgMagenta)
			for _, arg := range desiredProc.EnvVars.Args {
				printer.Println(fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset)
			}

			printer.Println("\nConstants", color.FgMagenta)
			for _, secret := range desiredProc.EnvVars.Secrets {
				printer.Println(fmt.Sprintf("%-40s %-100s", secret.Name, secret.Description), color.Reset)
			}

			printer.Println("\nFor executing a proc, run:\nproctor execute <proc_name> <args_name>", color.FgGreen)
		},
	}
}
