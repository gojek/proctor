package procs

import (
	"github.com/gojektech/proctor/cmd/procs/description"
	"github.com/gojektech/proctor/cmd/procs/execution"
	"github.com/gojektech/proctor/cmd/procs/list"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient engine.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proc",
		Short: "Interact with proctor procs",
		Long:  `Example: proctor proc <command>`,
	}

	descriptionCmd := description.NewCmd(printer, proctorEngineClient)
	cmd.AddCommand(descriptionCmd)

	executionCmd := execution.NewCmd(printer, proctorEngineClient)
	cmd.AddCommand(executionCmd)

	listCmd := list.NewCmd(printer, proctorEngineClient)
	cmd.AddCommand(listCmd)

	return cmd
}
