package list

import (
	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "[Deprecated][Correct usage: `proctor list`]",
		Run: func(cmd *cobra.Command, args []string) {
			printer.Println("[Deprecated] Correct usage: proctor list \n", color.FgRed)
		},
	}
}
