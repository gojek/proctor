package args

import (
	"fmt"
	"strings"

	"github.com/fatih/color"

	"proctor/internal/app/cli/utility/io"
)

func ParseArg(printer io.Printer, procArgs map[string]string, arg string) {
	parsedArg := strings.Split(arg, "=")

	if len(parsedArg) < 2 {
		printer.Println(fmt.Sprintf("%-40s %-100s", "\nIncorrect variable format\n", arg), color.FgRed)
		return
	}

	combinedArgValue := strings.Join(parsedArg[1:], "=")
	procArgs[parsedArg[0]] = combinedArgValue
}
