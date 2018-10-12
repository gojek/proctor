package version

import (
	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version of Proctor command-line tool",
		Long:  `Example: proctor version`,
		Run: func(cmd *cobra.Command, args []string) {
			printer.Println("😊  Proctor: A Developer Friendly Automation Orchestrator v0.2.0", color.Reset)
		},
	}
}
