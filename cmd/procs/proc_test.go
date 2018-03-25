package procs

import (
	"testing"

	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/stretchr/testify/assert"
)

func TestProcCmdUsage(t *testing.T) {
	procCmd := NewCmd(&io.MockPrinter{}, &daemon.MockClient{})
	assert.Equal(t, "proc", procCmd.Use)
	assert.Equal(t, "Interact with proctor procs", procCmd.Short)
	assert.Equal(t, "Example: proctor proc <command>", procCmd.Long)
}

func TestProcCmdSubCommands(t *testing.T) {
	procCmd := NewCmd(&io.MockPrinter{}, &daemon.MockClient{})

	subCommands := procCmd.Commands()

	assert.Equal(t, "describe", subCommands[0].Name())
	assert.Equal(t, "execute", subCommands[1].Name())
	assert.Equal(t, "list", subCommands[2].Name())
}
