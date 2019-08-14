package args

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"

	"proctor/internal/app/cli/utility/io"
)

func TestParseArg(t *testing.T) {
	procArgs := make(map[string]string)
	mockPrinter := &io.MockPrinter{}
	ParseArg(mockPrinter, procArgs, "foo=moo")
	assert.Equal(t, procArgs["foo"], "moo")
}

func TestParseArgError(t *testing.T) {
	procArgs := make(map[string]string)
	mockPrinter := &io.MockPrinter{}

	mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", "\nIncorrect variable format\n", "foo"), color.FgRed)
	defer mockPrinter.AssertExpectations(t)

	ParseArg(mockPrinter, procArgs, "foo")
}
