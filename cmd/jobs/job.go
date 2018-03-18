package jobs

import (
	"github.com/gojektech/proctor/cmd/jobs/description"
	"github.com/gojektech/proctor/cmd/jobs/execution"
	"github.com/gojektech/proctor/cmd/jobs/list"
	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient engine.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Interact with proctor jobs",
		Long:  `Example: proctor job <command>`,
	}

	descriptionCmd := description.NewCmd(printer, proctorEngineClient)
	cmd.AddCommand(descriptionCmd)

	executionCmd := execution.NewCmd(printer, proctorEngineClient)
	cmd.AddCommand(executionCmd)

	listCmd := list.NewCmd(printer, proctorEngineClient)
	cmd.AddCommand(listCmd)

	return cmd
}
