package jobs

import (
	"testing"

	"github.com/gojekfarm/proctor/engine"
	"github.com/gojekfarm/proctor/io"
	"github.com/stretchr/testify/assert"
)

func TestJobCmdUsage(t *testing.T) {
	jobCmd := NewCmd(&io.MockPrinter{}, &engine.MockClient{})
	assert.Equal(t, "job", jobCmd.Use)
	assert.Equal(t, "Interact with proctor jobs", jobCmd.Short)
	assert.Equal(t, "Example: proctor job <command>", jobCmd.Long)
}

func TestJobCmdSubCommands(t *testing.T) {
	jobCmd := NewCmd(&io.MockPrinter{}, &engine.MockClient{})

	subCommands := jobCmd.Commands()

	assert.Equal(t, "describe", subCommands[0].Name())
	assert.Equal(t, "execute", subCommands[1].Name())
	assert.Equal(t, "list", subCommands[2].Name())
}
