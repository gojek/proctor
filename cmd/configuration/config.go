package configuration

import (
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

func NewCmd(printer io.Printer, proctorEngineClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "config",
		Short:   "configure proctor with arguments given",
		Long:    "To configure a proctor, this command helps configuring proctor by storing emailId and accessToken locally",
		Example: "proctor config set EMAIL_ID=someone@somewhere.com ACCESS_TOKEN=XXXXXXXXXX",

		Run: func(cmd *cobra.Command, args []string) {
		},
	}
}
