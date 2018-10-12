package procs

import (
	"github.com/gojektech/proctor/cmd/procs/description"
	"github.com/gojektech/proctor/cmd/procs/execution"
	"github.com/gojektech/proctor/cmd/procs/list"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proc",
		Short: "[Deprecated][Correct Usage: `proctor list/describe/execute`]",
	}

	descriptionCmd := description.NewCmd(printer)
	cmd.AddCommand(descriptionCmd)

	executionCmd := execution.NewCmd(printer)
	cmd.AddCommand(executionCmd)

	listCmd := list.NewCmd(printer)
	cmd.AddCommand(listCmd)

	return cmd
}
