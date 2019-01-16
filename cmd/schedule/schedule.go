package schedule

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:     "schedule",
		Short:   "Schedule proctor jobs",
		Long:    "This command helps to maange scheduled proctor jobs",
		Example: fmt.Sprintf("proctor schedule help"),
		Args:    cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			printer.Println(fmt.Sprintf("Print:"), color.FgRed)
		},
	}
}

