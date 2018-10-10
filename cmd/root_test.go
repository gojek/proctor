package cmd

import (
	"testing"

	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRootCmdUsage(t *testing.T) {
	Execute(&io.MockPrinter{}, &daemon.MockClient{})

	assert.Equal(t, "proctor", rootCmd.Use)
	assert.Equal(t, "A command-line interface to run procs", rootCmd.Short)
	assert.Equal(t, "A command-line interface to interact with proctord, the heart of Proctor: An Automation Orchestrator", rootCmd.Long)
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
	Execute(&io.MockPrinter{}, &daemon.MockClient{})

	assert.True(t, contains(rootCmd.Commands(), "describe"))
	assert.True(t, contains(rootCmd.Commands(), "execute"))
	assert.True(t, contains(rootCmd.Commands(), "help"))
	assert.True(t, contains(rootCmd.Commands(), "list"))
	assert.True(t, contains(rootCmd.Commands(), "proc"))
	assert.True(t, contains(rootCmd.Commands(), "version"))
}
