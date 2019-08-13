package version

import (
	"fmt"
	"proctor/internal/app/cli/command/version/github"
	"proctor/internal/app/cli/utility/io"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const ClientVersion = "v2.0.0"

func NewCmd(printer io.Printer, fetcher github.LatestReleaseFetcher) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version of Proctor command-line tool",
		Long:  `Example: proctor version`,
		Run: func(cmd *cobra.Command, args []string) {
			printer.Println(fmt.Sprintf("Proctor: A Developer Friendly Automation Orchestrator %s", ClientVersion), color.Reset)
			release, e := fetcher.LatestRelease("gojektech", "proctor")
			if e == nil && release != ClientVersion {
				printer.Println(fmt.Sprintf("Your version of Proctor client is out of date! The latest version is %s You can update by either running brew upgrade proctor or downloading a release for your OS here: https://proctor/releases", release), color.Reset)
			}
		},
	}
}
