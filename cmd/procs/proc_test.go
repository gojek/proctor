package procs

import (
	"testing"

	"github.com/gojektech/proctor/io"
	"github.com/stretchr/testify/assert"
)

func TestProcCmdUsage(t *testing.T) {
	procCmd := NewCmd(&io.MockPrinter{})
	assert.Equal(t, "proc", procCmd.Use)
	assert.Equal(t, "[Deprecated][Correct Usage: `proctor list/describe/execute`]", procCmd.Short)
}

func TestProcCmdSubCommands(t *testing.T) {
	procCmd := NewCmd(&io.MockPrinter{})

	subCommands := procCmd.Commands()

	assert.Equal(t, "describe", subCommands[0].Name())
	assert.Equal(t, "execute", subCommands[1].Name())
	assert.Equal(t, "list", subCommands[2].Name())
}
