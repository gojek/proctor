package template

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"proctor/internal/app/cli/daemon"
	utilFile "proctor/internal/app/cli/utility/file"
	utilIO "proctor/internal/app/cli/utility/io"
	modelMetadata "proctor/internal/pkg/model/metadata"
)

func NewCmd(printer utilIO.Printer, proctorDClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "template",
		Short:   "Get input template of a procs",
		Long:    "To get input template of a procs, this command retrieve an example template derived from stored metadata",
		Example: "proctor template say-hello-world say-hello-world.yaml",
		Args:    cobra.MinimumNArgs(2),

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				printer.Println("Incorrect command. See `proctor template --help` for usage", color.FgRed)
				return
			}

			userProvidedProcName := args[0]
			filename := args[1]

			procList, err := proctorDClient.ListProcs()
			if err != nil {
				printer.Println(err.Error(), color.FgRed)
				return
			}

			desiredProc := modelMetadata.Metadata{}
			for _, proc := range procList {
				if userProvidedProcName == proc.Name {
					desiredProc = proc
				}
			}
			if len(desiredProc.Name) == 0 {
				printer.Println(fmt.Sprintf("Proctor doesn't support Proc `%s`\nRun `proctor list` to view supported Procs", userProvidedProcName), color.FgRed)
				return
			}

			printer.Println("\nArgs", color.FgMagenta)
			for _, arg := range desiredProc.EnvVars.Args {
				printer.Println(fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset)
			}

			err = utilFile.WriteYAML(filename, desiredProc.EnvVars.Args)
			if err != nil {
				printer.Println(fmt.Sprintf("Error writing template file: %s", err.Error()), color.FgRed)
				return
			}

			printer.Println(fmt.Sprintf("\nTo %s, run:\nproctor execute %s -f %s ARG_ONE=foo ARG_TWO=bar", userProvidedProcName, userProvidedProcName, filename), color.FgGreen)
		},
	}
}
