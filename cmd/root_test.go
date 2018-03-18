package cmd

import (
	"testing"

	"github.com/gojektech/proctor/engine"
	"github.com/gojektech/proctor/io"
	"github.com/stretchr/testify/assert"
)

func TestRootCmdUsage(t *testing.T) {
	Execute(&io.MockPrinter{}, &engine.MockClient{})

	assert.Equal(t, "proctor", rootCmd.Use)
	assert.Equal(t, "A command-line tool to interact with proctor-engine, the heart of Proctor: An Automation Framework", rootCmd.Long)
}
func TestRootCmdSubCommands(t *testing.T) {
	Execute(&io.MockPrinter{}, &engine.MockClient{})

	subCommands := rootCmd.Commands()

	assert.Equal(t, "help", subCommands[0].Name())
	assert.Equal(t, "job", subCommands[1].Name())
	assert.Equal(t, "job", subCommands[2].Name())
	assert.Equal(t, "version", subCommands[3].Name())
	assert.Equal(t, "version", subCommands[4].Name())
}
