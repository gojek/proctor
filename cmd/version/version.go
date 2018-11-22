package version

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

const ClientVersion = "v0.3.0"

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version of Proctor command-line tool",
		Long:  `Example: proctor version`,
		Run: func(cmd *cobra.Command, args []string) {
			printer.Println(fmt.Sprintf("Proctor: A Developer Friendly Automation Orchestrator %s", ClientVersion), color.Reset)
		},
	}
}
