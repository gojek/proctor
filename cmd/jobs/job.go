package jobs

import (
	"github.com/gojekfarm/proctor/cmd/jobs/description"
	"github.com/gojekfarm/proctor/cmd/jobs/execution"
	"github.com/gojekfarm/proctor/cmd/jobs/list"
	"github.com/gojekfarm/proctor/engine"
	"github.com/gojekfarm/proctor/io"
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
