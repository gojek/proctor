package description

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "[Deprecated][Correct usage: `proctor describe <proc>`]",
		Run: func(cmd *cobra.Command, args []string) {
			printer.Println(fmt.Sprintf("[Deprecated] Correct usage:\tproctor describe <proc>"), color.FgRed)
		},
	}
}
