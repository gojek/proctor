package execution

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "execute",
		Short: "[Deprecated][Correct usage: `proctor execute <proc> [args]`]",

		Run: func(cmd *cobra.Command, args []string) {
			printer.Println(fmt.Sprintf("[Deprecated] Correct usage:\tproctor execute <proc> [args]"), color.FgRed)
		},
	}
}
