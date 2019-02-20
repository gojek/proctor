package version

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/cmd/version/github"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
)

const ClientVersion = "v0.6.0"

func NewCmd(printer io.Printer, fetcher github.LatestReleaseFetcher) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version of Proctor command-line tool",
		Long:  `Example: proctor version`,
		Run: func(cmd *cobra.Command, args []string) {
			printer.Println(fmt.Sprintf("Proctor: A Developer Friendly Automation Orchestrator %s", ClientVersion), color.Reset)
			release, e := fetcher.LatestRelease("gojektech", "proctor")
			if e == nil && *release.TagName != ClientVersion {
				printer.Println(fmt.Sprintf("Your version of Proctor client is out of date! The latest version is %s You can update by either running brew upgrade proctor or downloading a release for your OS here: https://github.com/gojektech/proctor/releases", *release.TagName), color.Reset)
			}
		},
	}
}
