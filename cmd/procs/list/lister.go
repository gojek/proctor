package list

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient engine.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List procs available with proctor for execution",
		Long:  "Example: proctor proc list",
		Run: func(cmd *cobra.Command, args []string) {
			procList, err := proctorEngineClient.ListProcs()
			if err != nil {
				printer.Println("Error fetching list of procs. Please check configuration and network connectivity", color.FgRed)
				return
			}

			printer.Println("Proctor Procs List:\n", color.FgGreen)

			for _, proc := range procList {
				printer.Println(fmt.Sprintf("%-40s %-100s", proc.Name, proc.Description), color.Reset)
			}

			printer.Println("\nFor detailed information of procs, run:\nproctor proc describe <proc_name>", color.FgGreen)
		},
	}
}
