package cli

import (
	"proctor/shared/io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"proctor/cli/version/github"
	"proctor/daemon"
)

func TestRootCmdUsage(t *testing.T) {
	Execute(&io.MockPrinter{}, &daemon.MockClient{}, &github.MockClient{})

	assert.Equal(t, "proctor", rootCmd.Use)
	assert.Equal(t, "A command-line interface to run procs", rootCmd.Short)
	assert.Equal(t, "A command-line interface to run procs", rootCmd.Long)
}

func contains(commands []*cobra.Command, commandName string) bool {
	for _, command := range commands {
		if commandName == command.Name() {
			return true
		}
	}
	return false
}

func TestRootCmdSubCommands(t *testing.T) {
	Execute(&io.MockPrinter{}, &daemon.MockClient{}, &github.MockClient{})

	assert.True(t, contains(rootCmd.Commands(), "describe"))
	assert.True(t, contains(rootCmd.Commands(), "execute"))
	assert.True(t, contains(rootCmd.Commands(), "help"))
	assert.True(t, contains(rootCmd.Commands(), "list"))
	assert.True(t, contains(rootCmd.Commands(), "config"))
	assert.True(t, contains(rootCmd.Commands(), "version"))
	assert.True(t, contains(rootCmd.Commands(), "schedule"))
}
