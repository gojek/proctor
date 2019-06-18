package view

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	proctor_config "proctor/config"
	"proctor/io"
)

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:     "show",
		Short:   "Show proctor client config",
		Long:    "This command helps view proctor client config",
		Example: fmt.Sprintf("proctor config show"),

		Run: func(cmd *cobra.Command, args []string) {
			configFile := filepath.Join(proctor_config.ConfigFileDir(), "proctor.yaml")
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				printer.Println(fmt.Sprintf("Client Config is absent: %s", configFile), color.FgRed)
				printer.Println(fmt.Sprintf("Setup config using `proctor config PROCTOR_HOST=some.host ...`"), color.FgRed)
				return
			}

			existingProctorConfig, err := ioutil.ReadFile(configFile)
			if err != nil {
				printer.Println(fmt.Sprintf("Error reading config file: %s", configFile), color.FgRed)
				return
			}

			printer.Println(string(existingProctorConfig))
		},
	}
}
