package version

import (
	"fmt"
	"proctor/internal/app/cli/command/version/github"
	"proctor/internal/pkg/io"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmdUsage(t *testing.T) {
	githubClient := &github.MockClient{}
	versionCmd := NewCmd(&io.MockPrinter{}, githubClient)
	assert.Equal(t, "version", versionCmd.Use)
	assert.Equal(t, "Print version of Proctor command-line tool", versionCmd.Short)
	assert.Equal(t, "Example: proctor version", versionCmd.Long)
}

func TestLatestVersionCmd(t *testing.T) {
	mockPrinter := &io.MockPrinter{}
	githubClient := &github.MockClient{}
	versionCmd := NewCmd(mockPrinter, githubClient)
	version := "v0.6.0"

	mockPrinter.On("Println", fmt.Sprintf("Proctor: A Developer Friendly Automation Orchestrator %s", ClientVersion), color.Reset).Once()
	githubClient.On("LatestRelease", "gojektech", "proctor").Return(version, nil)

	versionCmd.Run(&cobra.Command{}, []string{})

	mockPrinter.AssertExpectations(t)
}

func TestOldVersionCmd(t *testing.T) {
	mockPrinter := &io.MockPrinter{}
	githubClient := &github.MockClient{}
	version := "v1000.0.0"
	versionCmd := NewCmd(mockPrinter, githubClient)

	mockPrinter.On("Println", fmt.Sprintf("Proctor: A Developer Friendly Automation Orchestrator %s", ClientVersion), color.Reset).Once()
	mockPrinter.On("Println", fmt.Sprintf("Your version of Proctor client is out of date!"+
		" The latest version is %s You can update by either running brew upgrade proctor or downloading a release for your OS here:"+
		" https://proctor/releases", version), color.Reset).Once()
	githubClient.On("LatestRelease", "gojektech", "proctor").Return(version, nil)

	versionCmd.Run(&cobra.Command{}, []string{})

	mockPrinter.AssertExpectations(t)
	githubClient.AssertExpectations(t)
}
