package version

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gojekfarm/proctor/io"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmdUsage(t *testing.T) {
	versionCmd := NewCmd(&io.MockPrinter{})
	assert.Equal(t, "version", versionCmd.Use)
	assert.Equal(t, "Print version of Proctor command-line tool", versionCmd.Short)
	assert.Equal(t, "Example: proctor version", versionCmd.Long)
}

func TestVersionCmd(t *testing.T) {
	mockPrinter := &io.MockPrinter{}
	versionCmd := NewCmd(mockPrinter)

	mockPrinter.On("Println", "😊  Proctor: An Automation Framework v0.1.0", color.Reset).Once()

	versionCmd.Run(&cobra.Command{}, []string{})

	mockPrinter.AssertExpectations(t)
}
