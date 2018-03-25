package cmd

import (
	"testing"

	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/stretchr/testify/assert"
)

func TestRootCmdUsage(t *testing.T) {
	Execute(&io.MockPrinter{}, &daemon.MockClient{})

	assert.Equal(t, "proctor", rootCmd.Use)
	assert.Equal(t, "A command-line interface to interact with proctord, the heart of Proctor: An Automation Orchestrator", rootCmd.Long)
}
func TestRootCmdSubCommands(t *testing.T) {
	Execute(&io.MockPrinter{}, &daemon.MockClient{})

	subCommands := rootCmd.Commands()

	assert.Equal(t, "help", subCommands[0].Name())
	assert.Equal(t, "proc", subCommands[1].Name())
	assert.Equal(t, "proc", subCommands[2].Name())
	assert.Equal(t, "version", subCommands[3].Name())
	assert.Equal(t, "version", subCommands[4].Name())
}
