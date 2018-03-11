package version

import (
	"github.com/fatih/color"
	"github.com/gojekfarm/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version of Proctor command-line tool",
		Long:  `Example: proctor version`,
		Run: func(cmd *cobra.Command, args []string) {
			printer.Println("😊  Proctor: An Automation Framework v0.1.0", color.Reset)
		},
	}
}
